package common

import (
	"fmt"
	"io"
	"net/http"
)

// HTTPResponse represents a standardized HTTP response
type HTTPResponse struct {
	StatusCode int
	Body       []byte
	Error      error
}

// MakeHTTPRequest makes an HTTP request and returns a standardized response
func MakeHTTPRequest(req *http.Request) *HTTPResponse {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return &HTTPResponse{
			Error: fmt.Errorf("error making request: %w", err),
		}
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("error closing response body: %v\n", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &HTTPResponse{
			Error: fmt.Errorf("error reading response body: %w", err),
		}
	}

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       body,
	}
}
