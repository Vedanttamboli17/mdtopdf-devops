package handlers

import (
    "log/slog"

    "github.com/gofiber/fiber/v2"
    "mdtopdf/api/services"
)

func Status(c *fiber.Ctx) error {
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

    response := fiber.Map{
        "file_id": job.ID,
        "status":  job.Status,
    }
    if job.Status == "completed" {
        response["download_url"] = job.OutputURL
    }

    return c.JSON(response)
}