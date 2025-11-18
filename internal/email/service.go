package email

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// Service handles email sending via Gmail API
type Service struct {
	gmailService    *gmail.Service
	fromEmail       string
	fromName        string
	credentialsPath string
	tokenPath       string
}

// NewService creates a new email service using Gmail API
func NewService(credentialsPath, tokenPath, fromEmail, fromName string) (*Service, error) {
	ctx := context.Background()

	// Read credentials file
	b, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %w", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.GmailSendScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %w", err)
	}

	// Redirect URI will be set dynamically based on available port
	// OOB flow is deprecated, so we use a local HTTP server

	client := getClient(config, tokenPath)
	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Gmail client: %w", err)
	}

	return &Service{
		gmailService:    srv,
		fromEmail:       fromEmail,
		fromName:        fromName,
		credentialsPath: credentialsPath,
		tokenPath:       tokenPath,
	}, nil
}

// getClient retrieves a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config, tokenPath string) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first time.
	tok, err := tokenFromFile(tokenPath)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokenPath, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	// Start a local server to receive the OAuth callback
	codeCh := make(chan string)
	errCh := make(chan error)

	// Try to find an available port (start from 8082 to avoid conflicts with main server)
	var listener net.Listener
	var err error
	var port string
	ports := []string{"8082", "8083", "8084", "8085"}

	for _, p := range ports {
		listener, err = net.Listen("tcp", ":"+p)
		if err == nil {
			port = p
			break
		}
	}

	if listener == nil {
		log.Fatalf("Unable to start local server on any available port")
	}
	defer listener.Close()

	redirectURL := fmt.Sprintf("http://localhost:%s/oauth2callback", port)
	config.RedirectURL = redirectURL

	// Create a new mux for this server to avoid conflicts
	mux := http.NewServeMux()

	mux.HandleFunc("/oauth2callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			errCh <- fmt.Errorf("no authorization code received")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Authorization failed. No code received."))
			return
		}
		codeCh <- code
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Authorization successful! You can close this window."))
	})

	// Start HTTP server
	server := &http.Server{
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// Generate authorization URL
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser:\n%v\n\n", authURL)
	fmt.Printf("Waiting for authorization on http://localhost:%s/oauth2callback...\n", port)

	// Wait for authorization code or error
	var authCode string
	select {
	case authCode = <-codeCh:
		// Authorization code received, shutdown server
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	case err := <-errCh:
		server.Shutdown(context.Background())
		log.Fatalf("Error during authorization: %v", err)
	case <-time.After(5 * time.Minute):
		server.Shutdown(context.Background())
		log.Fatalf("Authorization timeout. Please try again.")
	}

	// Exchange authorization code for token
	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// SendTestEmail sends a test email to the specified address
func (s *Service) SendTestEmail(toEmail string) error {
	subject := "Test Email from Lam Phuong API"
	body := fmt.Sprintf(`Hello,

This is a test email from Lam Phuong API.

If you received this email, the email service is working correctly.

Best regards,
%s`, s.fromName)

	return s.sendEmail(toEmail, subject, body)
}

// sendEmail sends an email using Gmail API
func (s *Service) sendEmail(toEmail, subject, body string) error {
	// Validate email addresses
	if !isValidEmail(toEmail) {
		return fmt.Errorf("invalid recipient email address: %s", toEmail)
	}

	// Create email message
	from := s.fromEmail
	if s.fromName != "" {
		from = fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail)
	}

	// Build email message with proper headers
	message := fmt.Sprintf("From: %s\r\n", from)
	message += fmt.Sprintf("To: %s\r\n", toEmail)
	message += fmt.Sprintf("Subject: %s\r\n", subject)
	message += "MIME-Version: 1.0\r\n"
	message += "Content-Type: text/plain; charset=UTF-8\r\n"
	message += "\r\n"
	message += body

	// Encode message as base64url
	rawMessage := base64.URLEncoding.EncodeToString([]byte(message))

	// Create Gmail message
	msg := &gmail.Message{
		Raw: rawMessage,
	}

	// Send message
	_, err := s.gmailService.Users.Messages.Send("me", msg).Do()
	if err != nil {
		return fmt.Errorf("failed to send email via Gmail API: %w", err)
	}

	return nil
}

// isValidEmail performs basic email validation
func isValidEmail(email string) bool {
	if email == "" {
		return false
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	if parts[0] == "" || parts[1] == "" {
		return false
	}
	return true
}
