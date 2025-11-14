package main

import (
	"log"
	"time"

	_ "lam-phuong-api/docs" // Import docs for Swagger
	"lam-phuong-api/internal/config"
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

	// Create user handler with JWT configuration
	tokenExpiry := time.Duration(cfg.Auth.TokenExpiry) * time.Hour
	userHandler := user.NewHandler(userRepo, cfg.Auth.JWTSecret, tokenExpiry)

	router := server.NewRouter(locationHandler, userHandler, cfg.Auth.JWTSecret)

	// Use server address from config
	serverAddr := cfg.ServerAddress()
	log.Printf("Starting server on %s", serverAddr)
	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
