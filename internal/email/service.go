package email

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/smtp"
)

// Service handles email sending
type Service struct {
	smtpHost     string
	smtpPort     string
	smtpUsername string
	smtpPassword string
	fromEmail    string
	fromName     string
}

// NewService creates a new email service
func NewService(smtpHost, smtpPort, smtpUsername, smtpPassword, fromEmail, fromName string) *Service {
	return &Service{
		smtpHost:     smtpHost,
		smtpPort:     smtpPort,
		smtpUsername: smtpUsername,
		smtpPassword: smtpPassword,
		fromEmail:    fromEmail,
		fromName:     fromName,
	}
}

// SendVerificationEmail sends an email verification email to the user
func (s *Service) SendVerificationEmail(toEmail, verificationToken, baseURL string) error {
	verificationURL := fmt.Sprintf("%s/api/auth/verify-email?token=%s", baseURL, verificationToken)
	
	subject := "Verify Your Email Address"
	body := fmt.Sprintf(`Hello,

Thank you for registering! Please verify your email address by clicking the link below:

%s

This link will expire in 24 hours.

If you did not create an account, please ignore this email.

Best regards,
%s`, verificationURL, s.fromName)

	return s.sendEmail(toEmail, subject, body)
}

// sendEmail sends an email using SMTP
func (s *Service) sendEmail(toEmail, subject, body string) error {
	// If SMTP is not configured, log and skip sending (for development)
	if s.smtpHost == "" || s.smtpPort == "" {
		fmt.Printf("[EMAIL] Would send email to %s\n", toEmail)
		fmt.Printf("[EMAIL] Subject: %s\n", subject)
		fmt.Printf("[EMAIL] Body: %s\n", body)
		return nil
	}

	// Set up authentication
	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)

	// Create email message
	from := s.fromEmail
	if s.fromName != "" {
		from = fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail)
	}

	msg := []byte(fmt.Sprintf("From: %s\r\n", from) +
		fmt.Sprintf("To: %s\r\n", toEmail) +
		fmt.Sprintf("Subject: %s\r\n", subject) +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"\r\n" +
		body + "\r\n")

	// Send email
	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)
	err := smtp.SendMail(addr, auth, s.fromEmail, []string{toEmail}, msg)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// GenerateVerificationToken generates a secure random token for email verification
func GenerateVerificationToken() (string, error) {
	b := make([]byte, 32) // 64 character hex string
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

