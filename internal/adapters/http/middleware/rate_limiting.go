package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// InMemoryRateLimiter is a simple in-memory rate limiter for demonstration
// In production, use Redis or similar
type InMemoryRateLimiter struct {
	mu          sync.Mutex
	requests    map[string][]time.Time
	maxRequests int
	window      time.Duration
}

func NewInMemoryRateLimiter(maxRequests int, window time.Duration) *InMemoryRateLimiter {
	return &InMemoryRateLimiter{
		requests:    make(map[string][]time.Time),
		maxRequests: maxRequests,
		window:      window,
	}
}

func (rl *InMemoryRateLimiter) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := c.IP()

		rl.mu.Lock()
		now := time.Now()
		times := rl.requests[key]

		// Remove old requests outside the window
		var validTimes []time.Time
		for _, t := range times {
			if now.Sub(t) < rl.window {
				validTimes = append(validTimes, t)
			}
		}

		if len(validTimes) >= rl.maxRequests {
			rl.mu.Unlock()
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Rate limit exceeded",
			})
		}

		validTimes = append(validTimes, now)
		rl.requests[key] = validTimes
		rl.mu.Unlock()

		return c.Next()
	}
}
