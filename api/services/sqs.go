package services

import (
    "context"
    "encoding/json"
    "fmt"
    "os"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/sqs"
)

var sqsClient *sqs.Client

type SQSMessage struct {
    FileID string `json:"file_id"`
    S3Key  string `json:"s3_key"`
}

func InitSQS() error {
    cfg, err := config.LoadDefaultConfig(context.TODO(),
        config.WithRegion(os.Getenv("AWS_REGION")),
    )
    if err != nil {
        return fmt.Errorf("failed to load AWS config: %w", err)
    }
    sqsClient = sqs.NewFromConfig(cfg)
    return nil
}

func SendMessage(msg SQSMessage) error {
    body, err := json.Marshal(msg)
    if err != nil {
        return fmt.Errorf("marshal failed: %w", err)
    }
    _, err = sqsClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
        QueueUrl:    aws.String(os.Getenv("SQS_QUEUE_URL")),
        MessageBody: aws.String(string(body)),
    })
    return err
}