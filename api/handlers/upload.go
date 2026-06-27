package handlers

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/docplatform/api/services"
	"github.com/docplatform/shared/metrics"
	"github.com/docplatform/shared/models"
	"github.com/docplatform/shared/queue"
)

func Upload(c *fiber.Ctx) error {
	tenant, ok := c.Locals("tenant").(*models.Tenant)
	if !ok || tenant == nil {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// Enforce monthly quota
	if tenant.JobsThisMonth >= tenant.MonthlyQuota {
		return c.Status(429).JSON(fiber.Map{
			"error": "Monthly quota exceeded",
			"quota": tenant.MonthlyQuota,
			"used":  tenant.JobsThisMonth,
		})
	}

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Provide a file with field name 'file'"})
	}

	ext := filepath.Ext(file.Filename)
	if ext != ".md" && ext != ".docx" {
		return c.Status(400).JSON(fiber.Map{"error": "Only .md and .docx files are accepted"})
	}

	// Per-format size limit: 5MB for Markdown, 20MB for DOCX
	maxSize := int64(5 * 1024 * 1024)
	if ext == ".docx" {
		maxSize = 20 * 1024 * 1024
	}
	if file.Size > maxSize {
		return c.Status(400).JSON(fiber.Map{
			"error": fmt.Sprintf("File too large. Max %dMB allowed for %s files", maxSize/(1024*1024), ext),
		})
	}

	// Optional webhook callback URL. Must be HTTPS — we refuse plain HTTP so
	// the HMAC-signed payload (and any presigned URLs in it) is not sent in
	// cleartext.
	webhookURL := c.FormValue("webhook_url")
	if webhookURL != "" {
		u, perr := url.Parse(webhookURL)
		if perr != nil || u.Scheme != "https" || u.Host == "" {
			return c.Status(400).JSON(fiber.Map{"error": "webhook_url must be a valid HTTPS URL"})
		}
	}

	fileID := uuid.New().String()

	var (
		s3Key   string
		jobType string
		format  string
	)
	switch ext {
	case ".md":
		s3Key = fmt.Sprintf("uploads/%s.md", fileID)
		jobType = "markdown-pdf"
		format = "md"
	case ".docx":
		s3Key = fmt.Sprintf("uploads/%s.docx", fileID)
		jobType = "docx-pdf"
		format = "docx"
	}

	f, err := file.Open()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Could not open file"})
	}
	defer f.Close()

	// Upload to S3
	if err := services.UploadMarkdown(s3Key, f); err != nil {
		slog.Error("S3 upload failed", "service", "api", "file_id", fileID, "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to store file"})
	}

	// Create job record in DynamoDB
	job := models.Job{
		ID:          fileID,
		TenantID:    tenant.TenantID,
		JobType:     jobType,
		Status:      "processing",
		Format:      format,
		Priority:    1,
		InputURL:    s3Key,
		OutputURL:   "",
		WebhookURL:  webhookURL,
		RetryCount:  0,
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
		CompletedAt: "",
	}
	if err := services.CreateJob(job); err != nil {
		slog.Error("DynamoDB write failed", "service", "api", "file_id", fileID, "error", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create job record"})
	}

	// Route to the correct SQS queue based on file type
	msg := queue.JobMessage{
		FileID:   fileID,
		S3Key:    s3Key,
		JobType:  jobType,
		TenantID: tenant.TenantID,
		Priority: 1,
	}

	var sendErr error
	switch ext {
	case ".md":
		sendErr = services.SendMessage(msg)
	case ".docx":
		docxQueueURL := os.Getenv("DOCX_SQS_QUEUE_URL")
		if docxQueueURL == "" {
			return c.Status(503).JSON(fiber.Map{"error": "DOCX processing not configured"})
		}
		sendErr = services.SendMessageToURL(docxQueueURL, msg)
	}

	if sendErr != nil {
		slog.Error("SQS send failed", "service", "api", "file_id", fileID, "error", sendErr)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to queue job"})
	}

	// Increment job counter — best-effort, non-fatal
	if err := services.IncrementJobCount(tenant.APIKey); err != nil {
		slog.Warn("Failed to increment job count", "service", "api", "tenant_id", tenant.TenantID, "error", err)
	}

	metrics.JobsReceivedTotal.WithLabelValues(jobType, tenant.TenantID).Inc()
	slog.Info("Job queued",
		"service", "api",
		"event", "job_created",
		"request_id", c.Locals("request_id"),
		"job_id", fileID,
		"tenant_id", tenant.TenantID,
		"job_type", jobType,
	)
	return c.Status(202).JSON(fiber.Map{
		"file_id": fileID,
		"status":  "processing",
		"message": "File uploaded. Poll /status/" + fileID + " to track progress.",
	})
}
