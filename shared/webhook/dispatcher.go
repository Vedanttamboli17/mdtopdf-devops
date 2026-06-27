package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"
)

// WebhookPayload is the JSON body POSTed to a tenant's webhook URL when a job
// reaches a terminal state.
type WebhookPayload struct {
	FileID       string `json:"file_id"`
	Status       string `json:"status"`
	JobType      string `json:"job_type"`
	DownloadURL  string `json:"download_url,omitempty"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	CompletedAt  string `json:"completed_at"`
	Error        string `json:"error,omitempty"`
}

// backoffs are the waits before each retry. Their count defines the number of
// retries (initial attempt + len(backoffs) total attempts).
var backoffs = []time.Duration{10 * time.Second, 30 * time.Second, 90 * time.Second}

// requestTimeout reads WEBHOOK_TIMEOUT_SECONDS (default 10s).
func requestTimeout() time.Duration {
	if v := os.Getenv("WEBHOOK_TIMEOUT_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return time.Duration(n) * time.Second
		}
	}
	return 10 * time.Second
}

// Dispatch POSTs the payload to webhookURL with an HMAC-SHA256 signature over
// the JSON body in the X-Signature header. It retries up to len(backoffs) times
// with exponential backoff. A delivery failure is returned (and logged by the
// caller) but does NOT mark the job failed — the job succeeded, only the
// callback did not land.
func Dispatch(webhookURL string, secret string, payload WebhookPayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal webhook payload: %w", err)
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	client := &http.Client{Timeout: requestTimeout()}

	var lastErr error
	for attempt := 0; attempt <= len(backoffs); attempt++ {
		if attempt > 0 {
			time.Sleep(backoffs[attempt-1])
		}
		lastErr = deliver(client, webhookURL, signature, body)
		if lastErr == nil {
			if attempt > 0 {
				slog.Info("Webhook delivered after retry", "file_id", payload.FileID, "attempt", attempt+1)
			}
			return nil
		}
		slog.Warn("Webhook delivery attempt failed",
			"file_id", payload.FileID, "attempt", attempt+1, "error", lastErr)
	}
	return fmt.Errorf("webhook delivery failed after %d attempts: %w", len(backoffs)+1, lastErr)
}

func deliver(client *http.Client, url, signature string, body []byte) error {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Signature", signature)
	req.Header.Set("User-Agent", "docplatform-webhook/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body) // drain so the connection can be reused

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("non-2xx response: %d", resp.StatusCode)
	}
	return nil
}
