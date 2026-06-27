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

var (
	tenantClient *dynamodb.Client
	tenantsTable string
)

func InitTenantTable() error {
	tenantsTable = os.Getenv("TENANTS_TABLE")
	if tenantsTable == "" {
		return fmt.Errorf("TENANTS_TABLE env var not set")
	}
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(os.Getenv("AWS_REGION")),
	)
	if err != nil {
		return fmt.Errorf("tenant DynamoDB config error: %w", err)
	}
	tenantClient = dynamodb.NewFromConfig(cfg)
	return nil
}

func GetTenantByAPIKey(apiKey string) (*models.Tenant, error) {
	result, err := tenantClient.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(tenantsTable),
		Key: map[string]types.AttributeValue{
			"api_key": &types.AttributeValueMemberS{Value: apiKey},
		},
	})
	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, nil
	}
	var tenant models.Tenant
	if err := attributevalue.UnmarshalMap(result.Item, &tenant); err != nil {
		return nil, err
	}
	return &tenant, nil
}

func CreateTenant(tenant models.Tenant) error {
	item, err := attributevalue.MarshalMap(tenant)
	if err != nil {
		return fmt.Errorf("marshal failed: %w", err)
	}
	_, err = tenantClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(tenantsTable),
		Item:      item,
	})
	return err
}

// IncrementJobCount atomically adds 1 to jobs_this_month.
// Takes apiKey because that is the tenants table primary key.
func IncrementJobCount(apiKey string) error {
	_, err := tenantClient.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName: aws.String(tenantsTable),
		Key: map[string]types.AttributeValue{
			"api_key": &types.AttributeValueMemberS{Value: apiKey},
		},
		UpdateExpression: aws.String("ADD jobs_this_month :one"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":one": &types.AttributeValueMemberN{Value: "1"},
		},
	})
	return err
}
