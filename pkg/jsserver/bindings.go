package jsserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dop251/goja"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Database bindings

func (s *JSWebServer) createDBExecFunc(vm *goja.Runtime) func(string, ...interface{}) map[string]interface{} {
	return func(query string, args ...interface{}) map[string]interface{} {
		result, err := s.db.Exec(query, args...)
		if err != nil {
			panic(vm.NewGoError(err))
		}

		lastInsertID, _ := result.LastInsertId()
		rowsAffected, _ := result.RowsAffected()

		return map[string]interface{}{
			"lastInsertId": lastInsertID,
			"rowsAffected": rowsAffected,
		}
	}
}

func (s *JSWebServer) createDBQueryFunc(vm *goja.Runtime) func(string, ...interface{}) []map[string]interface{} {
	return func(query string, args ...interface{}) []map[string]interface{} {
		rows, err := s.db.Query(query, args...)
		if err != nil {
			panic(vm.NewGoError(err))
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			panic(vm.NewGoError(err))
		}

		var results []map[string]interface{}
		for rows.Next() {
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				panic(vm.NewGoError(err))
			}

			row := make(map[string]interface{})
			for i, col := range columns {
				val := values[i]
				if b, ok := val.([]byte); ok {
					row[col] = string(b)
				} else {
					row[col] = val
				}
			}
			results = append(results, row)
		}

		if err := rows.Err(); err != nil {
			panic(vm.NewGoError(err))
		}

		return results
	}
}

func (s *JSWebServer) createDBPrepareFunc(vm *goja.Runtime) func(string) map[string]interface{} {
	return func(query string) map[string]interface{} {
		stmt, err := s.db.Prepare(query)
		if err != nil {
			panic(vm.NewGoError(err))
		}

		return map[string]interface{}{
			"exec": func(args ...interface{}) map[string]interface{} {
				result, err := stmt.Exec(args...)
				if err != nil {
					panic(vm.NewGoError(err))
				}

				lastInsertID, _ := result.LastInsertId()
				rowsAffected, _ := result.RowsAffected()

				return map[string]interface{}{
					"lastInsertId": lastInsertID,
					"rowsAffected": rowsAffected,
				}
			},
			"close": func() {
				stmt.Close()
			},
		}
	}
}

func (s *JSWebServer) createDBTransactionFunc(vm *goja.Runtime) func(goja.Value) interface{} {
	return func(fn goja.Value) interface{} {
		tx, err := s.db.Begin()
		if err != nil {
			panic(vm.NewGoError(err))
		}

		// Create transaction object
		txObj := vm.NewObject()
		txObj.Set("exec", func(query string, args ...interface{}) map[string]interface{} {
			result, err := tx.Exec(query, args...)
			if err != nil {
				tx.Rollback()
				panic(vm.NewGoError(err))
			}

			lastInsertID, _ := result.LastInsertId()
			rowsAffected, _ := result.RowsAffected()

			return map[string]interface{}{
				"lastInsertId": lastInsertID,
				"rowsAffected": rowsAffected,
			}
		})

		txObj.Set("query", func(query string, args ...interface{}) []map[string]interface{} {
			rows, err := tx.Query(query, args...)
			if err != nil {
				tx.Rollback()
				panic(vm.NewGoError(err))
			}
			defer rows.Close()

			columns, err := rows.Columns()
			if err != nil {
				tx.Rollback()
				panic(vm.NewGoError(err))
			}

			var results []map[string]interface{}
			for rows.Next() {
				values := make([]interface{}, len(columns))
				valuePtrs := make([]interface{}, len(columns))
				for i := range values {
					valuePtrs[i] = &values[i]
				}

				if err := rows.Scan(valuePtrs...); err != nil {
					tx.Rollback()
					panic(vm.NewGoError(err))
				}

				row := make(map[string]interface{})
				for i, col := range columns {
					val := values[i]
					if b, ok := val.([]byte); ok {
						row[col] = string(b)
					} else {
						row[col] = val
					}
				}
				results = append(results, row)
			}

			return results
		})

		// Execute the function
		var result goja.Value
		if callable, ok := goja.AssertFunction(fn); ok {
			result, err = callable(goja.Undefined(), s.vm.ToValue(txObj))
		} else {
			err = fmt.Errorf("transaction function is not callable")
		}
		if err != nil {
			tx.Rollback()
			panic(err)
		}

		if err := tx.Commit(); err != nil {
			panic(vm.NewGoError(err))
		}

		return result
	}
}

