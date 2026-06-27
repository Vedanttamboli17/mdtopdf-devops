package handlers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/docplatform/api/services"
	"github.com/docplatform/shared/models"
)

// Events handles GET /events/:file_id.
// Opens a Server-Sent Events stream that delivers a single job_update event
// when the job completes, then closes. Heartbeats every 15 s prevent proxy
// timeouts. GET /status/:file_id continues to work for polling clients.
func Events(c *fiber.Ctx, cm *services.ConnectionManager) error {
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
		return c.Status(500).JSON(fiber.Map{"error": "Failed to check job status"})
	}
	if job == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Job not found"})
	}
	if job.TenantID != tenant.TenantID {
		return c.Status(403).JSON(fiber.Map{"error": "Forbidden"})
	}

	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("X-Accel-Buffering", "no")

	// Already in a terminal state — respond immediately without long-polling
	if job.Status == "completed" || job.Status == "failed" {
		c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
			sseJobUpdate(w, services.JobEvent{
				FileID:       job.ID,
				Status:       job.Status,
				DownloadURL:  job.OutputURL,
				ThumbnailURL: job.ThumbnailURL,
			})
		})
		return nil
	}

	// Register before SetBodyStreamWriter so any notification arriving in the gap
	// between here and the stream writer starting is buffered (cap 1) and not lost.
	ch := cm.Register(fileID)

	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		defer cm.Deregister(fileID, ch) // safe: no-op if Notify already removed the entry

		// Drain an event that may have arrived before the stream writer started
		select {
		case event, ok := <-ch:
			if ok {
				sseJobUpdate(w, event)
			}
			return
		default:
		}

		// Confirm to the client that the stream is live
		fmt.Fprintf(w, "event: connected\ndata: {\"file_id\":%q}\n\n", fileID)
		if err := w.Flush(); err != nil {
			return // Client disconnected before first byte
		}

		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case event, ok := <-ch:
				if !ok {
					// Channel closed: a newer SSE connection superseded ours
					return
				}
				sseJobUpdate(w, event)
				return // One terminal event per stream

			case <-ticker.C:
				fmt.Fprintf(w, "event: heartbeat\ndata: {\"ts\":%d}\n\n", time.Now().Unix())
				if err := w.Flush(); err != nil {
					return // Client disconnected; detected on flush error
				}
			}
		}
	})

	return nil
}

func sseJobUpdate(w *bufio.Writer, event services.JobEvent) {
	type payload struct {
		FileID       string `json:"file_id"`
		Status       string `json:"status"`
		DownloadURL  string `json:"download_url,omitempty"`
		ThumbnailURL string `json:"thumbnail_url,omitempty"`
	}
	data, _ := json.Marshal(payload{
		FileID:       event.FileID,
		Status:       event.Status,
		DownloadURL:  event.DownloadURL,
		ThumbnailURL: event.ThumbnailURL,
	})
	fmt.Fprintf(w, "event: job_update\ndata: %s\n\n", data)
	w.Flush()
}
