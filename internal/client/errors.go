package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// Sentinel errors for common HTTP status codes.
var (
	ErrBadRequest          = errors.New("bad request")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrNotFound            = errors.New("not found")
	ErrInternalServerError = errors.New("internal server error")
	ErrServiceUnavailable  = errors.New("service unavailable")
)

// ClientError represents an error from the server with status code and message.
type ClientError struct {
	StatusCode int
	Message    string
	Err        error
}

// Error implements the error interface.
func (e *ClientError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("HTTP %d: %s: %v", e.StatusCode, e.Message, e.Err)
	}
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

// Unwrap returns the wrapped error.
func (e *ClientError) Unwrap() error {
	return e.Err
}

// parseErrorResponse attempts to parse JSON error response from server.
func parseErrorResponse(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &ClientError{
			StatusCode: resp.StatusCode,
			Message:    "failed to read response body",
			Err:        err,
		}
	}

	// Try to parse JSON error
	var errResp struct {
		Error string `json:"error"`
	}

	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != "" {
		return statusCodeToError(resp.StatusCode, errResp.Error)
	}

	// Fallback to status text if JSON parsing fails
	return statusCodeToError(resp.StatusCode, string(body))
}

// statusCodeToError converts HTTP status code to appropriate error.
func statusCodeToError(statusCode int, message string) error {
	// Wrap with sentinel error for easy error checking
	var sentinelErr error
	switch statusCode {
	case http.StatusBadRequest:
		sentinelErr = ErrBadRequest
	case http.StatusUnauthorized:
		sentinelErr = ErrUnauthorized
	case http.StatusNotFound:
		sentinelErr = ErrNotFound
	case http.StatusInternalServerError:
		sentinelErr = ErrInternalServerError
	case http.StatusServiceUnavailable:
		sentinelErr = ErrServiceUnavailable
	}

	return &ClientError{
		StatusCode: statusCode,
		Message:    message,
		Err:        sentinelErr,
	}
}
