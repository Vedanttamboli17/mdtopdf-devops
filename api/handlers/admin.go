package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/docplatform/api/services"
	"github.com/docplatform/shared/models"
)

// CreateTenant handles POST /admin/tenants.
// Protected by the ADMIN_SECRET env var via X-Admin-Secret header.
func CreateTenant(c *fiber.Ctx) error {
	adminSecret := os.Getenv("ADMIN_SECRET")
	if adminSecret == "" {
		return c.Status(503).JSON(fiber.Map{"error": "Admin endpoint not configured"})
	}
	if c.Get("X-Admin-Secret") != adminSecret {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid admin secret"})
	}

	var body struct {
		Name          string `json:"name"`
		Plan          string `json:"plan"`
		RateLimit     int    `json:"rate_limit"`
		MonthlyQuota  int    `json:"monthly_quota"`
		WebhookSecret string `json:"webhook_secret"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if body.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "name is required"})
	}
	if body.Plan == "" {
		body.Plan = "free"
	}
	if body.RateLimit == 0 {
		body.RateLimit = 5
	}
	if body.MonthlyQuota == 0 {
		body.MonthlyQuota = 100
	}

	// sk_ + 32 hex chars (16 random bytes → 32 hex chars)
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to generate API key"})
	}
	apiKey := fmt.Sprintf("sk_%s", hex.EncodeToString(raw))

	tenant := models.Tenant{
		APIKey:          apiKey,
		TenantID:        uuid.New().String(),
		Name:            body.Name,
		Plan:            body.Plan,
		RateLimit:       body.RateLimit,
		MonthlyQuota:    body.MonthlyQuota,
		JobsThisMonth:   0,
		QuotaResetMonth: time.Now().UTC().Format("2006-01"),
		WebhookSecret:   body.WebhookSecret,
		CreatedAt:       time.Now().UTC().Format(time.RFC3339),
		Active:          true,
	}

	if err := services.CreateTenant(tenant); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create tenant"})
	}

	return c.Status(201).JSON(fiber.Map{
		"api_key":       tenant.APIKey,
		"tenant_id":     tenant.TenantID,
		"name":          tenant.Name,
		"plan":          tenant.Plan,
		"monthly_quota": tenant.MonthlyQuota,
	})
}
