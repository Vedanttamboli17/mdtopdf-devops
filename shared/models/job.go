package models

type Job struct {
	ID          string `json:"id" dynamodbav:"id"`
	TenantID    string `json:"tenant_id" dynamodbav:"tenant_id"`
	JobType     string `json:"job_type" dynamodbav:"job_type"`
	Status      string `json:"status" dynamodbav:"status"`
	Format      string `json:"format" dynamodbav:"format"`
	Priority    int    `json:"priority" dynamodbav:"priority"`
	InputURL    string `json:"input_url" dynamodbav:"input_url"`
	OutputURL   string `json:"output_url" dynamodbav:"output_url"`
	WebhookURL  string `json:"webhook_url" dynamodbav:"webhook_url"`
	RetryCount  int    `json:"retry_count" dynamodbav:"retry_count"`
	CreatedAt   string `json:"created_at" dynamodbav:"created_at"`
	CompletedAt  string `json:"completed_at" dynamodbav:"completed_at"`
	ThumbnailURL string `json:"thumbnail_url" dynamodbav:"thumbnail_url"`
}
