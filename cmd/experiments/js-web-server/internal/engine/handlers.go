package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/dop251/goja"
	"github.com/rs/zerolog/log"
)

// minInt returns the minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ExpressRequest represents an Express.js compatible request object
type ExpressRequest struct {
	Method   string                 `json:"method"`
	URL      string                 `json:"url"`
	Path     string                 `json:"path"`
	Query    map[string]interface{} `json:"query"`
	Headers  map[string]interface{} `json:"headers"`
	Body     interface{}            `json:"body"`
	Cookies  map[string]string      `json:"cookies"`
	IP       string                 `json:"ip"`
	Protocol string                 `json:"protocol"`
	Hostname string                 `json:"hostname"`
	Params   map[string]string      `json:"params"`
}

// ExpressResponse represents an Express.js compatible response object
type ExpressResponse struct {
	StatusCode int                 `json:"statusCode"`
	Headers    map[string]string   `json:"headers"`
	Cookies    []*http.Cookie      `json:"cookies"`
	writer     http.ResponseWriter `json:"-"`
	engine     *Engine             `json:"-"`
	sent       bool                `json:"-"`
}

// Express.js response methods

// Status sets the HTTP status code
func (r *ExpressResponse) Status(code interface{}) *ExpressResponse {
	log.Debug().Interface("code", code).Bool("sent", r.sent).Msg("ExpressResponse.Status called")

	if r.sent {
		log.Debug().Msg("Response already sent, ignoring Status call")
		return r
	}
	if statusCode, ok := code.(float64); ok {
		r.StatusCode = int(statusCode)
		log.Debug().Int("statusCode", r.StatusCode).Msg("Status set from float64")
	} else if statusCode, ok := code.(int); ok {
		r.StatusCode = statusCode
		log.Debug().Int("statusCode", r.StatusCode).Msg("Status set from int")
	} else {
		log.Debug().Interface("code", code).Str("type", fmt.Sprintf("%T", code)).Msg("Unknown status code type")
	}
	return r
}

// Send sends a response
func (r *ExpressResponse) Send(data interface{}) error {
	log.Debug().Interface("data", data).Bool("sent", r.sent).Int("statusCode", r.StatusCode).Msg("ExpressResponse.Send called")

	if r.sent {
		log.Debug().Msg("Response already sent, ignoring Send call")
		return nil
	}
	r.sent = true

	// Set default status if not set
	if r.StatusCode == 0 {
		r.StatusCode = 200
	}

	// Set any pending headers
	for key, value := range r.Headers {
		r.writer.Header().Set(key, value)
		log.Debug().Str("key", key).Str("value", value).Msg("Setting header")
	}

	// Set any pending cookies
	for _, cookie := range r.Cookies {
		http.SetCookie(r.writer, cookie)
		log.Debug().Str("name", cookie.Name).Str("value", cookie.Value).Msg("Setting cookie")
	}

	switch v := data.(type) {
	case string:
		// Only auto-detect content type if not already set
		if r.writer.Header().Get("Content-Type") == "" {
			if isHTML(v) {
				r.writer.Header().Set("Content-Type", "text/html; charset=utf-8")
				log.Debug().Msg("Detected HTML content")
			} else if isJSON(v) {
				r.writer.Header().Set("Content-Type", "application/json")
				log.Debug().Msg("Detected JSON content")
			} else {
				r.writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
				log.Debug().Msg("Detected plain text content")
			}
		}
		r.writer.WriteHeader(r.StatusCode)
		log.Debug().Int("statusCode", r.StatusCode).Str("content", v).Msg("Writing string response")
		_, err := r.writer.Write([]byte(v))
		return err
	case []byte:
		// Only set content type if not already set
		if r.writer.Header().Get("Content-Type") == "" {
			r.writer.Header().Set("Content-Type", "application/octet-stream")
		}
		r.writer.WriteHeader(r.StatusCode)
		log.Debug().Int("statusCode", r.StatusCode).Int("bytes", len(v)).Msg("Writing byte response")
		_, err := r.writer.Write(v)
		return err
	default:
		// Only set JSON content type if not already set
		if r.writer.Header().Get("Content-Type") == "" {
			r.writer.Header().Set("Content-Type", "application/json")
		}
		r.writer.WriteHeader(r.StatusCode)
		log.Debug().Int("statusCode", r.StatusCode).Interface("data", v).Msg("Writing JSON object response")
		return json.NewEncoder(r.writer).Encode(v)
	}
}

// JSON sends a JSON response
func (r *ExpressResponse) Json(data interface{}) error {
	log.Debug().Interface("data", data).Bool("sent", r.sent).Int("statusCode", r.StatusCode).Msg("ExpressResponse.JSON called")

	if r.sent {
		log.Debug().Msg("Response already sent, ignoring JSON call")
		return nil
	}
	r.sent = true

	if r.StatusCode == 0 {
		r.StatusCode = 200
	}

	// Set any pending headers
	for key, value := range r.Headers {
		r.writer.Header().Set(key, value)
		log.Debug().Str("key", key).Str("value", value).Msg("Setting header")
	}

	// Set any pending cookies
	for _, cookie := range r.Cookies {
		http.SetCookie(r.writer, cookie)
		log.Debug().Str("name", cookie.Name).Str("value", cookie.Value).Msg("Setting cookie")
	}

	r.writer.Header().Set("Content-Type", "application/json")
	r.writer.WriteHeader(r.StatusCode)
	log.Debug().Int("statusCode", r.StatusCode).Msg("Writing JSON response")
	return json.NewEncoder(r.writer).Encode(data)
}

