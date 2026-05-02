package handlers

import (
    "fmt"
    "log/slog"
    "path/filepath"

    "github.com/gofiber/fiber/v2"
    "github.com/google/uuid"
    "mdtopdf/api/models"
    "mdtopdf/api/services"
)

func Upload(c *fiber.Ctx) error {
    file, err := c.FormFile("file")
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Provide a file with field name 'file'"})
    }

    // Validate file extension
    if filepath.Ext(file.Filename) != ".md" {
        return c.Status(400).JSON(fiber.Map{"error": "Only .md (Markdown) files are accepted"})
    }

    // Validate size (max 5MB)
    if file.Size > 5*1024*1024 {
        return c.Status(400).JSON(fiber.Map{"error": "File too large. Max 5MB allowed"})
    }

    fileID := uuid.New().String()
    s3Key := fmt.Sprintf("uploads/%s.md", fileID)

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
        ID:         fileID,
        Status:     "processing",
        InputURL:   s3Key,
        OutputURL:  "",
        RetryCount: 0,
    }
    if err := services.CreateJob(job); err != nil {
        slog.Error("DynamoDB write failed", "service", "api", "file_id", fileID, "error", err)
        return c.Status(500).JSON(fiber.Map{"error": "Failed to create job record"})
    }

    // Send message to SQS
    if err := services.SendMessage(services.SQSMessage{FileID: fileID, S3Key: s3Key}); err != nil {
        slog.Error("SQS send failed", "service", "api", "file_id", fileID, "error", err)
        return c.Status(500).JSON(fiber.Map{"error": "Failed to queue job"})
    }

    slog.Info("Job queued", "service", "api", "event", "job_created", "file_id", fileID)
    return c.Status(202).JSON(fiber.Map{
        "file_id": fileID,
        "status":  "processing",
        "message": "File uploaded. Poll /status/" + fileID + " to track progress.",
    })
}