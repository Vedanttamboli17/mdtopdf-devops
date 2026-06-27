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
	"github.com/docplatform/worker-docx/processor"
	"github.com/docplatform/worker-docx/services"
)

const maxRetries = 3

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
			slog.Error("Init failed", "service", "worker-docx", "component", init.name, "error", err)
			os.Exit(1)
		}
	}

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		slog.Info("Metrics server listening", "service", "worker-docx", "port", 9090)
		if err := http.ListenAndServe(":9090", mux); err != nil {
			slog.Error("Metrics server error", "service", "worker-docx", "error", err)
		}
	}()

	slog.Info("Worker started", "service", "worker-docx", "event", "startup")

	for {
		messages, err := services.PollMessages()
		if err != nil {
			slog.Error("SQS poll error", "service", "worker-docx", "error", err)
			time.Sleep(5 * time.Second)
			continue
		}
		metrics.SQSMessagesPolled.WithLabelValues("worker-docx").Add(float64(len(messages)))
		for _, msg := range messages {
			go processJob(msg)
		}
	}
}

func processJob(msg services.ReceivedMessage) {
	start := time.Now()
	fileID := msg.Payload.FileID
	s3Key := msg.Payload.S3Key
	jobType := "docx-pdf"
	tenantID := msg.Payload.TenantID

	slog.Info("Processing job",
		"service", "worker-docx",
		"event", "job_received",
		"job_id", fileID,
		"tenant_id", tenantID,
		"job_type", jobType,
	)

	job, err := services.GetJob(fileID)
	if err != nil || job == nil {
		slog.Error("Job not found", "service", "worker-docx", "job_id", fileID)
		return
	}

	if job.RetryCount >= maxRetries {
		duration := time.Since(start).Seconds()
		slog.Error("Max retries reached, marking failed",
			"service", "worker-docx",
			"job_id", fileID,
			"tenant_id", tenantID,
			"retry_count", job.RetryCount,
			"duration_ms", int64(duration*1000),
		)
		services.UpdateJobFailed(fileID)
		services.DeleteMessage(msg.ReceiptHandle)
		metrics.JobsFailedTotal.WithLabelValues(jobType, "max_retries").Inc()
		metrics.JobDuration.WithLabelValues(jobType).Observe(duration)
		return
	}

	// Download DOCX from S3
	docxContent, err := services.DownloadFile(s3Key)
	if err != nil {
		slog.Error("S3 download failed", "service", "worker-docx", "file_id", fileID, "error", err)
		services.IncrementRetry(fileID, job.RetryCount)
		return
	}

	// Convert DOCX to PDF via LibreOffice
	pdfPath := fmt.Sprintf("/tmp/%s.pdf", fileID)
	if err := processor.ConvertDOCXToPDF(docxContent, pdfPath); err != nil {
		slog.Error("Conversion failed", "service", "worker-docx", "file_id", fileID, "error", err)
		services.IncrementRetry(fileID, job.RetryCount)
		return
	}
	defer os.Remove(pdfPath)

	// Upload PDF to S3
	pdfKey := fmt.Sprintf("pdfs/%s.pdf", fileID)
	if err := services.UploadPDF(pdfKey, pdfPath); err != nil {
		slog.Error("PDF upload failed", "service", "worker-docx", "file_id", fileID, "error", err)
		services.IncrementRetry(fileID, job.RetryCount)
		return
	}

	// Generate presigned download URL
	url, err := services.GeneratePresignedURL(pdfKey)
	if err != nil {
		slog.Error("Presign failed", "service", "worker-docx", "file_id", fileID, "error", err)
		services.IncrementRetry(fileID, job.RetryCount)
		return
	}

	if err := services.UpdateJobCompleted(fileID, url); err != nil {
		slog.Error("DynamoDB update failed", "service", "worker-docx", "file_id", fileID, "error", err)
		return
	}

	// Enqueue thumbnail generation (best-effort — thumbnail is post-processing)
	thumbEnqueued := false
	if err := services.SendThumbnailMessage(queue.ThumbnailMessage{
		FileID:   fileID,
		PDFS3Key: pdfKey,
		TenantID: msg.Payload.TenantID,
	}); err != nil {
		slog.Warn("Failed to enqueue thumbnail job", "service", "worker-docx", "file_id", fileID, "error", err)
	} else if os.Getenv("THUMB_SQS_QUEUE_URL") != "" {
		thumbEnqueued = true
	}

	// SSE notification
	if os.Getenv("NOTIFICATION_SQS_QUEUE_URL") != "" {
		if !thumbEnqueued {
			// No thumb worker active — push the PDF result immediately
			if err := services.SendNotificationMessage(queue.NotificationMessage{
				FileID:      fileID,
				Status:      "completed",
				DownloadURL: url,
				TenantID:    msg.Payload.TenantID,
			}); err != nil {
				slog.Warn("Notification send failed", "service", "worker-docx", "file_id", fileID, "error", err)
			}
		} else {
			// Thumb worker is the primary notifier. This 30s fallback unblocks the
			// SSE client if thumb fails; re-reading DynamoDB captures thumbnail_url
			// if thumb completed in time.
			go func() {
				time.Sleep(30 * time.Second)
				job, err := services.GetJob(fileID)
				if err != nil || job == nil {
					return
				}
				if err := services.SendNotificationMessage(queue.NotificationMessage{
					FileID:       fileID,
					Status:       job.Status,
					DownloadURL:  job.OutputURL,
					ThumbnailURL: job.ThumbnailURL,
					TenantID:     msg.Payload.TenantID,
				}); err != nil {
					slog.Warn("Fallback notification failed", "service", "worker-docx", "file_id", fileID, "error", err)
				}
			}()
		}
	}

	services.DeleteMessage(msg.ReceiptHandle)
	duration := time.Since(start).Seconds()
	metrics.JobsCompletedTotal.WithLabelValues(jobType, tenantID).Inc()
	metrics.JobDuration.WithLabelValues(jobType).Observe(duration)
	slog.Info("Job completed",
		"service", "worker-docx",
		"event", "conversion_complete",
		"job_id", fileID,
		"tenant_id", tenantID,
		"job_type", jobType,
		"duration_ms", int64(duration*1000),
	)
}
