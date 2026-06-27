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
    "github.com/docplatform/shared/models"
)

var dbClient *dynamodb.Client

func InitDynamoDB() error {
    cfg, err := config.LoadDefaultConfig(context.TODO(),
        config.WithRegion(os.Getenv("AWS_REGION")),
    )
    if err != nil {
        return fmt.Errorf("DynamoDB config error: %w", err)
    }
    dbClient = dynamodb.NewFromConfig(cfg)
    return nil
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

func UpdateJobCompleted(id, outputURL string) error {
    _, err := dbClient.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
        TableName: aws.String(os.Getenv("DYNAMODB_TABLE")),
        Key: map[string]types.AttributeValue{
            "id": &types.AttributeValueMemberS{Value: id},
        },
        UpdateExpression: aws.String("SET #s = :s, output_url = :u"),
        ExpressionAttributeNames: map[string]string{"#s": "status"},
        ExpressionAttributeValues: map[string]types.AttributeValue{
            ":s": &types.AttributeValueMemberS{Value: "completed"},
            ":u": &types.AttributeValueMemberS{Value: outputURL},
        },
    })
    return err
}

func UpdateJobFailed(id string) error {
    _, err := dbClient.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
        TableName: aws.String(os.Getenv("DYNAMODB_TABLE")),
        Key: map[string]types.AttributeValue{
            "id": &types.AttributeValueMemberS{Value: id},
        },
        UpdateExpression: aws.String("SET #s = :s"),
        ExpressionAttributeNames: map[string]string{"#s": "status"},
        ExpressionAttributeValues: map[string]types.AttributeValue{
            ":s": &types.AttributeValueMemberS{Value: "failed"},
        },
    })
    return err
}

func IncrementRetry(id string, current int) error {
    _, err := dbClient.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
        TableName: aws.String(os.Getenv("DYNAMODB_TABLE")),
        Key: map[string]types.AttributeValue{
            "id": &types.AttributeValueMemberS{Value: id},
        },
        UpdateExpression: aws.String("SET retry_count = :rc"),
        ExpressionAttributeValues: map[string]types.AttributeValue{
            ":rc": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", current+1)},
        },
    })
    return err
}
