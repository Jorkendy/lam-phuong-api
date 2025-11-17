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

	// Initialize seed data
	locationSeed := []location.Location{
		{ID: "1", Name: "Main Library", Slug: "main-library"},
		{ID: "2", Name: "West Branch", Slug: "west-branch"},
	}

	// Create in-memory repository
	baseRepo := location.NewInMemoryRepository(locationSeed)

	// Wrap with Airtable repository for persistence
	airtableClient, err := cfg.NewAirtableClient()
	if err != nil {
		log.Fatalf("Failed to create Airtable client: %v", err)
	}
	locationRepo := location.NewAirtableRepository(baseRepo, airtableClient, cfg.Airtable.LocationsTableName)

	locationHandler := location.NewHandler(locationRepo)

	// Initialize user seed data
	userSeed := []user.User{}

	// Create in-memory user repository
	baseUserRepo := user.NewInMemoryRepository(userSeed)

	// Wrap with Airtable repository for persistence
	userRepo := user.NewAirtableRepository(baseUserRepo, airtableClient, cfg.Airtable.UsersTableName)

	// Initialize email service
	emailService := email.NewService(
		cfg.Email.SMTPHost,
		cfg.Email.SMTPPort,
		cfg.Email.SMTPUsername,
		cfg.Email.SMTPPassword,
		cfg.Email.FromEmail,
		cfg.Email.FromName,
	)

	// Create user handler with JWT configuration and email service
	tokenExpiry := time.Duration(cfg.Auth.TokenExpiry) * time.Hour
	userHandler := user.NewHandler(userRepo, cfg.Auth.JWTSecret, tokenExpiry, emailService, cfg.Email.BaseURL)

	router := server.NewRouter(locationHandler, userHandler, cfg.Auth.JWTSecret)

	// Use server address from config
	serverAddr := cfg.ServerAddress()
	log.Printf("Starting server on %s", serverAddr)
	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
