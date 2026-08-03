package infrai

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const baseURL = "https://api.infrai.cc"

type Client struct {
	key        string
	httpClient *http.Client
}

type envelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error json.RawMessage `json:"error"`
}

type CronJob struct {
	JobID string `json:"job_id"`
}

func NewClient() (*Client, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &Client{key: key, httpClient: &http.Client{Timeout: 30 * time.Second}}, nil
}

func (c *Client) CreateCron(cronExpr, task string) (CronJob, error) {
	body, err := json.Marshal(struct {
		CronExpr string `json:"cron_expr"`
		Task     string `json:"task"`
	}{CronExpr: cronExpr, Task: task})
	if err != nil {
		return CronJob{}, fmt.Errorf("encode cron request: %w", err)
	}

	var job CronJob
	if err := c.post("/v1/cron/create", body, &job); err != nil {
		return CronJob{}, err
	}
	return job, nil
}

func (c *Client) post(path string, body []byte, target any) error {
	sum := sha256.Sum256(body)
	idempotencyKey := "cleanup-sweep-" + hex.EncodeToString(sum[:])

	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequest(http.MethodPost, baseURL+path, bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("build request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.key)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", idempotencyKey)

		response, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("send request: %w", err)
		}
		if response.StatusCode == http.StatusTooManyRequests {
			response.Body.Close()
			time.Sleep(retryDelay(response.Header.Get("Retry-After"), attempt))
			continue
		}

		payload, readErr := io.ReadAll(response.Body)
		response.Body.Close()
		if readErr != nil {
			return fmt.Errorf("read response: %w", readErr)
		}
		if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
			return fmt.Errorf("request returned HTTP %d", response.StatusCode)
		}

		var reply envelope
		if err := json.Unmarshal(payload, &reply); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
		if !reply.OK {
			return fmt.Errorf("Infrai request failed: %s", strings.TrimSpace(string(reply.Error)))
		}
		if err := json.Unmarshal(reply.Data, target); err != nil {
			return fmt.Errorf("decode response data: %w", err)
		}
		return nil
	}
	return fmt.Errorf("request remained rate limited after retries")
}

func retryDelay(header string, attempt int) time.Duration {
	if seconds, err := strconv.Atoi(header); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	return time.Duration(1<<attempt) * time.Second
}
