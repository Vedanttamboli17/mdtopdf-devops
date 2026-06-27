package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// Correlation assigns a unique X-Request-ID to every request.
// If the client supplies one we echo it back; otherwise we generate a UUID.
// The ID is stored in c.Locals("request_id") for use in log lines.
func Correlation(c *fiber.Ctx) error {
	requestID := c.Get("X-Request-ID")
	if requestID == "" {
		requestID = uuid.New().String()
	}
	c.Locals("request_id", requestID)
	c.Set("X-Request-ID", requestID)
	return c.Next()
}
