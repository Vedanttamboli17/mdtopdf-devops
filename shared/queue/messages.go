package queue

type JobMessage struct {
	FileID     string `json:"file_id"`
	S3Key      string `json:"s3_key"`
	JobType    string `json:"job_type"`
	TenantID   string `json:"tenant_id"`
	WebhookURL string `json:"webhook_url"`
	Priority   int    `json:"priority"`
}

type ThumbnailMessage struct {
	FileID   string `json:"file_id"`
	PDFS3Key string `json:"pdf_s3_key"`
	TenantID string `json:"tenant_id"`
}

// NotificationMessage is published to NOTIFICATION_SQS_QUEUE_URL when a job
// reaches a terminal state. The API's notification poller reads these and
// pushes them to waiting SSE clients.
type NotificationMessage struct {
	FileID       string `json:"file_id"`
	Status       string `json:"status"`
	DownloadURL  string `json:"download_url,omitempty"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	TenantID     string `json:"tenant_id"`
}
