package main

import (
	"nihongo-api/internal/adapters/http/router"
	"nihongo-api/internal/adapters/storage/mongo"
	"nihongo-api/internal/application/service"
	"nihongo-api/pkg/config"
	"nihongo-api/pkg/database"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

func main() {
	// Initialize logger
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	if os.Getenv("APP_ENV") == "production" {
		logger = logger.Level(zerolog.InfoLevel)
	} else {
		logger = logger.Level(zerolog.DebugLevel)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load config")
	}

	// Initialize database
	db, err := database.ConnectMongo(cfg.Database.MongoURI)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to MongoDB")
	}
	defer func() {
		if err := database.CloseMongo(db.Client()); err != nil {
			logger.Error().Err(err).Msg("Error disconnecting from MongoDB")
		}
	}()

	// Initialize Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.Database.RedisAddr,
	})
	defer func() {
		if err := rdb.Close(); err != nil {
			logger.Error().Err(err).Msg("Error closing Redis connection")
		}
	}()

	// Initialize repositories
	userRepo := mongo.NewMongoUserRepository(db)
	subRepo := mongo.NewMongoSubscriptionRepository(db)
	syllableRepo := mongo.NewMongoSyllableRepository(db)
	courseRepo := mongo.NewMongoCourseRepository(db)
	kanjiRepo := mongo.NewMongoKanjiRepository(db)
	progressRepo := mongo.NewMongoProgressRepository(db)

	// Initialize services
	userService := service.NewUserService(userRepo, subRepo, logger)
	subscriptionService := service.NewSubscriptionService(subRepo, userRepo, userService, logger)
	courseService := service.NewCourseService(courseRepo)
	progressService := service.NewProgressService(progressRepo)

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			logger.Error().Err(err).Str("path", c.Path()).Msg("Request error")
			return c.Status(500).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Setup routes
	// Prefer config revenuecat.webhook_secrets (comma-separated) but fall back to env var
	webhookSecrets := []string{}
	if cfg.RevenueCat.WebhookSecrets != "" {
		for _, s := range strings.Split(cfg.RevenueCat.WebhookSecrets, ",") {
			ss := strings.TrimSpace(s)
			if ss != "" {
				webhookSecrets = append(webhookSecrets, ss)
			}
		}
	} else {
		envSecret := os.Getenv("APP_REVENUECAT_WEBHOOK_SECRET")
		if envSecret != "" {
			for _, s := range strings.Split(envSecret, ",") {
				ss := strings.TrimSpace(s)
				if ss != "" {
					webhookSecrets = append(webhookSecrets, ss)
				}
			}
		}
	}
	if len(webhookSecrets) == 0 {
		logger.Fatal().Msg("APP_REVENUECAT_WEBHOOK_SECRET(s) required")
	}
	router.SetupRoutes(app, userService, subscriptionService, courseService, progressService, syllableRepo, kanjiRepo, cfg.Auth.JWTSecret, webhookSecrets, logger)

	// Start server
	go func() {
		logger.Info().Str("port", cfg.Server.Port).Msg("Server starting")
		if err := app.Listen(":" + cfg.Server.Port); err != nil {
			logger.Error().Err(err).Msg("Server error")
		}
	}()

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	logger.Info().Msg("Shutting down server...")

	if err := app.Shutdown(); err != nil {
		logger.Error().Err(err).Msg("Server shutdown error")
	}

	logger.Info().Msg("Server stopped")
}
