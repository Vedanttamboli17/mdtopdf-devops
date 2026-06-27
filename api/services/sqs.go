package services

import (
    "context"
    "encoding/json"
    "fmt"
    "os"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/sqs"
    "github.com/docplatform/shared/queue"
)

var sqsClient *sqs.Client

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

func SendMessage(msg queue.JobMessage) error {
	return SendMessageToURL(os.Getenv("SQS_QUEUE_URL"), msg)
}

func SendMessageToURL(queueURL string, msg queue.JobMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}
	_, err = sqsClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
		QueueUrl:    aws.String(queueURL),
		MessageBody: aws.String(string(body)),
	})
	return err
}
