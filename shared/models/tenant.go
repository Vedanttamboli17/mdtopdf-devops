package models

type Tenant struct {
	APIKey          string `json:"api_key" dynamodbav:"api_key"`
	TenantID        string `json:"tenant_id" dynamodbav:"tenant_id"`
	Name            string `json:"name" dynamodbav:"name"`
	Plan            string `json:"plan" dynamodbav:"plan"`
	RateLimit       int    `json:"rate_limit" dynamodbav:"rate_limit"`
	MonthlyQuota    int    `json:"monthly_quota" dynamodbav:"monthly_quota"`
	JobsThisMonth   int    `json:"jobs_this_month" dynamodbav:"jobs_this_month"`
	QuotaResetMonth string `json:"quota_reset_month" dynamodbav:"quota_reset_month"`
	WebhookSecret   string `json:"webhook_secret" dynamodbav:"webhook_secret"`
	CreatedAt       string `json:"created_at" dynamodbav:"created_at"`
	Active          bool   `json:"active" dynamodbav:"active"`
}
