package services

import (
    "context"
    "fmt"
    "io"
    "os"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/s3"
    "time"
)

var s3Client *s3.Client

func InitS3() error {
    cfg, err := config.LoadDefaultConfig(context.TODO(),
        config.WithRegion(os.Getenv("AWS_REGION")),
    )
    if err != nil {
        return fmt.Errorf("S3 config error: %w", err)
    }
    s3Client = s3.NewFromConfig(cfg)
    return nil
}

func DownloadFile(key string) ([]byte, error) {
    result, err := s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
        Bucket: aws.String(os.Getenv("S3_BUCKET")),
        Key:    aws.String(key),
    })
    if err != nil {
        return nil, fmt.Errorf("S3 GetObject failed: %w", err)
    }
    defer result.Body.Close()
    return io.ReadAll(result.Body)
}

func UploadPDF(key string, filePath string) error {
    f, err := os.Open(filePath)
    if err != nil {
        return fmt.Errorf("cannot open PDF: %w", err)
    }
    defer f.Close()

    _, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
        Bucket:      aws.String(os.Getenv("S3_BUCKET")),
        Key:         aws.String(key),
        Body:        f,
        ContentType: aws.String("application/pdf"),
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
        return "", err
    }
    return req.URL, nil
}