package router

import (
	"nihongo-api/internal/application/service"
	"nihongo-api/internal/ports"
	"time"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

// SetupRoutes configures all HTTP routes
func SetupRoutes(app *fiber.App, userService *service.UserService, courseService *service.CourseService, progressService *service.ProgressService, syllableRepo ports.SyllableRepository) {
	api := app.Group("/api")

	// Health check
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Syllables routes
	syllables := api.Group("/syllables")
	syllables.Get("/", func(c *fiber.Ctx) error {
		syllables, err := syllableRepo.GetAll(c.Context())
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(syllables)
	})

	syllables.Get("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		syllable, err := syllableRepo.GetByID(c.Context(), id)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "Syllable not found"})
		}
		return c.JSON(syllable)
	})

	// Auth routes
	auth := api.Group("/auth")
	auth.Post("/register", func(c *fiber.Ctx) error {
		var req struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}

		user, err := userService.RegisterUser(c.Context(), req.Name, req.Email, req.Password)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		return c.Status(201).JSON(user)
	})

	auth.Post("/login", func(c *fiber.Ctx) error {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}

		user, err := userService.AuthenticateUser(c.Context(), req.Email, req.Password)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
		}

		// Generate JWT token
		token := jwt.New(jwt.SigningMethodHS256)
		claims := token.Claims.(jwt.MapClaims)
		claims["user_id"] = user.ID.Hex()
		claims["email"] = user.Email
		claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

		jwtSecret := viper.GetString("auth.jwt_secret")
		t, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to generate token"})
		}

		return c.JSON(fiber.Map{"token": t, "user": user})
	})

	// JWT middleware
	jwtSecret := viper.GetString("auth.jwt_secret")
	jwtMiddleware := jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(jwtSecret)},
	})

	// Protected routes
	protected := api.Group("/protected", jwtMiddleware)
	protected.Get("/courses", func(c *fiber.Ctx) error {
		courses, err := courseService.GetAllCourses(c.Context())
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(courses)
	})

	protected.Get("/profile", func(c *fiber.Ctx) error {
		user := c.Locals("user").(*jwt.Token)
		claims := user.Claims.(jwt.MapClaims)
		userID := claims["user_id"].(string)

		userData, err := userService.GetUserByID(c.Context(), userID)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "User not found"})
		}

		return c.JSON(userData)
	})
}
