package services

import (
    "context"
    "fmt"
    "os"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
    "mdtopdf/api/models"
)

var dbClient *dynamodb.Client

func InitDynamoDB() error {
    cfg, err := config.LoadDefaultConfig(context.TODO(),
        config.WithRegion(os.Getenv("AWS_REGION")),
    )
    if err != nil {
        return fmt.Errorf("failed to load AWS config: %w", err)
    }
    dbClient = dynamodb.NewFromConfig(cfg)
    return nil
}

func CreateJob(job models.Job) error {
    item, err := attributevalue.MarshalMap(job)
    if err != nil {
        return fmt.Errorf("marshal failed: %w", err)
    }
    _, err = dbClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
        TableName: aws.String(os.Getenv("DYNAMODB_TABLE")),
        Item:      item,
    })
    return err
}

func GetJob(id string) (*models.Job, error) {
    result, err := dbClient.GetItem(context.TODO(), &dynamodb.GetItemInput{
        TableName: aws.String(os.Getenv("DYNAMODB_TABLE")),
        Key: map[string]types.AttributeValue{
            "id": &types.AttributeValueMemberS{Value: id},
        },
    })
    if err != nil {
        return nil, err
    }
    if result.Item == nil {
        return nil, nil
    }
    var job models.Job
    if err := attributevalue.UnmarshalMap(result.Item, &job); err != nil {
        return nil, err
    }
    return &job, nil
}