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

func UpdateThumbnailURL(id, thumbnailURL string) error {
	_, err := dbClient.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName: aws.String(os.Getenv("DYNAMODB_TABLE")),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
		UpdateExpression: aws.String("SET thumbnail_url = :t"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":t": &types.AttributeValueMemberS{Value: thumbnailURL},
		},
	})
	return err
}

// GetJob fetches a job record by id (the jobs table primary key).
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

// GetTenantByID looks up a tenant by tenant_id. The tenants table is keyed by
// api_key, so this is a filtered Scan — acceptable at this scale (few tenants),
// and only invoked when a job has a webhook to deliver.
func GetTenantByID(tenantID string) (*models.Tenant, error) {
	table := os.Getenv("TENANTS_TABLE")
	if table == "" {
		return nil, fmt.Errorf("TENANTS_TABLE env var not set")
	}
	out, err := dbClient.Scan(context.TODO(), &dynamodb.ScanInput{
		TableName:        aws.String(table),
		FilterExpression: aws.String("tenant_id = :tid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":tid": &types.AttributeValueMemberS{Value: tenantID},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(out.Items) == 0 {
		return nil, nil
	}
	var tenant models.Tenant
	if err := attributevalue.UnmarshalMap(out.Items[0], &tenant); err != nil {
		return nil, err
	}
	return &tenant, nil
}
