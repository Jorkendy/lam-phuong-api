package email

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// Service handles email sending via Gmail API
type Service struct {
	gmailService *gmail.Service
	fromEmail    string
	fromName     string
}

// NewService creates a new email service using Gmail API
func NewService(clientID, clientSecret, refreshToken, fromEmail, fromName string) (*Service, error) {
	ctx := context.Background()

	// Create OAuth config from environment variables
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{gmail.GmailSendScope},
		RedirectURL:  "http://localhost:8080/oauth2callback",
	}

	// Create token from refresh token
	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	// Create HTTP client with token
	client := config.Client(ctx, token)

	// Create Gmail service
	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Gmail client: %w", err)
	}

	return &Service{
		gmailService: srv,
		fromEmail:    fromEmail,
		fromName:     fromName,
	}, nil
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
