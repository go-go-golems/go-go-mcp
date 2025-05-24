package engine

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/dop251/goja"
	"github.com/rs/zerolog/log"
)

// RequestObject represents the enhanced request object passed to JavaScript handlers
type RequestObject struct {
	Method   string                 `json:"method"`
	URL      string                 `json:"url"`
	Path     string                 `json:"path"`
	Query    map[string]interface{} `json:"query"`
	Headers  map[string]interface{} `json:"headers"`
	Body     string                 `json:"body,omitempty"`
	Params   map[string]string      `json:"params,omitempty"`   // URL path parameters
	Cookies  map[string]string      `json:"cookies,omitempty"`  // Request cookies
	RemoteIP string                 `json:"remoteIP,omitempty"` // Client IP address
}

// ResponseObject represents the enhanced response object returned from JavaScript handlers
type ResponseObject struct {
	Status      int               `json:"status,omitempty"`      // HTTP status code (defaults to 200)
	Headers     map[string]string `json:"headers,omitempty"`     // Response headers
	Body        interface{}       `json:"body,omitempty"`        // Response body
	ContentType string            `json:"contentType,omitempty"` // Content-Type header override
	Cookies     []ResponseCookie  `json:"cookies,omitempty"`     // Response cookies
	Redirect    string            `json:"redirect,omitempty"`    // Redirect URL (sets 302 status)
}

// ResponseCookie represents a cookie to be set in the response
type ResponseCookie struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Path     string `json:"path,omitempty"`
	Domain   string `json:"domain,omitempty"`
	MaxAge   int    `json:"maxAge,omitempty"`   // In seconds
	Secure   bool   `json:"secure,omitempty"`   // HTTPS only
	HttpOnly bool   `json:"httpOnly,omitempty"` // No JavaScript access
	SameSite string `json:"sameSite,omitempty"` // "Strict", "Lax", or "None"
}

