package client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/t-eckert/fave/internal"
)

// Client is an HTTP client for the Fave bookmark API.
type Client struct {
	config Config
	http   *http.Client
}

// New creates a new Client with the given configuration.
func New(config Config) (*Client, error) {
	// Create custom transport with connection pooling
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   config.DialTimeout,
			KeepAlive: config.KeepAlive,
		}).DialContext,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return &Client{
		config: config,
		http: &http.Client{
			Transport: transport,
			Timeout:   config.Timeout,
		},
	}, nil
}

// Close cleans up client resources.
func (c *Client) Close() {
	c.http.CloseIdleConnections()
}

// Add creates a new bookmark and returns its ID.
func (c *Client) Add(bookmark internal.Bookmark) (int, error) {
	body, err := json.Marshal(bookmark)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal bookmark: %w", err)
	}

	var result struct {
		ID int `json:"id"`
	}

	err = c.doWithRetry("POST", "/bookmarks", body, http.StatusCreated, &result)
	if err != nil {
		return 0, fmt.Errorf("add bookmark: %w", err)
	}

	return result.ID, nil
}

// List returns all bookmarks.
func (c *Client) List() (map[int]internal.Bookmark, error) {
	var bookmarks map[int]internal.Bookmark

	err := c.doWithRetry("GET", "/bookmarks", nil, http.StatusOK, &bookmarks)
	if err != nil {
		return nil, fmt.Errorf("list bookmarks: %w", err)
	}

	return bookmarks, nil
}

// Get retrieves a bookmark by ID.
func (c *Client) Get(id int) (*internal.Bookmark, error) {
	var bookmark internal.Bookmark

	path := fmt.Sprintf("/bookmarks/%d", id)
	err := c.doWithRetry("GET", path, nil, http.StatusOK, &bookmark)
	if err != nil {
		return nil, fmt.Errorf("get bookmark: %w", err)
	}

	return &bookmark, nil
}

// Update updates an existing bookmark.
func (c *Client) Update(id int, bookmark internal.Bookmark) error {
	body, err := json.Marshal(bookmark)
	if err != nil {
		return fmt.Errorf("failed to marshal bookmark: %w", err)
	}

	path := fmt.Sprintf("/bookmarks/%d", id)
	var result struct {
		ID int `json:"id"`
	}

	err = c.doWithRetry("PUT", path, body, http.StatusOK, &result)
	if err != nil {
		return fmt.Errorf("update bookmark: %w", err)
	}

	return nil
}

// Delete removes a bookmark by ID.
func (c *Client) Delete(id int) error {
	path := fmt.Sprintf("/bookmarks/%d", id)
	var result struct {
		ID int `json:"id"`
	}

	err := c.doWithRetry("DELETE", path, nil, http.StatusOK, &result)
	if err != nil {
		return fmt.Errorf("delete bookmark: %w", err)
	}

	return nil
}

// Health checks if the server is healthy.
func (c *Client) Health() error {
	var result struct {
		Status string `json:"status"`
	}

	err := c.doWithRetry("GET", "/health", nil, http.StatusOK, &result)
	if err != nil {
		return fmt.Errorf("health check: %w", err)
	}

	if result.Status != "healthy" {
		return fmt.Errorf("server unhealthy: %s", result.Status)
	}

	return nil
}

// doWithRetry performs an HTTP request with retry logic and exponential backoff.
func (c *Client) doWithRetry(method, path string, body []byte, expectedStatus int, result any) error {
	var lastErr error

	for attempt := 0; attempt <= c.config.RetryAttempts; attempt++ {
		// Calculate delay for this attempt (exponential backoff)
		if attempt > 0 {
			delay := time.Duration(attempt) * c.config.RetryDelay
			if delay > c.config.RetryMaxDelay {
				delay = c.config.RetryMaxDelay
			}
			time.Sleep(delay)
		}

		// Perform request
		err := c.doRequest(method, path, body, expectedStatus, result)
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry on client errors (4xx except 429)
		if clientErr, ok := lastErr.(*ClientError); ok {
			if clientErr.StatusCode >= 400 && clientErr.StatusCode < 500 && clientErr.StatusCode != 429 {
				return lastErr
			}
		}
	}

	return lastErr
}

// doRequest performs a single HTTP request without retries.
func (c *Client) doRequest(method, path string, body []byte, expectedStatus int, result any) error {
	url := c.config.Host + path

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Add authentication if password is configured
	if c.config.Password != "" {
		c.addAuth(req)
	}

	// Execute request
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != expectedStatus {
		return parseErrorResponse(resp)
	}

	// Parse response body if result is provided
	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// addAuth adds HTTP Basic Authentication to the request.
func (c *Client) addAuth(req *http.Request) {
	// Username can be anything; only password matters
	credentials := "user:" + c.config.Password
	encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
	req.Header.Set("Authorization", "Basic "+encoded)
}
