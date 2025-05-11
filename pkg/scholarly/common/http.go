package common

import (
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
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
			log.Warn().Err(err).Msg("error closing response body")
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &HTTPResponse{
			Error: fmt.Errorf("error reading response body: %w", err),
		}
	}

	if resp.StatusCode != http.StatusOK {
		log.Warn().Int("status_code", resp.StatusCode).Str("url", req.URL.String()).Msg("HTTP request failed")
		log.Debug().Str("url", req.URL.String()).Bytes("response_body", body).Msg("HTTP request failed body")
	}

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       body,
	}
}