// registerHandler registers an HTTP handler function with enhanced request/response support
// Usage: registerHandler(method, path, handler [, options])
func (e *Engine) registerHandler(method, path string, handler goja.Value, args ...goja.Value) {
	callable, ok := goja.AssertFunction(handler)
	if !ok {
		panic(e.rt.NewTypeError("Handler must be a function"))
	}

	// Parse optional options object
	var options map[string]interface{}
	if len(args) > 0 && !goja.IsUndefined(args[0]) && !goja.IsNull(args[0]) {
		if exported := args[0].Export(); exported != nil {
			if opts, ok := exported.(map[string]interface{}); ok {
				options = opts
			} else if contentType, ok := exported.(string); ok {
				// Backward compatibility: treat string as contentType
				options = map[string]interface{}{"contentType": contentType}
			}
		}
	}

	// Extract content type from options
	var contentType string
	if options != nil {
		if ct, ok := options["contentType"].(string); ok {
			contentType = ct
		}
	}

	handlerInfo := &HandlerInfo{
		Fn:          callable,
		ContentType: contentType,
		Options:     options,
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if e.handlers[path] == nil {
		e.handlers[path] = make(map[string]*HandlerInfo)
	}
	e.handlers[path][method] = handlerInfo

	if contentType != "" {
		log.Info().Str("method", method).Str("path", path).Str("content-type", contentType).Msg("Registered HTTP handler with content type")
	} else {
		log.Info().Str("method", method).Str("path", path).Msg("Registered HTTP handler")
	}
}

// registerFile registers a file handler function
func (e *Engine) registerFile(path string, handler goja.Value) {
	callable, ok := goja.AssertFunction(handler)
	if !ok {
		panic(e.rt.NewTypeError("File handler must be a function"))
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	e.files[path] = callable
	log.Info().Str("path", path).Msg("Registered file handler")
}

// createEnhancedRequestObject creates an enhanced JavaScript-compatible request object
func (e *Engine) createEnhancedRequestObject(r *http.Request) *RequestObject {
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

	// Parse cookies
	cookies := make(map[string]string)
	for _, cookie := range r.Cookies() {
		cookies[cookie.Name] = cookie.Value
	}

	// Extract remote IP
	remoteIP := r.RemoteAddr
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if parts := strings.Split(xff, ","); len(parts) > 0 {
			remoteIP = strings.TrimSpace(parts[0])
		}
	} else if xri := r.Header.Get("X-Real-IP"); xri != "" {
		remoteIP = xri
	}

	// Read body if present (for POST/PUT requests)
	var body string
	if r.Body != nil && r.ContentLength > 0 {
		// Note: In a production system, you'd want to limit body size
		if bodyBytes := make([]byte, r.ContentLength); r.ContentLength > 0 {
			if n, err := r.Body.Read(bodyBytes); err == nil && n > 0 {
				body = string(bodyBytes[:n])
			}
		}
	}

	return &RequestObject{
		Method:   r.Method,
		URL:      r.URL.String(),
		Path:     r.URL.Path,
		Query:    query,
		Headers:  headers,
		Body:     body,
		Cookies:  cookies,
		RemoteIP: remoteIP,
		// Params will be populated by path parameter matching if implemented
	}
}

// writeEnhancedResponse writes the JavaScript result to the HTTP response with enhanced features
func (e *Engine) writeEnhancedResponse(w http.ResponseWriter, result goja.Value, contentTypeOverride ...string) error {
	if result == nil || goja.IsUndefined(result) {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}

	exported := result.Export()

	// Check if result is a ResponseObject
	if respObj, ok := exported.(map[string]interface{}); ok {
		// Check if this looks like a ResponseObject
		if hasResponseFields(respObj) {
			return e.writeResponseObject(w, respObj)
		}
	}

	// Fallback to legacy response handling
	return e.writeLegacyResponse(w, exported, contentTypeOverride...)
}

// hasResponseFields checks if the object has response-like fields
func hasResponseFields(obj map[string]interface{}) bool {
	responseFields := []string{"status", "headers", "body", "contentType", "cookies", "redirect"}
	for _, field := range responseFields {
		if _, exists := obj[field]; exists {
			return true
		}
	}
	return false
}

// writeResponseObject handles ResponseObject format
func (e *Engine) writeResponseObject(w http.ResponseWriter, respObj map[string]interface{}) error {
	// Set status code (default to 200)
	status := 200
	if statusVal, ok := respObj["status"]; ok {
		if statusFloat, ok := statusVal.(float64); ok {
			status = int(statusFloat)
		} else if statusInt, ok := statusVal.(int); ok {
			status = statusInt
		}
	}

	// Handle redirect
	if redirectURL, ok := respObj["redirect"].(string); ok && redirectURL != "" {
		// Validate redirect URL
		if _, err := url.Parse(redirectURL); err == nil {
			if status < 300 || status >= 400 {
				status = http.StatusFound // 302
			}
			w.Header().Set("Location", redirectURL)
			w.WriteHeader(status)
			return nil
		} else {
			log.Warn().Str("redirect", redirectURL).Msg("Invalid redirect URL")
		}
	}

	// Set response headers
	if headers, ok := respObj["headers"].(map[string]interface{}); ok {
		for k, v := range headers {
			if headerValue, ok := v.(string); ok {
				w.Header().Set(k, headerValue)
			}
		}
	}

	// Set cookies
	if cookies, ok := respObj["cookies"].([]interface{}); ok {
		for _, cookieData := range cookies {
			if cookieMap, ok := cookieData.(map[string]interface{}); ok {
				cookie := e.createHTTPCookie(cookieMap)
				if cookie != nil {
					http.SetCookie(w, cookie)
				}
			}
		}
	}

	// Determine content type
	contentType := "application/json" // default
	if ct, ok := respObj["contentType"].(string); ok && ct != "" {
		contentType = ct
	}

	// Get body
	body := respObj["body"]

	// Write response
	if body == nil {
		w.WriteHeader(status)
		return nil
	}

	switch v := body.(type) {
	case []byte:
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(status)
		_, err := w.Write(v)
		return err
	case string:
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(status)
		_, err := w.Write([]byte(v))
		return err
	default:
		// JSON response
		if contentType == "application/json" || (!strings.Contains(contentType, "/")) {
			contentType = "application/json"
		}
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(status)
		return json.NewEncoder(w).Encode(v)
	}
}

// createHTTPCookie creates an http.Cookie from JavaScript cookie data
func (e *Engine) createHTTPCookie(cookieData map[string]interface{}) *http.Cookie {
	name, ok := cookieData["name"].(string)
	if !ok || name == "" {
		return nil
	}

	value, ok := cookieData["value"].(string)
	if !ok {
		value = ""
	}

	cookie := &http.Cookie{
		Name:  name,
		Value: value,
	}

	// Optional fields
	if path, ok := cookieData["path"].(string); ok && path != "" {
		cookie.Path = path
	}

	if domain, ok := cookieData["domain"].(string); ok && domain != "" {
		cookie.Domain = domain
	}

	if maxAge, ok := cookieData["maxAge"]; ok {
		if maxAgeFloat, ok := maxAge.(float64); ok {
			cookie.MaxAge = int(maxAgeFloat)
		} else if maxAgeInt, ok := maxAge.(int); ok {
			cookie.MaxAge = maxAgeInt
		}
	}

	if secure, ok := cookieData["secure"].(bool); ok {
		cookie.Secure = secure
	}

	if httpOnly, ok := cookieData["httpOnly"].(bool); ok {
		cookie.HttpOnly = httpOnly
	}

	if sameSite, ok := cookieData["sameSite"].(string); ok {
		switch strings.ToLower(sameSite) {
		case "strict":
			cookie.SameSite = http.SameSiteStrictMode
		case "lax":
			cookie.SameSite = http.SameSiteLaxMode
		case "none":
			cookie.SameSite = http.SameSiteNoneMode
		}
	}

	return cookie
}

// writeLegacyResponse handles legacy response format for backward compatibility
func (e *Engine) writeLegacyResponse(w http.ResponseWriter, exported interface{}, contentTypeOverride ...string) error {
	switch v := exported.(type) {
	case []byte:
		// Raw bytes - write directly
		w.Header().Set("Content-Type", "application/octet-stream")
		_, err := w.Write(v)
		return err
	case string:
		// String response - use override or detect content type
		var contentType string
		if len(contentTypeOverride) > 0 && contentTypeOverride[0] != "" {
			contentType = contentTypeOverride[0]
		} else {
			contentType = "text/plain; charset=utf-8"
			if isHTML(v) {
				contentType = "text/html; charset=utf-8"
			} else if isJSON(v) {
				contentType = "application/json"
			}
		}
		w.Header().Set("Content-Type", contentType)
		_, err := w.Write([]byte(v))
		return err
	default:
		// JSON response
		w.Header().Set("Content-Type", "application/json")
		return json.NewEncoder(w).Encode(v)
	}
}

// Helper functions for content type detection
func isHTML(s string) bool {
	trimmed := strings.TrimSpace(s)
	return strings.HasPrefix(strings.ToLower(trimmed), "<!doctype html") ||
		strings.HasPrefix(strings.ToLower(trimmed), "<html") ||
		strings.HasPrefix(trimmed, "<!")
}

func isJSON(s string) bool {
	trimmed := strings.TrimSpace(s)
	return (strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) ||
		(strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]"))
}

// Utility functions for JavaScript
func (e *Engine) setupHTTPUtilities() {
	// HTTP status codes
	e.rt.Set("HTTP", map[string]interface{}{
		"OK":                    200,
		"CREATED":               201,
		"ACCEPTED":              202,
		"NO_CONTENT":            204,
		"MOVED_PERMANENTLY":     301,
		"FOUND":                 302,
		"NOT_MODIFIED":          304,
		"BAD_REQUEST":           400,
		"UNAUTHORIZED":          401,
		"FORBIDDEN":             403,
		"NOT_FOUND":             404,
		"METHOD_NOT_ALLOWED":    405,
		"CONFLICT":              409,
		"INTERNAL_SERVER_ERROR": 500,
		"NOT_IMPLEMENTED":       501,
		"BAD_GATEWAY":           502,
		"SERVICE_UNAVAILABLE":   503,
	})

	// Response helper functions
	e.rt.Set("Response", map[string]interface{}{
		"json":     e.jsResponseJSON,
		"text":     e.jsResponseText,
		"html":     e.jsResponseHTML,
		"redirect": e.jsResponseRedirect,
		"error":    e.jsResponseError,
	})
}

// JavaScript response helper functions
func (e *Engine) jsResponseJSON(data interface{}, args ...interface{}) map[string]interface{} {
	response := map[string]interface{}{
		"body":        data,
		"contentType": "application/json",
	}

	// Optional status code
	if len(args) > 0 {
		if status, ok := args[0].(float64); ok {
			response["status"] = int(status)
		} else if status, ok := args[0].(int); ok {
			response["status"] = status
		}
	}

	return response
}

func (e *Engine) jsResponseText(text string, args ...interface{}) map[string]interface{} {
	response := map[string]interface{}{
		"body":        text,
		"contentType": "text/plain; charset=utf-8",
	}

	if len(args) > 0 {
		if status, ok := args[0].(float64); ok {
			response["status"] = int(status)
		} else if status, ok := args[0].(int); ok {
			response["status"] = status
		}
	}

	return response
}

func (e *Engine) jsResponseHTML(html string, args ...interface{}) map[string]interface{} {
	response := map[string]interface{}{
		"body":        html,
		"contentType": "text/html; charset=utf-8",
	}

	if len(args) > 0 {
		if status, ok := args[0].(float64); ok {
			response["status"] = int(status)
		} else if status, ok := args[0].(int); ok {
			response["status"] = status
		}
	}

	return response
}

func (e *Engine) jsResponseRedirect(url string, args ...interface{}) map[string]interface{} {
	status := 302 // Default to Found/Temporary redirect
	if len(args) > 0 {
		if statusCode, ok := args[0].(float64); ok {
			status = int(statusCode)
		} else if statusCode, ok := args[0].(int); ok {
			status = statusCode
		}
	}

	return map[string]interface{}{
		"redirect": url,
		"status":   status,
	}
}

func (e *Engine) jsResponseError(message string, args ...interface{}) map[string]interface{} {
	status := 500 // Default to Internal Server Error
	if len(args) > 0 {
		if statusCode, ok := args[0].(float64); ok {
			status = int(statusCode)
		} else if statusCode, ok := args[0].(int); ok {
			status = statusCode
		}
	}

	return map[string]interface{}{
		"body": map[string]interface{}{
			"error":   message,
			"status":  status,
			"success": false,
		},
		"status":      status,
		"contentType": "application/json",
	}
}

// parsePathParams extracts path parameters from URL (basic implementation)
// This is a simplified version - in production you'd want a more robust router
func parsePathParams(pattern, path string) map[string]string {
	params := make(map[string]string)

	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	if len(patternParts) != len(pathParts) {
		return params
	}

	for i, part := range patternParts {
		if strings.HasPrefix(part, ":") {
			paramName := part[1:]
			params[paramName] = pathParts[i]
		}
	}

	return params
}

// Helper to convert string to int for query parameters
func parseIntParam(value string, defaultValue int) int {
	if i, err := strconv.Atoi(value); err == nil {
		return i
	}
	return defaultValue
}

// Helper to convert string to bool for query parameters
func parseBoolParam(value string, defaultValue bool) bool {
	if b, err := strconv.ParseBool(value); err == nil {
		return b
	}
	return defaultValue
}
