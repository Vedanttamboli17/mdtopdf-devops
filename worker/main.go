package main

import (
    "fmt"
    "log/slog"
    "os"
    "time"

    "github.com/joho/godotenv"
    "mdtopdf/worker/processor"
    "mdtopdf/worker/services"
)

const maxRetries = 3

func main() {
    _ = godotenv.Load()

    // Initialize all AWS clients
    for _, init := range []struct {
        name string
        fn   func() error
    }{
        {"S3", services.InitS3},
        {"DynamoDB", services.InitDynamoDB},
        {"SQS", services.InitSQS},
    } {
        if err := init.fn(); err != nil {
            slog.Error("Init failed", "service", init.name, "error", err)
            os.Exit(1)
        }
    }

    slog.Info("Worker started", "service", "worker", "event", "startup")

    // Continuous polling loop
    for {
        messages, err := services.PollMessages()
        if err != nil {
            slog.Error("SQS poll error", "service", "worker", "error", err)
            time.Sleep(5 * time.Second)
            continue
        }

        for _, msg := range messages {
            go processJob(msg)
        }
    }
}

func processJob(msg services.ReceivedMessage) {
    fileID := msg.Payload.FileID
    s3Key := msg.Payload.S3Key

    slog.Info("Processing job", "service", "worker", "event", "job_received", "file_id", fileID)

    // Fetch current job state from DynamoDB
    job, err := services.GetJob(fileID)
    if err != nil || job == nil {
        slog.Error("Job not found", "service", "worker", "file_id", fileID)
        return // Don't delete — let SQS retry
    }

    // Application-level retry gate
    if job.RetryCount >= maxRetries {
        slog.Error("Max retries reached, marking failed",
            "service", "worker", "file_id", fileID, "retry_count", job.RetryCount)
        services.UpdateJobFailed(fileID)
        services.DeleteMessage(msg.ReceiptHandle)
        return
    }

    // Download Markdown from S3
    mdContent, err := services.DownloadFile(s3Key)
    if err != nil {
        slog.Error("S3 download failed", "service", "worker", "file_id", fileID, "error", err)
        services.IncrementRetry(fileID, job.RetryCount)
        return // Message stays in SQS for next attempt
    }

    // Convert Markdown → PDF
    pdfPath := fmt.Sprintf("/tmp/%s.pdf", fileID)
    if err := processor.ConvertMDToPDF(mdContent, pdfPath); err != nil {
        slog.Error("Conversion failed", "service", "worker", "file_id", fileID, "error", err)
        services.IncrementRetry(fileID, job.RetryCount)
        return
    }
    defer os.Remove(pdfPath)

    // Upload PDF to S3
    pdfKey := fmt.Sprintf("pdfs/%s.pdf", fileID)
    if err := services.UploadPDF(pdfKey, pdfPath); err != nil {
        slog.Error("PDF upload failed", "service", "worker", "file_id", fileID, "error", err)
        services.IncrementRetry(fileID, job.RetryCount)
        return
    }

    // Generate presigned download URL
    url, err := services.GeneratePresignedURL(pdfKey)
    if err != nil {
        slog.Error("Presign failed", "service", "worker", "file_id", fileID, "error", err)
        services.IncrementRetry(fileID, job.RetryCount)
        return
    }

    // Mark job as completed
    if err := services.UpdateJobCompleted(fileID, url); err != nil {
        slog.Error("DynamoDB update failed", "service", "worker", "file_id", fileID, "error", err)
        return
    }

    // Remove message from SQS — success
    services.DeleteMessage(msg.ReceiptHandle)
    slog.Info("Job completed",
        "service", "worker",
        "event", "conversion_complete",
        "file_id", fileID,
    )
}