package main

import (
	"context"
	"log"
	"nihongo-api/internal/adapters/http/router"
	"nihongo-api/internal/adapters/storage/mongo"
	"nihongo-api/internal/application/service"
	"nihongo-api/pkg/database"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
)

func main() {
	// Load configuration
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	// Initialize database
	db, err := database.ConnectMongo(viper.GetString("database.mongo_uri"))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := db.Client().Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}()

	// Initialize repositories
	userRepo := mongo.NewMongoUserRepository(db)
	syllableRepo := mongo.NewMongoSyllableRepository(db)

	// Initialize services
	userService := service.NewUserService(userRepo)
	courseService := service.NewCourseService(nil)     // TODO: implement course repo
	progressService := service.NewProgressService(nil) // TODO: implement progress repo

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(500).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Setup routes
	router.SetupRoutes(app, userService, courseService, progressService, syllableRepo)

	// Start server
	port := viper.GetString("server.port")
	go func() {
		log.Printf("Server starting on port %s", port)
		if err := app.Listen(":" + port); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	log.Println("Shutting down server...")

	if err := app.Shutdown(); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
