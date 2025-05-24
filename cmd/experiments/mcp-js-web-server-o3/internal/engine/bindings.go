package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/dop251/goja"
	"github.com/rs/zerolog/log"
)

// setupBindings configures JavaScript bindings for the runtime
func (e *Engine) setupBindings() {
	// SQLite database binding
	e.rt.Set("db", map[string]interface{}{
		"query": e.jsQuery,
	})

	// Handler registration
	e.rt.Set("registerHandler", e.registerHandler)
	e.rt.Set("registerFile", e.registerFile)

	// Console logging
	e.rt.Set("console", map[string]interface{}{
		"log":   e.consoleLog,
		"error": e.consoleError,
		"info":  e.consoleInfo,
		"warn":  e.consoleWarn,
		"debug": e.consoleDebug,
	})

	// Basic utilities
	e.rt.Set("JSON", map[string]interface{}{
		"stringify": e.jsonStringify,
		"parse":     e.jsonParse,
	})

	// Global state object for persistence across script executions
	e.rt.RunString(`
		if (typeof globalState === 'undefined') {
			globalState = {};
		}
	`)
	log.Debug().Msg("JavaScript bindings configured")
}

// jsQuery executes SQL queries and returns results as JavaScript objects
func (e *Engine) jsQuery(query string, args ...interface{}) []map[string]interface{} {
	log.Debug().Str("query", query).Interface("args", args).Msg("Executing SQL query")

	// Convert JavaScript arrays to individual arguments
	var flatArgs []interface{}
	for _, arg := range args {
		if slice, ok := arg.([]interface{}); ok {
			// If argument is a slice, spread its elements
			flatArgs = append(flatArgs, slice...)
		} else {
			// Otherwise, add the argument as-is
			flatArgs = append(flatArgs, arg)
		}
	}

	log.Debug().Str("query", query).Interface("flatArgs", flatArgs).Msg("Flattened SQL arguments")

	rows, err := e.db.Query(query, flatArgs...)
	if err != nil {
		log.Error().Err(err).Str("query", query).Interface("args", flatArgs).Msg("SQL query error")
		return nil
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		log.Error().Err(err).Msg("SQL columns error")
		return nil
	}

	var result []map[string]interface{}
	for rows.Next() {
		vals := make([]interface{}, len(cols))
		scan := make([]interface{}, len(cols))
		for i := range vals {
			scan[i] = &vals[i]
		}

		if err := rows.Scan(scan...); err != nil {
			log.Error().Err(err).Msg("SQL scan error")
			continue
		}

		rec := make(map[string]interface{})
		for i, col := range cols {
			rec[col] = vals[i]
		}
		result = append(result, rec)
	}

	log.Debug().Int("rows", len(result)).Msg("SQL query completed")
	return result
}

