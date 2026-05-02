package models

type Job struct {
    ID         string `json:"id" dynamodbav:"id"`
    Status     string `json:"status" dynamodbav:"status"`
    InputURL   string `json:"input_url" dynamodbav:"input_url"`
    OutputURL  string `json:"output_url" dynamodbav:"output_url"`
    RetryCount int    `json:"retry_count" dynamodbav:"retry_count"`
}