// Server bindings

func (s *JSWebServer) createServerRouteFunc(vm *goja.Runtime, method string) func(string, goja.Value) {
	return func(path string, handler goja.Value) {
		if _, ok := goja.AssertFunction(handler); !ok {
			panic(vm.NewTypeError("Handler must be a function"))
		}

		routeKey := fmt.Sprintf("%s %s", method, path)

		s.mu.Lock()
		s.routes[routeKey] = JSRoute{
			Method:  method,
			Path:    path,
			Handler: handler,
			Created: time.Now(),
		}
		s.mu.Unlock()

		log.Printf("Registered route: %s %s", method, path)
	}
}

func (s *JSWebServer) createServerFileFunc(vm *goja.Runtime) func(string, goja.Value, string) {
	return func(path string, generator goja.Value, mimeType string) {
		if _, ok := goja.AssertFunction(generator); !ok {
			panic(vm.NewTypeError("Generator must be a function"))
		}

		s.mu.Lock()
		s.files[path] = JSFile{
			Path:      path,
			Generator: generator,
			MimeType:  mimeType,
			Created:   time.Now(),
		}
		s.mu.Unlock()

		log.Printf("Registered file: %s (%s)", path, mimeType)
	}
}

func (s *JSWebServer) createServerStaticFunc(vm *goja.Runtime) func(string, string, string) {
	return func(path, content, mimeType string) {
		s.mu.Lock()
		s.files[path] = JSFile{
			Path:      path,
			Generator: vm.ToValue(func() string { return content }),
			MimeType:  mimeType,
			Created:   time.Now(),
		}
		s.mu.Unlock()

		log.Printf("Registered static file: %s (%s)", path, mimeType)
	}
}

// State bindings

func (s *JSWebServer) createStateGetFunc(vm *goja.Runtime) func(string) interface{} {
	return func(key string) interface{} {
		s.mu.RLock()
		value, exists := s.globalState[key]
		s.mu.RUnlock()

		if !exists {
			return nil
		}
		return value
	}
}

func (s *JSWebServer) createStateSetFunc(vm *goja.Runtime) func(string, interface{}) {
	return func(key string, value interface{}) {
		s.mu.Lock()
		s.globalState[key] = value
		s.mu.Unlock()
	}
}

func (s *JSWebServer) createStateDeleteFunc(vm *goja.Runtime) func(string) {
	return func(key string) {
		s.mu.Lock()
		delete(s.globalState, key)
		s.mu.Unlock()
	}
}

func (s *JSWebServer) createStateClearFunc(vm *goja.Runtime) func() {
	return func() {
		s.mu.Lock()
		s.globalState = make(map[string]interface{})
		s.mu.Unlock()
	}
}

func (s *JSWebServer) createStateKeysFunc(vm *goja.Runtime) func() []string {
	return func() []string {
		s.mu.RLock()
		keys := make([]string, 0, len(s.globalState))
		for k := range s.globalState {
			keys = append(keys, k)
		}
		s.mu.RUnlock()
		return keys
	}
}

// Utils bindings

func (s *JSWebServer) createUtilsLogFunc(vm *goja.Runtime) func(...interface{}) {
	return func(args ...interface{}) {
		log.Println(args...)
	}
}

