package handlers

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/docplatform/api/services"
	"github.com/docplatform/shared/models"
)

func Status(c *fiber.Ctx) error {
	tenant, ok := c.Locals("tenant").(*models.Tenant)
	if !ok || tenant == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	fileID := c.Params("file_id")
	if fileID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "file_id is required"})
	}

	job, err := services.GetJob(fileID)
	if err != nil {
		slog.Error("DynamoDB get failed", "service", "api", "file_id", fileID, "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to retrieve job"})
	}
	if job == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Job not found"})
	}

	// Tenants can only see their own jobs
	if job.TenantID != tenant.TenantID {
		return c.Status(403).JSON(fiber.Map{"error": "Forbidden"})
	}

	response := fiber.Map{
		"file_id": job.ID,
		"status":  job.Status,
	}
	// Echo back the registered webhook, masked to the first 20 chars so the
	// client can confirm registration without us leaking the full URL.
	if job.WebhookURL != "" {
		masked := job.WebhookURL
		if len(masked) > 20 {
			masked = masked[:20] + "..."
		}
		response["webhook_url"] = masked
	}
	if job.Status == "completed" {
		response["download_url"] = job.OutputURL
		if job.ThumbnailURL != "" {
			response["thumbnail_url"] = job.ThumbnailURL
		}
	}

	return c.JSON(response)
}
