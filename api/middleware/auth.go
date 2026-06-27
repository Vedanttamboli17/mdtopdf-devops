package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/docplatform/api/services"
	"github.com/docplatform/shared/models"
)

type cacheEntry struct {
	tenant    *models.Tenant
	fetchedAt time.Time
}

var (
	tenantCache sync.Map
	cacheTTL    = 60 * time.Second
)

// Auth validates the X-API-Key header for every request except /health.
// On success it attaches the resolved *models.Tenant to c.Locals("tenant").
// Tenant records are cached in memory for 60 seconds to reduce DynamoDB reads.
func Auth(c *fiber.Ctx) error {
	if c.Path() == "/health" || c.Path() == "/metrics" {
		return c.Next()
	}

	// Header is preferred. The api_key query param is a fallback for the SSE
	// /events endpoint, because the browser EventSource API cannot send headers.
	apiKey := c.Get("X-API-Key")
	if apiKey == "" {
		apiKey = c.Query("api_key")
	}
	if apiKey == "" {
		return c.Status(401).JSON(fiber.Map{
			"error": "X-API-Key header or api_key query param is required",
		})
	}

	// Check in-memory cache
	if raw, ok := tenantCache.Load(apiKey); ok {
		entry := raw.(cacheEntry)
		if time.Since(entry.fetchedAt) < cacheTTL {
			if !entry.tenant.Active {
				return c.Status(401).JSON(fiber.Map{"error": "Tenant account is inactive"})
			}
			c.Locals("tenant", entry.tenant)
			return c.Next()
		}
		tenantCache.Delete(apiKey)
	}

	// Cache miss — fetch from DynamoDB
	tenant, err := services.GetTenantByAPIKey(apiKey)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Authentication service unavailable"})
	}
	if tenant == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid API key"})
	}
	if !tenant.Active {
		return c.Status(401).JSON(fiber.Map{"error": "Tenant account is inactive"})
	}

	tenantCache.Store(apiKey, cacheEntry{tenant: tenant, fetchedAt: time.Now()})
	c.Locals("tenant", tenant)
	return c.Next()
}
