package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *Client) MarkStarted(buildID string) error {
	return c.retry(func() error {
		return c.post(fmt.Sprintf("http://localhost:8080/internal/builds/%s/running", buildID), nil)
	})
}

func (c *Client) MarkCompleted(buildID string, duration int64) error {
	payload := map[string]interface{}{
		"duration": duration,
	}

	return c.retry(func() error {
		return c.post(fmt.Sprintf("http://localhost:8080/internal/builds/%s/completed", buildID), payload)
	})
}

func (c *Client) MarkFailed(buildID string, duration int64) error {
	payload := map[string]string{"error": "build faled"}

	return c.retry(func() error {
		return c.post(fmt.Sprintf("http://localhost:8080/internal/builds/%s/failed", buildID), payload)
	})
}

func (c *Client) post(path string, payload interface{}) error {
	var body bytes.Buffer
	if payload != nil {
		_ = json.NewEncoder(&body).Encode(payload)
	}

	resp, err := c.http.Post(c.baseURL+path, "application/json", &body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) retry(fn func() error) error {
	for i := 0; i < 3; i++ {
		if err := fn(); err != nil {
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}
		return nil
	}
	return fmt.Errorf("request failed after retries")
}
