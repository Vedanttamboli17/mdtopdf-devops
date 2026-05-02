package services

import (
    "context"
    "fmt"
    "io"
    "os"
    "time"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/s3"
)

var s3Client *s3.Client

func InitS3() error {
    cfg, err := config.LoadDefaultConfig(context.TODO(),
        config.WithRegion(os.Getenv("AWS_REGION")),
    )
    if err != nil {
        return fmt.Errorf("failed to load AWS config: %w", err)
    }
    s3Client = s3.NewFromConfig(cfg)
    return nil
}

func UploadMarkdown(key string, file io.Reader) error {
    _, err := s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
        Bucket:      aws.String(os.Getenv("S3_BUCKET")),
        Key:         aws.String(key),
        Body:        file,
        ContentType: aws.String("text/markdown"),
    })
    return err
}

func GeneratePresignedURL(key string) (string, error) {
    presignClient := s3.NewPresignClient(s3Client)
    req, err := presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
        Bucket: aws.String(os.Getenv("S3_BUCKET")),
        Key:    aws.String(key),
    }, s3.WithPresignExpires(10*time.Minute))
    if err != nil {
        return "", fmt.Errorf("presign failed: %w", err)
    }
    return req.URL, nil
}