func (s *JSWebServer) createUtilsSleepFunc(vm *goja.Runtime) func(int) {
	return func(ms int) {
		time.Sleep(time.Duration(ms) * time.Millisecond)
	}
}

func (s *JSWebServer) createUtilsFetchFunc(vm *goja.Runtime) func(string, ...map[string]interface{}) map[string]interface{} {
	return func(url string, options ...map[string]interface{}) map[string]interface{} {
		client := &http.Client{Timeout: 30 * time.Second}

		method := "GET"
		var body []byte
		headers := make(map[string]string)

		if len(options) > 0 {
			opts := options[0]
			if m, ok := opts["method"].(string); ok {
				method = m
			}
			if h, ok := opts["headers"].(map[string]interface{}); ok {
				for k, v := range h {
					if s, ok := v.(string); ok {
						headers[k] = s
					}
				}
			}
			if b, ok := opts["body"]; ok {
				if s, ok := b.(string); ok {
					body = []byte(s)
				} else {
					bodyBytes, err := json.Marshal(b)
					if err != nil {
						panic(vm.NewGoError(err))
					}
					body = bodyBytes
					headers["Content-Type"] = "application/json"
				}
			}
		}

		req, err := http.NewRequest(method, url, bytes.NewReader(body))
		if err != nil {
			panic(vm.NewGoError(err))
		}

		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := client.Do(req)
		if err != nil {
			panic(vm.NewGoError(err))
		}
		defer resp.Body.Close()

		var responseBody interface{}
		contentType := resp.Header.Get("Content-Type")
		if contentType == "application/json" {
			json.NewDecoder(resp.Body).Decode(&responseBody)
		}

		return map[string]interface{}{
			"status":     resp.StatusCode,
			"statusText": resp.Status,
			"headers":    resp.Header,
			"body":       responseBody,
		}
	}
}

func (s *JSWebServer) createUtilsUUIDFunc(vm *goja.Runtime) func() string {
	return func() string {
		return uuid.New().String()
	}
}

func (s *JSWebServer) createUtilsNowFunc(vm *goja.Runtime) func() int64 {
	return func() int64 {
		return time.Now().UnixMilli()
	}
}

func (s *JSWebServer) createUtilsEnvFunc(vm *goja.Runtime) func(string) string {
	return func(key string) string {
		return os.Getenv(key)
	}
}

// Helper functions for request/response handling

func (s *JSWebServer) createJSRequest(r *http.Request) map[string]interface{} {
	// Parse form and JSON body
	var body interface{}
	if r.Header.Get("Content-Type") == "application/json" {
		json.NewDecoder(r.Body).Decode(&body)
	} else {
		r.ParseForm()
		formData := make(map[string]interface{})
		for k, v := range r.Form {
			if len(v) == 1 {
				formData[k] = v[0]
			} else {
				formData[k] = v
			}
		}
		body = formData
	}

	// Extract path parameters
	vars := mux.Vars(r)
	params := make(map[string]interface{})
	for k, v := range vars {
		params[k] = v
	}

	// Extract query parameters
	query := make(map[string]interface{})
	for k, v := range r.URL.Query() {
		if len(v) == 1 {
			query[k] = v[0]
		} else {
			query[k] = v
		}
	}

	return map[string]interface{}{
		"method":  r.Method,
		"url":     r.URL.String(),
		"path":    r.URL.Path,
		"headers": r.Header,
		"body":    body,
		"params":  params,
		"query":   query,
	}
}

func (s *JSWebServer) createJSResponse(w http.ResponseWriter) map[string]interface{} {
	return map[string]interface{}{
		"status": func(code int) map[string]interface{} {
			w.WriteHeader(code)
			return s.createJSResponse(w)
		},
		"json": func(data interface{}) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(data)
		},
		"text": func(text string) {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(text))
		},
		"html": func(html string) {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(html))
		},
		"header": func(key, value string) {
			w.Header().Set(key, value)
		},
	}
}
