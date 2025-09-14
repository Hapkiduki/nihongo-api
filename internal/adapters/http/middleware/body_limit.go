package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// BodyLimit creates a middleware that limits the request body size
// Default limit is 64KB for webhooks to prevent DoS
func BodyLimit(maxBytes int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if len(c.Body()) > maxBytes {
			return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
				"error": "Request body too large",
			})
		}
		return c.Next()
	}
}

// WebhookBodyLimit limits webhook bodies to 64KB
func WebhookBodyLimit() fiber.Handler {
	return BodyLimit(64 * 1024) // 64KB
}
