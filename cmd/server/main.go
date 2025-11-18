package main

import (
	"log"
	"os"
	"strings"
	"time"

	docs "lam-phuong-api/docs" // Import docs for Swagger
	"lam-phuong-api/internal/config"
	"lam-phuong-api/internal/email"
	"lam-phuong-api/internal/location"
	"lam-phuong-api/internal/server"
	"lam-phuong-api/internal/user"
)

// @title           Lam Phuong API
// @version         1.0
// @description     API for managing locations
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api

// @schemes   http https

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Type "Bearer" followed by a space and JWT token.
func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Validate required configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Configure Swagger host/schemes so deployed instances don't default to localhost
	swaggerHost := strings.TrimSpace(os.Getenv("SWAGGER_HOST"))
	if swaggerHost == "" {
		swaggerHost = cfg.ServerAddress()
	}
	docs.SwaggerInfo.Host = swaggerHost

	if schemesEnv := strings.TrimSpace(os.Getenv("SWAGGER_SCHEMES")); schemesEnv != "" {
		parts := strings.Split(schemesEnv, ",")
		docs.SwaggerInfo.Schemes = nil
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" {
				docs.SwaggerInfo.Schemes = append(docs.SwaggerInfo.Schemes, part)
			}
		}
	} else {
		// Default to https for non-local hosts
		if strings.Contains(swaggerHost, "localhost") || strings.HasPrefix(swaggerHost, "127.") {
			docs.SwaggerInfo.Schemes = []string{"http"}
		} else {
			docs.SwaggerInfo.Schemes = []string{"https", "http"}
		}
	}

	// Create Airtable client
	airtableClient, err := cfg.NewAirtableClient()
	if err != nil {
		log.Fatalf("Failed to create Airtable client: %v", err)
	}

	// Create Airtable repositories directly
	locationRepo := location.NewAirtableRepository(airtableClient, cfg.Airtable.LocationsTableName)
	locationHandler := location.NewHandler(locationRepo)

	userRepo := user.NewAirtableRepository(airtableClient, cfg.Airtable.UsersTableName)

	// Create user handler with JWT configuration
	tokenExpiry := time.Duration(cfg.Auth.TokenExpiry) * time.Hour
	userHandler := user.NewHandler(userRepo, cfg.Auth.JWTSecret, tokenExpiry)

	// Initialize email service (Gmail API)
	var emailService *email.Service
	var emailHandler *email.Handler
	if cfg.Email.ClientID != "" && cfg.Email.ClientSecret != "" && cfg.Email.RefreshToken != "" {
		var err error
		emailService, err = email.NewService(
			cfg.Email.ClientID,
			cfg.Email.ClientSecret,
			cfg.Email.RefreshToken,
			cfg.Email.FromEmail,
			cfg.Email.FromName,
		)
		if err != nil {
			log.Printf("Warning: Failed to initialize email service: %v", err)
			log.Printf("Email functionality will be disabled. Check your Gmail API credentials.")
		} else {
			emailHandler = email.NewHandler(emailService)
			// Set email service for user handler to send verification emails
			if cfg.Email.BaseURL != "" {
				userHandler.SetEmailService(emailService, cfg.Email.BaseURL)
			}
		}
	}

	router := server.NewRouter(locationHandler, userHandler, emailHandler, cfg.Auth.JWTSecret)

	// Use server address from config
	serverAddr := cfg.ServerAddress()
	log.Printf("Starting server on %s", serverAddr)
	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
