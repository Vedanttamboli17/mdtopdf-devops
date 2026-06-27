package services

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/docplatform/shared/queue"
)

// StartNotificationPoller runs forever, long-polling the notification SQS queue.
// Call as a goroutine from main() after InitSQS() has been called.
func StartNotificationPoller(cm *ConnectionManager) {
	queueURL := os.Getenv("NOTIFICATION_SQS_QUEUE_URL")
	if queueURL == "" {
		slog.Warn("NOTIFICATION_SQS_QUEUE_URL not set — SSE push notifications disabled")
		return
	}

	slog.Info("Notification poller started", "queue_url", queueURL)

	for {
		result, err := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(queueURL),
			MaxNumberOfMessages: 10,
			WaitTimeSeconds:     20, // Long-poll — saves CPU when queue is idle
		})
		if err != nil {
			slog.Error("Notification poll failed", "error", err)
			continue // WaitTimeSeconds already throttles empty-queue loops
		}

		for _, msg := range result.Messages {
			var notif queue.NotificationMessage
			if err := json.Unmarshal([]byte(*msg.Body), &notif); err != nil {
				slog.Error("Malformed notification message, discarding", "error", err)
				deleteNotif(queueURL, msg.ReceiptHandle)
				continue
			}

			cm.Notify(notif.FileID, JobEvent{
				FileID:       notif.FileID,
				Status:       notif.Status,
				DownloadURL:  notif.DownloadURL,
				ThumbnailURL: notif.ThumbnailURL,
			})

			deleteNotif(queueURL, msg.ReceiptHandle)
		}
	}
}

func deleteNotif(queueURL string, receiptHandle *string) {
	if _, err := sqsClient.DeleteMessage(context.TODO(), &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueURL),
		ReceiptHandle: receiptHandle,
	}); err != nil {
		slog.Error("Failed to delete notification message", "error", err)
	}
}
