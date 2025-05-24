package engine

import (
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// RequestLog represents a single request and its associated logs
type RequestLog struct {
	ID        string                 `json:"id"`
	Method    string                 `json:"method"`
	Path      string                 `json:"path"`
	URL       string                 `json:"url"`
	Status    int                    `json:"status"`
	StartTime time.Time              `json:"startTime"`
	EndTime   time.Time              `json:"endTime"`
	Duration  time.Duration          `json:"duration"`
	Headers   map[string]interface{} `json:"headers"`
	Query     map[string]interface{} `json:"query"`
	Body      string                 `json:"body,omitempty"`
	Response  string                 `json:"response,omitempty"`
	Logs      []LogEntry             `json:"logs"`
	Error     string                 `json:"error,omitempty"`
	RemoteIP  string                 `json:"remoteIP"`
}

// LogEntry represents a single log message during request processing
type LogEntry struct {
	Timestamp time.Time   `json:"timestamp"`
	Level     string      `json:"level"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
}

// RequestLogger manages request logging and provides real-time access
type RequestLogger struct {
	mu       sync.RWMutex
	requests map[string]*RequestLog
	maxLogs  int
	order    []string // Keep track of insertion order for LRU
}

// NewRequestLogger creates a new request logger
func NewRequestLogger(maxLogs int) *RequestLogger {
	if maxLogs <= 0 {
		maxLogs = 100 // Default to keeping last 100 requests
	}

	return &RequestLogger{
		requests: make(map[string]*RequestLog),
		maxLogs:  maxLogs,
		order:    make([]string, 0),
	}
}

// StartRequest creates a new request log entry
func (rl *RequestLogger) StartRequest(r *http.Request) *RequestLog {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	requestID := generateRequestID()

	// Parse query parameters
	query := make(map[string]interface{})
	for k, v := range r.URL.Query() {
		if len(v) == 1 {
			query[k] = v[0]
		} else {
			query[k] = v
		}
	}

	// Parse headers
	headers := make(map[string]interface{})
	for k, v := range r.Header {
		if len(v) == 1 {
			headers[k] = v[0]
		} else {
			headers[k] = v
		}
	}

	// Extract remote IP
	remoteIP := r.RemoteAddr
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		remoteIP = xff
	} else if xri := r.Header.Get("X-Real-IP"); xri != "" {
		remoteIP = xri
	}

	// Read body if present
	var body string
	if r.Body != nil && r.ContentLength > 0 && r.ContentLength < 10240 { // Limit to 10KB
		if bodyBytes := make([]byte, r.ContentLength); r.ContentLength > 0 {
			if n, err := r.Body.Read(bodyBytes); err == nil && n > 0 {
				body = string(bodyBytes[:n])
			}
		}
	}

	requestLog := &RequestLog{
		ID:        requestID,
		Method:    r.Method,
		Path:      r.URL.Path,
		URL:       r.URL.String(),
		StartTime: time.Now(),
		Headers:   headers,
		Query:     query,
		Body:      body,
		RemoteIP:  remoteIP,
		Logs:      make([]LogEntry, 0),
	}

	// Add to requests map and order tracking
	rl.requests[requestID] = requestLog
	rl.order = append(rl.order, requestID)

	// Enforce max logs limit (LRU eviction)
	if len(rl.order) > rl.maxLogs {
		oldestID := rl.order[0]
		delete(rl.requests, oldestID)
		rl.order = rl.order[1:]
	}

	return requestLog
}

// FinishRequest completes a request log entry
func (rl *RequestLogger) FinishRequest(requestID string, status int, response string, err error) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if requestLog, exists := rl.requests[requestID]; exists {
		requestLog.EndTime = time.Now()
		requestLog.Duration = requestLog.EndTime.Sub(requestLog.StartTime)
		requestLog.Status = status
		requestLog.Response = response
		if err != nil {
			requestLog.Error = err.Error()
		}
	}
}

// AddLog adds a log entry to a specific request
func (rl *RequestLogger) AddLog(requestID, level, message string, data interface{}) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if requestLog, exists := rl.requests[requestID]; exists {
		logEntry := LogEntry{
			Timestamp: time.Now(),
			Level:     level,
			Message:   message,
			Data:      data,
		}
		requestLog.Logs = append(requestLog.Logs, logEntry)
	}
}

// GetAllRequests returns all request logs in reverse chronological order
func (rl *RequestLogger) GetAllRequests() []*RequestLog {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	result := make([]*RequestLog, 0, len(rl.requests))

	// Return in reverse order (newest first)
	for i := len(rl.order) - 1; i >= 0; i-- {
		if req, exists := rl.requests[rl.order[i]]; exists {
			result = append(result, req)
		}
	}

	return result
}

// GetRequestByID returns a specific request log
func (rl *RequestLogger) GetRequestByID(requestID string) (*RequestLog, bool) {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	req, exists := rl.requests[requestID]
	return req, exists
}

// GetRecentRequests returns the most recent N requests
func (rl *RequestLogger) GetRecentRequests(count int) []*RequestLog {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	if count <= 0 || count > len(rl.order) {
		count = len(rl.order)
	}

	result := make([]*RequestLog, 0, count)

	// Get the last N requests in reverse order
	for i := len(rl.order) - 1; i >= len(rl.order)-count; i-- {
		if req, exists := rl.requests[rl.order[i]]; exists {
			result = append(result, req)
		}
	}

	return result
}

// ClearLogs clears all request logs
func (rl *RequestLogger) ClearLogs() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.requests = make(map[string]*RequestLog)
	rl.order = make([]string, 0)
}

// GetStats returns basic statistics about the logged requests
func (rl *RequestLogger) GetStats() map[string]interface{} {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	stats := map[string]interface{}{
		"totalRequests": len(rl.requests),
		"maxLogs":       rl.maxLogs,
	}

	// Count by status code
	statusCounts := make(map[string]int)
	methodCounts := make(map[string]int)
	var totalDuration time.Duration

	for _, req := range rl.requests {
		// Status code stats
		statusKey := "unknown"
		if req.Status > 0 {
			if req.Status >= 200 && req.Status < 300 {
				statusKey = "2xx"
			} else if req.Status >= 300 && req.Status < 400 {
				statusKey = "3xx"
			} else if req.Status >= 400 && req.Status < 500 {
				statusKey = "4xx"
			} else if req.Status >= 500 {
				statusKey = "5xx"
			}
		}
		statusCounts[statusKey]++

		// Method stats
		methodCounts[req.Method]++

		// Duration stats
		totalDuration += req.Duration
	}

	stats["statusCounts"] = statusCounts
	stats["methodCounts"] = methodCounts

	if len(rl.requests) > 0 {
		stats["averageDuration"] = totalDuration / time.Duration(len(rl.requests))
	}

	return stats
}

// generateRequestID creates a unique request ID
func generateRequestID() string {
	return time.Now().Format("20060102-150405.000000") + "-" + randomString(6)
}

// randomString generates a random string of given length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// RequestLoggerMiddleware creates an HTTP middleware that captures request logs
func (rl *RequestLogger) RequestLoggerMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Skip logging for admin endpoints to avoid recursion
		if r.URL.Path == "/admin/logs" || r.URL.Path == "/admin/logs/api" {
			next(w, r)
			return
		}

		requestLog := rl.StartRequest(r)

		// Capture response
		responseRecorder := &ResponseRecorder{
			ResponseWriter: w,
			status:         200,
			body:           make([]byte, 0),
		}

		// Process request
		next(responseRecorder, r)

		// Finish logging
		responseBody := ""
		if len(responseRecorder.body) < 1024 { // Only capture small responses
			responseBody = string(responseRecorder.body)
		}

		rl.FinishRequest(requestLog.ID, responseRecorder.status, responseBody, nil)

		log.Debug().
			Str("requestID", requestLog.ID).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", responseRecorder.status).
			Dur("duration", time.Since(requestLog.StartTime)).
			Msg("Request completed")
	}
}

// ResponseRecorder captures HTTP response for logging
type ResponseRecorder struct {
	http.ResponseWriter
	status int
	body   []byte
}

func (rr *ResponseRecorder) WriteHeader(status int) {
	rr.status = status
	rr.ResponseWriter.WriteHeader(status)
}

func (rr *ResponseRecorder) Write(b []byte) (int, error) {
	if len(rr.body) < 1024 { // Limit captured response size
		rr.body = append(rr.body, b...)
	}
	return rr.ResponseWriter.Write(b)
}
