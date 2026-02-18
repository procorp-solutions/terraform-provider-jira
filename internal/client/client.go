package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Client is a thin wrapper around the JIRA Cloud REST API.
type Client struct {
	BaseURL    string
	Email      string
	APIToken   string
	HTTPClient *http.Client
}

// NewClient creates a new JIRA API client.
func NewClient(baseURL, email, apiToken string) *Client {
	return &Client{
		BaseURL:  baseURL,
		Email:    email,
		APIToken: apiToken,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// APIError represents an error returned by the JIRA API.
type APIError struct {
	StatusCode    int
	ErrorMessages []string          `json:"errorMessages"`
	Errors        map[string]string `json:"errors"`
}

func (e *APIError) Error() string {
	if len(e.ErrorMessages) > 0 {
		return fmt.Sprintf("JIRA API error (HTTP %d): %s", e.StatusCode, e.ErrorMessages[0])
	}
	if len(e.Errors) > 0 {
		for k, v := range e.Errors {
			return fmt.Sprintf("JIRA API error (HTTP %d): %s: %s", e.StatusCode, k, v)
		}
	}
	return fmt.Sprintf("JIRA API error (HTTP %d)", e.StatusCode)
}

// doRequest executes an HTTP request with authentication and error handling.
func (c *Client) doRequest(method, path string, body interface{}, result interface{}) error {
	fullURL := c.BaseURL + path

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, fullURL, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.Email, c.APIToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Handle rate limiting
	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := resp.Header.Get("Retry-After")
		if seconds, err := strconv.Atoi(retryAfter); err == nil {
			time.Sleep(time.Duration(seconds) * time.Second)
			return c.doRequest(method, path, body, result)
		}
		return fmt.Errorf("rate limited by JIRA API, retry after: %s", retryAfter)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := &APIError{StatusCode: resp.StatusCode}
		if len(respBody) > 0 {
			_ = json.Unmarshal(respBody, apiErr)
		}
		if apiErr.Error() == fmt.Sprintf("JIRA API error (HTTP %d)", resp.StatusCode) && len(respBody) > 0 {
			return fmt.Errorf("JIRA API error (HTTP %d): %s", resp.StatusCode, string(respBody))
		}
		return apiErr
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// Get sends a GET request.
func (c *Client) Get(path string, result interface{}) error {
	return c.doRequest(http.MethodGet, path, nil, result)
}

// Post sends a POST request.
func (c *Client) Post(path string, body interface{}, result interface{}) error {
	return c.doRequest(http.MethodPost, path, body, result)
}

// Put sends a PUT request.
func (c *Client) Put(path string, body interface{}, result interface{}) error {
	return c.doRequest(http.MethodPut, path, body, result)
}

// Delete sends a DELETE request.
func (c *Client) Delete(path string) error {
	return c.doRequest(http.MethodDelete, path, nil, nil)
}

// DeleteWithQuery sends a DELETE request with query parameters.
func (c *Client) DeleteWithQuery(path string, params url.Values) error {
	if len(params) > 0 {
		path = path + "?" + params.Encode()
	}
	return c.doRequest(http.MethodDelete, path, nil, nil)
}

// IsNotFound checks if the error is a 404 Not Found.
func IsNotFound(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == http.StatusNotFound
	}
	return false
}
