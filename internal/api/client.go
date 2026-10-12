package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	baseURL = strings.TrimRight(baseURL, "/")
	return &Client{
		baseURL: baseURL,
		http: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *Client) MarkStarted(buildID string) (bool, error) {
	status := ""
	err := c.retry(func() error {
		var err error
		status, err = c.post(fmt.Sprintf("/internal/builds/%s/RUNNING", buildID), nil)
		return err
	})
	return status == "RUNNING", err
}

func (c *Client) MarkCompleted(buildID string, duration int64) error {
	payload := map[string]interface{}{
		"duration": duration,
	}

	return c.retry(func() error {
		_, err := c.post(fmt.Sprintf("/internal/builds/%s/COMPLETED", buildID), payload)
		return err
	})
}

func (c *Client) MarkFailed(buildID string, duration int64, message string) error {
	payload := map[string]interface{}{"duration": duration, "error": message}

	return c.retry(func() error {
		_, err := c.post(fmt.Sprintf("/internal/builds/%s/FAILED", buildID), payload)
		return err
	})
}

func (c *Client) post(path string, payload interface{}) (string, error) {
	var body bytes.Buffer
	if payload != nil {
		_ = json.NewEncoder(&body).Encode(payload)
	}

	resp, err := c.http.Post(c.baseURL+path, "application/json", &body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Status, nil
}

func (c *Client) retry(fn func() error) error {
	for i := 0; i < 3; i++ {
		if err := fn(); err != nil {
			if i < 2 {
				time.Sleep(time.Duration(i+1) * time.Second)
			}
			continue
		}
		return nil
	}
	return fmt.Errorf("request failed after retries")
}