// Redirect redirects the request
func (r *ExpressResponse) Redirect(args ...interface{}) error {
	if r.sent {
		return nil
	}
	r.sent = true

	status := 302
	var url string

	if len(args) == 1 {
		// Single argument - URL with default 302 status
		if u, ok := args[0].(string); ok {
			url = u
		}
	} else if len(args) == 2 {
		// Two arguments - status and URL
		if s, ok := args[0].(int); ok {
			status = s
		} else if s, ok := args[0].(float64); ok {
			status = int(s)
		}
		if u, ok := args[1].(string); ok {
			url = u
		}
	}

	if url == "" {
		return fmt.Errorf("redirect URL is required")
	}

	r.writer.Header().Set("Location", url)
	r.writer.WriteHeader(status)
	return nil
}

// Set sets a response header
func (r *ExpressResponse) Set(name, value string) *ExpressResponse {
	if !r.sent {
		r.Headers[name] = value
	}
	return r
}

// Cookie sets a response cookie
func (r *ExpressResponse) Cookie(name, value string, options ...interface{}) *ExpressResponse {
	if r.sent {
		return r
	}

	cookie := &http.Cookie{
		Name:  name,
		Value: value,
	}

	// Parse options if provided
	if len(options) > 0 {
		if opts, ok := options[0].(map[string]interface{}); ok {
			if path, ok := opts["path"].(string); ok {
				cookie.Path = path
			}
			if domain, ok := opts["domain"].(string); ok {
				cookie.Domain = domain
			}
			if maxAge, ok := opts["maxAge"].(float64); ok {
				cookie.MaxAge = int(maxAge)
			} else if maxAge, ok := opts["maxAge"].(int); ok {
				cookie.MaxAge = maxAge
			}
			if secure, ok := opts["secure"].(bool); ok {
				cookie.Secure = secure
			}
			if httpOnly, ok := opts["httpOnly"].(bool); ok {
				cookie.HttpOnly = httpOnly
			}
			if sameSite, ok := opts["sameSite"].(string); ok {
				switch strings.ToLower(sameSite) {
				case "strict":
					cookie.SameSite = http.SameSiteStrictMode
				case "lax":
					cookie.SameSite = http.SameSiteLaxMode
				case "none":
					cookie.SameSite = http.SameSiteNoneMode
				}
			}
		}
	}

	r.Cookies = append(r.Cookies, cookie)
	return r
}

