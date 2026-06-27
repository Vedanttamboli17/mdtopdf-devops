package services

import (
    "context"
    "encoding/json"
    "fmt"
    "os"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/sqs"
    "github.com/aws/aws-sdk-go-v2/service/sqs/types"
    "github.com/docplatform/shared/queue"
)

var sqsClient *sqs.Client

type ReceivedMessage struct {
    Payload       queue.JobMessage
    ReceiptHandle string
}

func InitSQS() error {
    cfg, err := config.LoadDefaultConfig(context.TODO(),
        config.WithRegion(os.Getenv("AWS_REGION")),
    )
    if err != nil {
        return fmt.Errorf("SQS config error: %w", err)
    }
    sqsClient = sqs.NewFromConfig(cfg)
    return nil
}

func PollMessages() ([]ReceivedMessage, error) {
    result, err := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
        QueueUrl:            aws.String(os.Getenv("SQS_QUEUE_URL")),
        MaxNumberOfMessages: 5,
        WaitTimeSeconds:     20, // Long polling
        AttributeNames:      []types.QueueAttributeName{"ApproximateReceiveCount"},
    })
    if err != nil {
        return nil, err
    }

    var messages []ReceivedMessage
    for _, m := range result.Messages {
        var payload queue.JobMessage
        if err := json.Unmarshal([]byte(*m.Body), &payload); err != nil {
            continue
        }
        messages = append(messages, ReceivedMessage{
            Payload:       payload,
            ReceiptHandle: *m.ReceiptHandle,
        })
    }
    return messages, nil
}

func DeleteMessage(receiptHandle string) error {
    _, err := sqsClient.DeleteMessage(context.TODO(), &sqs.DeleteMessageInput{
        QueueUrl:      aws.String(os.Getenv("SQS_QUEUE_URL")),
        ReceiptHandle: aws.String(receiptHandle),
    })
    return err
}

// SendNotificationMessage publishes a terminal job result to the notification queue
// so the API's SSE poller can push it to waiting clients. No-op if unconfigured.
func SendNotificationMessage(msg queue.NotificationMessage) error {
    notifURL := os.Getenv("NOTIFICATION_SQS_QUEUE_URL")
    if notifURL == "" {
        return nil
    }
    body, err := json.Marshal(msg)
    if err != nil {
        return fmt.Errorf("marshal failed: %w", err)
    }
    _, err = sqsClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
        QueueUrl:    aws.String(notifURL),
        MessageBody: aws.String(string(body)),
    })
    return err
}

// SendThumbnailMessage publishes to the thumbnail worker's queue.
// If THUMB_SQS_QUEUE_URL is unset the call is a no-op (thumbnail is optional).
func SendThumbnailMessage(msg queue.ThumbnailMessage) error {
    thumbURL := os.Getenv("THUMB_SQS_QUEUE_URL")
    if thumbURL == "" {
        return nil
    }
    body, err := json.Marshal(msg)
    if err != nil {
        return fmt.Errorf("marshal failed: %w", err)
    }
    _, err = sqsClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
        QueueUrl:    aws.String(thumbURL),
        MessageBody: aws.String(string(body)),
    })
    return err
}
