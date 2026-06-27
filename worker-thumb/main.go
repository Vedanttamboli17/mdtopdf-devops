package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/docplatform/shared/metrics"
	"github.com/docplatform/shared/queue"
	"github.com/docplatform/shared/webhook"
	"github.com/docplatform/worker-thumb/processor"
	"github.com/docplatform/worker-thumb/services"
)

func main() {
	_ = godotenv.Load()

	for _, init := range []struct {
		name string
		fn   func() error
	}{
		{"S3", services.InitS3},
		{"DynamoDB", services.InitDynamoDB},
		{"SQS", services.InitSQS},
	} {
		if err := init.fn(); err != nil {
			slog.Error("Init failed", "service", "worker-thumb", "component", init.name, "error", err)
			os.Exit(1)
		}
	}

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		slog.Info("Metrics server listening", "service", "worker-thumb", "port", 9090)
		if err := http.ListenAndServe(":9090", mux); err != nil {
			slog.Error("Metrics server error", "service", "worker-thumb", "error", err)
		}
	}()

	slog.Info("Worker started", "service", "worker-thumb", "event", "startup")

	for {
		messages, err := services.PollMessages()
		if err != nil {
			slog.Error("SQS poll error", "service", "worker-thumb", "error", err)
			time.Sleep(5 * time.Second)
			continue
		}
		metrics.SQSMessagesPolled.WithLabelValues("worker-thumb").Add(float64(len(messages)))
		for _, msg := range messages {
			go processJob(msg)
		}
	}
}

func processJob(msg services.ReceivedMessage) {
	start := time.Now()
	fileID := msg.Payload.FileID
	pdfS3Key := msg.Payload.PDFS3Key
	jobType := "thumbnail"
	tenantID := msg.Payload.TenantID

	slog.Info("Processing thumbnail",
		"service", "worker-thumb",
		"event", "job_received",
		"job_id", fileID,
		"tenant_id", tenantID,
		"job_type", jobType,
	)

	// Download the completed PDF from S3
	pdfContent, err := services.DownloadFile(pdfS3Key)
	if err != nil {
		slog.Error("S3 download failed", "service", "worker-thumb", "file_id", fileID, "error", err)
		return // message stays in SQS for retry
	}

	// Render first page to JPEG
	thumbPath := fmt.Sprintf("/tmp/%s-thumb.jpg", fileID)
	if err := processor.GenerateThumbnail(pdfContent, thumbPath); err != nil {
		slog.Error("Thumbnail generation failed", "service", "worker-thumb", "file_id", fileID, "error", err)
		return
	}
	defer os.Remove(thumbPath)

	// Upload thumbnail to S3
	thumbKey := fmt.Sprintf("thumbnails/%s.jpg", fileID)
	if err := services.UploadThumbnail(thumbKey, thumbPath); err != nil {
		slog.Error("Thumbnail upload failed", "service", "worker-thumb", "file_id", fileID, "error", err)
		return
	}

	// Generate presigned URL (10 min expiry, same as PDF URLs)
	thumbURL, err := services.GeneratePresignedURL(thumbKey)
	if err != nil {
		slog.Error("Presign failed", "service", "worker-thumb", "file_id", fileID, "error", err)
		return
	}

	// Stamp thumbnail_url onto the job record
	if err := services.UpdateThumbnailURL(fileID, thumbURL); err != nil {
		slog.Error("DynamoDB update failed", "service", "worker-thumb", "file_id", fileID, "error", err)
		return
	}

	// Generate a fresh presigned URL for the PDF — thumb worker is the last step in
	// the pipeline so it owns the final SSE notification that includes both URLs.
	pdfURL, err := services.GeneratePresignedURL(pdfS3Key)
	if err != nil {
		slog.Warn("Could not presign PDF URL for notification", "service", "worker-thumb", "file_id", fileID, "error", err)
		pdfURL = ""
	}

	if err := services.SendNotificationMessage(queue.NotificationMessage{
		FileID:       fileID,
		Status:       "completed",
		DownloadURL:  pdfURL,
		ThumbnailURL: thumbURL,
		TenantID:     msg.Payload.TenantID,
	}); err != nil {
		slog.Warn("Notification send failed", "service", "worker-thumb", "file_id", fileID, "error", err)
	}

	// Webhook callback — thumb worker is the last step, so it owns the final
	// HTTP callback. Best-effort: delivery failure never fails the job.
	if job, err := services.GetJob(fileID); err != nil {
		slog.Warn("Webhook skipped: job lookup failed", "service", "worker-thumb", "file_id", fileID, "error", err)
	} else if job != nil && job.WebhookURL != "" {
		tenant, terr := services.GetTenantByID(job.TenantID)
		if terr != nil || tenant == nil {
			slog.Warn("Webhook skipped: tenant lookup failed",
				"service", "worker-thumb", "file_id", fileID, "tenant_id", job.TenantID, "error", terr)
		} else {
			payload := webhook.WebhookPayload{
				FileID:       fileID,
				Status:       "completed",
				JobType:      job.JobType,
				DownloadURL:  pdfURL,
				ThumbnailURL: thumbURL,
				CompletedAt:  time.Now().UTC().Format(time.RFC3339),
			}
			// Dispatch blocks for up to ~130s across retries; run it off the
			// hot path so it doesn't delay metrics/logging for this message.
			go func(url, secret string, p webhook.WebhookPayload) {
				if err := webhook.Dispatch(url, secret, p); err != nil {
					slog.Error("Webhook delivery failed",
						"service", "worker-thumb", "file_id", p.FileID, "error", err)
				} else {
					slog.Info("Webhook delivered",
						"service", "worker-thumb", "file_id", p.FileID, "job_type", p.JobType)
				}
			}(job.WebhookURL, tenant.WebhookSecret, payload)
		}
	}

	services.DeleteMessage(msg.ReceiptHandle)
	duration := time.Since(start).Seconds()
	metrics.JobsCompletedTotal.WithLabelValues(jobType, tenantID).Inc()
	metrics.JobDuration.WithLabelValues(jobType).Observe(duration)
	slog.Info("Thumbnail generated",
		"service", "worker-thumb",
		"event", "thumbnail_complete",
		"job_id", fileID,
		"tenant_id", tenantID,
		"job_type", jobType,
		"duration_ms", int64(duration*1000),
	)
}