// End ends the response
func (r *ExpressResponse) End(data ...interface{}) error {
	if r.sent {
		return nil
	}
	r.sent = true

	if r.StatusCode == 0 {
		r.StatusCode = 200
	}

	// Set any pending headers
	for key, value := range r.Headers {
		r.writer.Header().Set(key, value)
	}

	// Set any pending cookies
	for _, cookie := range r.Cookies {
		http.SetCookie(r.writer, cookie)
	}

	if len(data) > 0 {
		return r.Send(data[0])
	} else {
		r.writer.WriteHeader(r.StatusCode)
		return nil
	}
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

	// Store the original path pattern for parameter extraction
	if options == nil {
		options = make(map[string]interface{})
	}
	options["pathPattern"] = path

	// XXX I don't think we need the ContentType and Options here any more since everything goes through app.get/*
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

// appGet registers a GET route handler (Express.js style)
func (e *Engine) appGet(path string, handler goja.Value) {
	e.registerHandler("GET", path, handler)
}

// appPost registers a POST route handler (Express.js style)
func (e *Engine) appPost(path string, handler goja.Value) {
	e.registerHandler("POST", path, handler)
}

// appPut registers a PUT route handler (Express.js style)
func (e *Engine) appPut(path string, handler goja.Value) {
	e.registerHandler("PUT", path, handler)
}

// appDelete registers a DELETE route handler (Express.js style)
func (e *Engine) appDelete(path string, handler goja.Value) {
	e.registerHandler("DELETE", path, handler)
}

// appPatch registers a PATCH route handler (Express.js style)
func (e *Engine) appPatch(path string, handler goja.Value) {
	e.registerHandler("PATCH", path, handler)
}

// appUse registers middleware or route handler (Express.js style)
func (e *Engine) appUse(args ...goja.Value) {
	// Basic implementation - if only one argument, it's a middleware for all routes
	// If two arguments, first is path and second is handler
	if len(args) == 1 {
		// Global middleware (simplified implementation)
		handler := args[0]
		// Register for common HTTP methods
		methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
		for _, method := range methods {
			e.registerHandler(method, "/*", handler)
		}
	} else if len(args) == 2 {
		// Path-specific handler
		pathValue := args[0].Export()
		if path, ok := pathValue.(string); ok {
			handler := args[1]
			// Register for common HTTP methods
			methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
			for _, method := range methods {
				e.registerHandler(method, path, handler)
			}
		}
	}
}

// Utility functions for JavaScript
func (e *Engine) setupHTTPUtilities() {
	// Express.js style app object
	if err := e.rt.Set("app", map[string]interface{}{
		"get":    e.appGet,
		"post":   e.appPost,
		"put":    e.appPut,
		"delete": e.appDelete,
		"patch":  e.appPatch,
		"use":    e.appUse,
	}); err != nil {
		log.Error().Err(err).Msg("Failed to set app binding")
	}

	// Legacy registerHandler for backward compatibility
	if err := e.rt.Set("registerHandler", e.registerHandler); err != nil {
		log.Error().Err(err).Msg("Failed to set registerHandler binding")
	}
	if err := e.rt.Set("registerFile", e.registerFile); err != nil {
		log.Error().Err(err).Msg("Failed to set registerFile binding")
	}

	// HTTP status codes (Express.js compatible)
	if err := e.rt.Set("HTTP", map[string]interface{}{
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
	}); err != nil {
		log.Error().Err(err).Msg("Failed to set HTTP constants binding")
	}

}

// pathMatches checks if a URL path matches a pattern with parameters
func pathMatches(pattern, path string) bool {
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	if len(patternParts) != len(pathParts) {
		return false
	}

	for i, part := range patternParts {
		if !strings.HasPrefix(part, ":") && part != pathParts[i] {
			return false
		}
	}

	return true
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

// createExpressRequestObject creates an Express.js compatible request object
func (e *Engine) createExpressRequestObject(r *http.Request) *ExpressRequest {
	log.Debug().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Int64("contentLength", r.ContentLength).
		Str("contentType", r.Header.Get("Content-Type")).
		Msg("Creating Express request object")

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
			headers[strings.ToLower(k)] = v[0]
		} else {
			headers[strings.ToLower(k)] = v
		}
	}

	// Parse cookies
	cookies := make(map[string]string)
	for _, cookie := range r.Cookies() {
		cookies[cookie.Name] = cookie.Value
	}

	// Extract client IP
	ip := r.RemoteAddr
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if parts := strings.Split(xff, ","); len(parts) > 0 {
			ip = strings.TrimSpace(parts[0])
		}
	} else if xri := r.Header.Get("X-Real-IP"); xri != "" {
		ip = xri
	}

	// Extract and parse request body
	body := extractRequestBody(r)
	log.Debug().
		Interface("body", body).
		Str("bodyType", fmt.Sprintf("%T", body)).
		Msg("Request body extracted")

	// Determine protocol
	protocol := "http"
	if r.TLS != nil {
		protocol = "https"
	}

	// Extract hostname (without port)
	hostname := r.Host
	if colonIndex := strings.Index(hostname, ":"); colonIndex != -1 {
		hostname = hostname[:colonIndex]
	}

	return &ExpressRequest{
		Method:   strings.ToLower(r.Method),
		URL:      r.URL.String(),
		Path:     r.URL.Path,
		Query:    query,
		Headers:  headers,
		Body:     body,
		Cookies:  cookies,
		IP:       ip,
		Protocol: protocol,
		Hostname: hostname,
		Params:   make(map[string]string), // will be populated by path matching
	}
}

// createExpressResponseObject creates an Express.js compatible response object
func (e *Engine) createExpressResponseObject(w http.ResponseWriter) *ExpressResponse {
	return &ExpressResponse{
		StatusCode: 200,
		Headers:    make(map[string]string),
		Cookies:    make([]*http.Cookie, 0),
		writer:     w,
		engine:     e,
		sent:       false,
	}
}

// Helper function to extract request body
func extractRequestBody(r *http.Request) interface{} {
	log.Debug().Bool("bodyIsNil", r.Body == nil).Int64("contentLength", r.ContentLength).Msg("extractRequestBody called")

	if r.Body == nil {
		log.Debug().Msg("Request body is nil")
		return nil
	}

	// Read the body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read request body")
		return nil
	}

	log.Debug().
		Int("bodyBytesLength", len(bodyBytes)).
		Str("bodyBytesPreview", string(bodyBytes[:minInt(len(bodyBytes), 100)])).
		Msg("Read request body bytes")

	// Restore the body for further processing
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Try to parse as JSON if Content-Type indicates JSON
	contentType := r.Header.Get("Content-Type")
	log.Debug().Str("contentType", contentType).Bool("isJSON", strings.Contains(contentType, "application/json")).Msg("Checking content type")

	if strings.Contains(contentType, "application/json") && len(bodyBytes) > 0 {
		var jsonData interface{}
		if err := json.Unmarshal(bodyBytes, &jsonData); err == nil {
			log.Debug().Interface("parsedJSON", jsonData).Msg("Successfully parsed JSON")
			return jsonData
		} else {
			log.Debug().Err(err).Msg("Failed to parse JSON")
		}
	}

	// Return as string for other content types
	result := string(bodyBytes)
	log.Debug().Str("finalResult", result).Msg("Returning body as string")
	return result
}
