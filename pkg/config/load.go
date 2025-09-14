package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config represents the application configuration structure
type Config struct {
	Server     ServerConfig     `mapstructure:"server" validate:"required"`
	Database   DatabaseConfig   `mapstructure:"database" validate:"required"`
	Auth       AuthConfig       `mapstructure:"auth" validate:"required"`
	RevenueCat RevenueCatConfig `mapstructure:"revenuecat" validate:"required"`
}

// ServerConfig holds server-related settings
type ServerConfig struct {
	Port string `mapstructure:"port" validate:"required"`
}

// DatabaseConfig holds database connection settings
type DatabaseConfig struct {
	MongoURI  string `mapstructure:"mongo_uri" validate:"required"`
	RedisAddr string `mapstructure:"redis_addr" validate:"required"`
}

// AuthConfig holds authentication settings
type AuthConfig struct {
	JWTSecret string `mapstructure:"jwt_secret" validate:"required,min=32"`
}

// RevenueCatConfig holds RevenueCat integration settings
type RevenueCatConfig struct {
	APIKey  string `mapstructure:"api_key" validate:"required"`
	BaseURL string `mapstructure:"base_url" validate:"required,url"`
	// WebhookSecrets is a comma-separated list of accepted webhook secrets for rotation
	WebhookSecrets string `mapstructure:"webhook_secrets" validate:"required"`
}

// Load loads and validates the configuration
func Load() (*Config, error) {
	// Load environment-specific .env file
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}
	envFile := ".env." + env
	if _, err := os.Stat(envFile); err == nil {
		if err := godotenv.Load(envFile); err != nil {
			return nil, fmt.Errorf("error loading %s: %w", envFile, err)
		}
		log.Printf("Loaded environment file: %s", envFile)
	} else {
		// Fallback to .env
		if _, err := os.Stat(".env"); err == nil {
			if err := godotenv.Load(".env"); err != nil {
				return nil, fmt.Errorf("error loading .env: %w", err)
			}
			log.Println("Loaded fallback .env")
		} else {
			log.Println("No .env file found, using system env vars")
		}
	}

	v := viper.New()

	// Set environment prefix and automatic env binding
	v.SetEnvPrefix("APP")
	v.AutomaticEnv()
	// Replace dots and underscores in env keys for nested configs
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// Load config.yml as fallback for defaults (only non-sensitive)
	v.SetConfigName("config")
	v.SetConfigType("yml")
	v.AddConfigPath(".")
	if err := v.MergeInConfig(); err != nil {
		log.Printf("Warning: could not load config.yml, using env vars only: %v", err)
	}

	// Unmarshal to Config struct
	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate the config
	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	log.Println("Configuration loaded successfully")
	return cfg, nil
}
