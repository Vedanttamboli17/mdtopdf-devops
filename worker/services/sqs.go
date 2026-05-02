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
)

var sqsClient *sqs.Client

type SQSMessage struct {
    FileID string `json:"file_id"`
    S3Key  string `json:"s3_key"`
}

type ReceivedMessage struct {
    Payload       SQSMessage
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
        var payload SQSMessage
        if err := json.Unmarshal([]byte(*m.Body), &payload); err != nil {
            continue // skip malformed messages
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