// registerHandler registers an HTTP handler function
// Usage: registerHandler(method, path, handler [, contentType])
func (e *Engine) registerHandler(method, path string, handler goja.Value, args ...goja.Value) {
	callable, ok := goja.AssertFunction(handler)
	if !ok {
		panic(e.rt.NewTypeError("Handler must be a function"))
	}

	// Optional content type parameter
	var contentType string
	if len(args) > 0 && !goja.IsUndefined(args[0]) && !goja.IsNull(args[0]) {
		contentType = args[0].String()
	}

	handlerInfo := &HandlerInfo{
		Fn:          callable,
		ContentType: contentType,
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

// consoleLog provides console.log functionality
func (e *Engine) consoleLog(args ...interface{}) {
	log.Info().Interface("args", args).Msg("JS console.log")
	fmt.Fprint(os.Stderr, "[JS] ")
	for i, arg := range args {
		if i > 0 {
			fmt.Fprint(os.Stderr, " ")
		}
		fmt.Fprint(os.Stderr, arg)
	}
	fmt.Fprintln(os.Stderr)
}

// consoleError provides console.error functionality
func (e *Engine) consoleError(args ...interface{}) {
	log.Error().Interface("args", args).Msg("JS console.error")
	fmt.Fprint(os.Stderr, "[JS ERROR] ")
	for i, arg := range args {
		if i > 0 {
			fmt.Fprint(os.Stderr, " ")
		}
		fmt.Fprint(os.Stderr, arg)
	}
	fmt.Fprintln(os.Stderr)
}

// consoleInfo provides console.info functionality
func (e *Engine) consoleInfo(args ...interface{}) {
	log.Info().Interface("args", args).Msg("JS console.info")
	fmt.Fprint(os.Stderr, "[JS INFO] ")
	for i, arg := range args {
		if i > 0 {
			fmt.Fprint(os.Stderr, " ")
		}
		fmt.Fprint(os.Stderr, arg)
	}
	fmt.Fprintln(os.Stderr)
}

// consoleWarn provides console.warn functionality
func (e *Engine) consoleWarn(args ...interface{}) {
	log.Warn().Interface("args", args).Msg("JS console.warn")
	fmt.Fprint(os.Stderr, "[JS WARN] ")
	for i, arg := range args {
		if i > 0 {
			fmt.Fprint(os.Stderr, " ")
		}
		fmt.Fprint(os.Stderr, arg)
	}
	fmt.Fprintln(os.Stderr)
}

// consoleDebug provides console.debug functionality
func (e *Engine) consoleDebug(args ...interface{}) {
	log.Debug().Interface("args", args).Msg("JS console.debug")
	fmt.Fprint(os.Stderr, "[JS DEBUG] ")
	for i, arg := range args {
		if i > 0 {
			fmt.Fprint(os.Stderr, " ")
		}
		fmt.Fprint(os.Stderr, arg)
	}
	fmt.Fprintln(os.Stderr)
}

// jsonStringify provides JSON.stringify functionality
func (e *Engine) jsonStringify(obj interface{}) string {
	data, err := json.Marshal(obj)
	if err != nil {
		return "null"
	}
	return string(data)
}

// jsonParse provides JSON.parse functionality
func (e *Engine) jsonParse(str string) interface{} {
	var result interface{}
	if err := json.Unmarshal([]byte(str), &result); err != nil {
		panic(e.rt.NewGoError(err))
	}
	return result
}

// ConsoleCapture holds original console functions and captured output
type ConsoleCapture struct {
	Log   func(...interface{})
	Error func(...interface{})
	Info  func(...interface{})
	Warn  func(...interface{})
	Debug func(...interface{})
}

// captureConsole replaces console functions to capture output
func (e *Engine) captureConsole(result *EvalResult) *ConsoleCapture {
	// Store original console functions
	original := &ConsoleCapture{
		Log:   e.consoleLog,
		Error: e.consoleError,
		Info:  e.consoleInfo,
		Warn:  e.consoleWarn,
		Debug: e.consoleDebug,
	}

	// Create capturing versions
	e.rt.Set("console", map[string]interface{}{
		"log":   func(args ...interface{}) { e.captureConsoleOutput(result, "log", args...) },
		"error": func(args ...interface{}) { e.captureConsoleOutput(result, "error", args...) },
		"info":  func(args ...interface{}) { e.captureConsoleOutput(result, "info", args...) },
		"warn":  func(args ...interface{}) { e.captureConsoleOutput(result, "warn", args...) },
		"debug": func(args ...interface{}) { e.captureConsoleOutput(result, "debug", args...) },
	})

	return original
}

// restoreConsole restores original console functions
func (e *Engine) restoreConsole(original *ConsoleCapture) {
	e.rt.Set("console", map[string]interface{}{
		"log":   original.Log,
		"error": original.Error,
		"info":  original.Info,
		"warn":  original.Warn,
		"debug": original.Debug,
	})
}

// captureConsoleOutput captures console output to the result
func (e *Engine) captureConsoleOutput(result *EvalResult, level string, args ...interface{}) {
	var parts []string
	for _, arg := range args {
		parts = append(parts, fmt.Sprint(arg))
	}
	output := fmt.Sprintf("[%s] %s", level, strings.Join(parts, " "))
	result.ConsoleLog = append(result.ConsoleLog, output)

	// Also call the original console function for logging
	switch level {
	case "log":
		e.consoleLog(args...)
	case "error":
		e.consoleError(args...)
	case "info":
		e.consoleInfo(args...)
	case "warn":
		e.consoleWarn(args...)
	case "debug":
		e.consoleDebug(args...)
	}
}
