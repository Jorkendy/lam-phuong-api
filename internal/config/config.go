package config

import (
	"fmt"
	"log"
	"strings"

	"lam-phuong-api/internal/airtable"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Airtable AirtableConfig `mapstructure:"airtable"`
	Auth     AuthConfig     `mapstructure:"auth"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port         string `mapstructure:"port"`
	Host         string `mapstructure:"host"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

// AirtableConfig holds Airtable-related configuration
type AirtableConfig struct {
	APIKey             string `mapstructure:"api_key"`
	BaseID             string `mapstructure:"base_id"`
	LocationsTableName string `mapstructure:"locations_table_name"`
	UsersTableName     string `mapstructure:"users_table_name"`
}

// AuthConfig holds authentication-related configuration
type AuthConfig struct {
	JWTSecret   string `mapstructure:"jwt_secret"`
	TokenExpiry int    `mapstructure:"token_expiry"` // in hours
}

var (
	// Global config instance
	globalConfig *Config
)

// Load initializes and loads configuration from environment variables and .env file
func Load() (*Config, error) {
	// Load .env file first (if it exists) - this loads env vars into the environment
	// Then viper will pick them up via AutomaticEnv()
	if err := godotenv.Load(); err != nil {
		// .env file not found; this is OK if we're using system env vars
		log.Printf(".env file not found, using system environment variables and defaults")
	} else {
		log.Printf("Loaded .env file")
	}

	// Enable environment variables
	// Viper will automatically read from environment variables
	viper.AutomaticEnv()
	// Replace dots with underscores for nested config keys
	// e.g., server.port becomes SERVER_PORT
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set default values
	setDefaults()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	globalConfig = &config
	return &config, nil
}

// Get returns the global configuration instance
func Get() *Config {
	if globalConfig == nil {
		log.Fatal("Config not loaded. Call config.Load() first.")
	}
	return globalConfig
}

// setDefaults sets default values for configuration
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.read_timeout", 15)
	viper.SetDefault("server.write_timeout", 15)

	// Airtable defaults (empty - should be set via env vars)
	viper.SetDefault("airtable.api_key", "")
	viper.SetDefault("airtable.base_id", "")
	viper.SetDefault("airtable.locations_table_name", "Địa điểm")
	viper.SetDefault("airtable.users_table_name", "Người dùng")

	// Auth defaults
	viper.SetDefault("auth.jwt_secret", "")
	viper.SetDefault("auth.token_expiry", 24) // 24 hours
}

// Validate checks if required configuration values are set
func (c *Config) Validate() error {
	if c.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}

	if c.Airtable.APIKey == "" {
		return fmt.Errorf("airtable API key is required (set AIRTABLE_API_KEY)")
	}

	if c.Airtable.BaseID == "" {
		return fmt.Errorf("airtable base ID is required (set AIRTABLE_BASE_ID)")
	}

	// LocationsTableName has a default value, so it's optional but we ensure it's set
	if c.Airtable.LocationsTableName == "" {
		c.Airtable.LocationsTableName = "Địa điểm" // Fallback to default if somehow empty
	}

	// UsersTableName has a default value, so it's optional but we ensure it's set
	if c.Airtable.UsersTableName == "" {
		c.Airtable.UsersTableName = "Người dùng" // Fallback to default if somehow empty
	}

	// Validate auth config
	if c.Auth.JWTSecret == "" {
		return fmt.Errorf("JWT secret is required (set AUTH_JWT_SECRET)")
	}

	if c.Auth.TokenExpiry <= 0 {
		c.Auth.TokenExpiry = 24 // Default to 24 hours
	}

	return nil
}

// ServerAddress returns the full server address (host:port)
func (c *Config) ServerAddress() string {
	return fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port)
}

// NewAirtableClient creates a new Airtable client using the configuration
func (c *Config) NewAirtableClient() (*airtable.Client, error) {
	return airtable.NewClient(c.Airtable.APIKey, c.Airtable.BaseID)
}
