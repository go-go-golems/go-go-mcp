package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// HTTPRequest represents a JavaScript HTTP request configuration
type HTTPRequest struct {
	URL     string                 `json:"url"`
	Method  string                 `json:"method"`
	Headers map[string]string      `json:"headers"`
	Body    interface{}            `json:"body"`
	Query   map[string]interface{} `json:"query"`
	Timeout int                    `json:"timeout"` // seconds
}

// HTTPResponse represents a JavaScript HTTP response
type HTTPResponse struct {
	Status     int               `json:"status"`
	StatusText string            `json:"statusText"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	JSON       interface{}       `json:"json"`
	OK         bool              `json:"ok"`
	URL        string            `json:"url"`
	Error      string            `json:"error,omitempty"`
}

// setupHTTPBindings configures HTTP request bindings for the JavaScript runtime
func (e *Engine) setupHTTPBindings() {
	// HTTP client with default timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Main fetch function (modern browser-like API)
	if err := e.rt.Set("fetch", func(urlOrOptions interface{}, options ...interface{}) map[string]interface{} {
		return e.jsFetch(client, urlOrOptions, options...)
	}); err != nil {
		log.Error().Err(err).Msg("Failed to set fetch binding")
	}

	// HTTP utility object with method shortcuts
	if err := e.rt.Set("HTTP", map[string]interface{}{
		"get": func(url string, options ...interface{}) map[string]interface{} {
			return e.jsHTTPMethod(client, "GET", url, options...)
		},
		"post": func(url string, options ...interface{}) map[string]interface{} {
			return e.jsHTTPMethod(client, "POST", url, options...)
		},
		"put": func(url string, options ...interface{}) map[string]interface{} {
			return e.jsHTTPMethod(client, "PUT", url, options...)
		},
		"delete": func(url string, options ...interface{}) map[string]interface{} {
			return e.jsHTTPMethod(client, "DELETE", url, options...)
		},
		"patch": func(url string, options ...interface{}) map[string]interface{} {
			return e.jsHTTPMethod(client, "PATCH", url, options...)
		},
		"head": func(url string, options ...interface{}) map[string]interface{} {
			return e.jsHTTPMethod(client, "HEAD", url, options...)
		},
	}); err != nil {
		log.Error().Err(err).Msg("Failed to set HTTP utility binding")
	}

	log.Debug().Msg("HTTP request bindings configured")
}

// jsFetch implements a fetch-like API for JavaScript
func (e *Engine) jsFetch(client *http.Client, urlOrOptions interface{}, options ...interface{}) map[string]interface{} {
	var req HTTPRequest

	// Parse arguments (fetch can be called as fetch(url) or fetch(url, options) or fetch(options))
	switch v := urlOrOptions.(type) {
	case string:
		req.URL = v
		req.Method = "GET"
		if len(options) > 0 {
			if opts, ok := options[0].(map[string]interface{}); ok {
				e.parseHTTPOptions(&req, opts)
			}
		}
	case map[string]interface{}:
		e.parseHTTPOptions(&req, v)
		if req.Method == "" {
			req.Method = "GET"
		}
	default:
		return map[string]interface{}{
			"error": "Invalid fetch arguments",
			"ok":    false,
		}
	}

	return e.executeHTTPRequest(client, &req)
}

// jsHTTPMethod implements HTTP method shortcuts (HTTP.get, HTTP.post, etc.)
func (e *Engine) jsHTTPMethod(client *http.Client, method, url string, options ...interface{}) map[string]interface{} {
	req := HTTPRequest{
		URL:    url,
		Method: method,
	}

	if len(options) > 0 {
		if opts, ok := options[0].(map[string]interface{}); ok {
			e.parseHTTPOptions(&req, opts)
		}
	}

	return e.executeHTTPRequest(client, &req)
}

// parseHTTPOptions parses JavaScript options object into HTTPRequest
func (e *Engine) parseHTTPOptions(req *HTTPRequest, options map[string]interface{}) {
	if url, ok := options["url"].(string); ok {
		req.URL = url
	}
	if method, ok := options["method"].(string); ok {
		req.Method = strings.ToUpper(method)
	}
	if headers, ok := options["headers"].(map[string]interface{}); ok {
		req.Headers = make(map[string]string)
		for k, v := range headers {
			req.Headers[k] = fmt.Sprint(v)
		}
	}
	if body := options["body"]; body != nil {
		req.Body = body
	}
	if query, ok := options["query"].(map[string]interface{}); ok {
		req.Query = query
	}
	if timeout, ok := options["timeout"].(float64); ok {
		req.Timeout = int(timeout)
	}
}

// executeHTTPRequest performs the actual HTTP request
func (e *Engine) executeHTTPRequest(client *http.Client, req *HTTPRequest) map[string]interface{} {
	log.Debug().Str("method", req.Method).Str("url", req.URL).Msg("Executing HTTP request")

	// Build URL with query parameters
	finalURL := req.URL
	if len(req.Query) > 0 {
		u, err := url.Parse(req.URL)
		if err != nil {
			return map[string]interface{}{
				"error": fmt.Sprintf("Invalid URL: %v", err),
				"ok":    false,
			}
		}

		values := u.Query()
		for k, v := range req.Query {
			switch val := v.(type) {
			case []interface{}:
				for _, item := range val {
					values.Add(k, fmt.Sprint(item))
				}
			default:
				values.Set(k, fmt.Sprint(v))
			}
		}
		u.RawQuery = values.Encode()
		finalURL = u.String()
	}

	// Prepare request body
	var bodyReader io.Reader
	var contentType string
	if req.Body != nil {
		switch body := req.Body.(type) {
		case string:
			bodyReader = strings.NewReader(body)
			contentType = "text/plain"
		case map[string]interface{}:
			jsonData, err := json.Marshal(body)
			if err != nil {
				return map[string]interface{}{
					"error": fmt.Sprintf("JSON encoding error: %v", err),
					"ok":    false,
				}
			}
			bodyReader = bytes.NewReader(jsonData)
			contentType = "application/json"
		default:
			// Try to convert to JSON
			jsonData, err := json.Marshal(body)
			if err != nil {
				bodyReader = strings.NewReader(fmt.Sprint(body))
				contentType = "text/plain"
			} else {
				bodyReader = bytes.NewReader(jsonData)
				contentType = "application/json"
			}
		}
	}

	// Create HTTP request
	httpReq, err := http.NewRequest(req.Method, finalURL, bodyReader)
	if err != nil {
		return map[string]interface{}{
			"error": fmt.Sprintf("Request creation error: %v", err),
			"ok":    false,
		}
	}

	// Set headers
	if contentType != "" && (req.Headers == nil || req.Headers["Content-Type"] == "") {
		httpReq.Header.Set("Content-Type", contentType)
	}
	if req.Headers != nil {
		for k, v := range req.Headers {
			httpReq.Header.Set(k, v)
		}
	}

	// Set timeout if specified
	if req.Timeout > 0 {
		client = &http.Client{
			Timeout: time.Duration(req.Timeout) * time.Second,
		}
	}

	// Execute request
	resp, err := client.Do(httpReq)
	if err != nil {
		log.Error().Err(err).Str("url", finalURL).Msg("HTTP request failed")
		return map[string]interface{}{
			"error": fmt.Sprintf("Request failed: %v", err),
			"ok":    false,
			"url":   finalURL,
		}
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read response body")
		return map[string]interface{}{
			"error": fmt.Sprintf("Failed to read response: %v", err),
			"ok":    false,
			"url":   finalURL,
		}
	}

	// Convert headers to map
	headers := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[k] = v[0] // Take first value for simplicity
		}
	}

	bodyStr := string(bodyBytes)
	response := map[string]interface{}{
		"status":     resp.StatusCode,
		"statusText": resp.Status,
		"headers":    headers,
		"body":       bodyStr,
		"ok":         resp.StatusCode >= 200 && resp.StatusCode < 300,
		"url":        finalURL,
	}

	// Try to parse JSON if content type suggests it
	contentType = resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/json") {
		var jsonData interface{}
		if err := json.Unmarshal(bodyBytes, &jsonData); err == nil {
			response["json"] = jsonData
		}
	}

	log.Debug().Int("status", resp.StatusCode).Str("url", finalURL).Msg("HTTP request completed")
	return response
}
