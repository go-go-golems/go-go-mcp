package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-go-golems/geppetto/pkg/js"
	"github.com/rs/zerolog/log"
)

// setupBindings configures JavaScript bindings for the runtime
func (e *Engine) setupBindings() {
	// SQLite database binding
	if err := e.rt.Set("db", map[string]interface{}{
		"query": e.jsQuery,
		"exec":  e.jsExec,
	}); err != nil {
		log.Error().Err(err).Msg("Failed to set db binding")
	}

	// Handler registration
	if err := e.rt.Set("registerHandler", e.registerHandler); err != nil {
		log.Error().Err(err).Msg("Failed to set registerHandler binding")
	}
	if err := e.rt.Set("registerFile", e.registerFile); err != nil {
		log.Error().Err(err).Msg("Failed to set registerFile binding")
	}

	// HTTP utilities and constants
	e.setupHTTPUtilities()

	// HTTP request bindings
	e.setupHTTPBindings()

	// Console logging
	if err := e.rt.Set("console", map[string]interface{}{
		"log":   e.consoleLog,
		"error": e.consoleError,
		"info":  e.consoleInfo,
		"warn":  e.consoleWarn,
		"debug": e.consoleDebug,
	}); err != nil {
		log.Error().Err(err).Msg("Failed to set console binding")
	}

	// Basic utilities
	if err := e.rt.Set("JSON", map[string]interface{}{
		"stringify": e.jsonStringify,
		"parse":     e.jsonParse,
	}); err != nil {
		log.Error().Err(err).Msg("Failed to set JSON binding")
	}

	// Global state object for persistence across script executions
	if _, err := e.rt.RunString(`
		if (typeof globalState === 'undefined') {
			globalState = {};
		}
	`); err != nil {
		log.Error().Err(err).Msg("Failed to initialize globalState")
	}

	// Setup Geppetto JavaScript APIs
	e.setupGeppettoAPIs()

	log.Debug().Msg("JavaScript bindings configured")
}

// jsQuery executes SQL queries and returns results as JavaScript objects
func (e *Engine) jsQuery(query string, args ...interface{}) []map[string]interface{} {
	startTime := time.Now()
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

		// Log database operation if we have a current request
		if e.currentReqID != "" {
			dbOp := DatabaseOperation{
				Timestamp:  startTime,
				Type:       "query",
				SQL:        query,
				Parameters: flatArgs,
				Error:      err.Error(),
				Duration:   time.Since(startTime),
			}
			e.reqLogger.AddDatabaseOperation(e.currentReqID, dbOp)
		}

		return nil
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close database rows")
		}
	}()

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

	duration := time.Since(startTime)
	log.Debug().Int("rows", len(result)).Dur("duration", duration).Msg("SQL query completed")

	// Log database operation if we have a current request
	if e.currentReqID != "" {
		dbOp := DatabaseOperation{
			Timestamp:  startTime,
			Type:       "query",
			SQL:        query,
			Parameters: flatArgs,
			Result:     fmt.Sprintf("%d rows returned", len(result)),
			Duration:   duration,
		}
		e.reqLogger.AddDatabaseOperation(e.currentReqID, dbOp)
	}

	return result
}

// jsExec executes SQL statements without returning rows (INSERT, UPDATE, DELETE, CREATE, etc.)
func (e *Engine) jsExec(query string, args ...interface{}) map[string]interface{} {
	startTime := time.Now()
	log.Debug().Str("query", query).Interface("args", args).Msg("Executing SQL exec")

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

	log.Debug().Str("query", query).Interface("flatArgs", flatArgs).Msg("Flattened SQL exec arguments")

	result, err := e.db.Exec(query, flatArgs...)
	if err != nil {
		log.Error().Err(err).Str("query", query).Interface("args", flatArgs).Msg("SQL exec error")

		// Log database operation if we have a current request
		if e.currentReqID != "" {
			dbOp := DatabaseOperation{
				Timestamp:  startTime,
				Type:       "exec",
				SQL:        query,
				Parameters: flatArgs,
				Error:      err.Error(),
				Duration:   time.Since(startTime),
			}
			e.reqLogger.AddDatabaseOperation(e.currentReqID, dbOp)
		}

		return map[string]interface{}{
			"error":   err.Error(),
			"success": false,
		}
	}

	// Get affected rows and last insert ID if available
	rowsAffected, _ := result.RowsAffected()
	lastInsertId, _ := result.LastInsertId()

	duration := time.Since(startTime)
	log.Debug().Int64("rowsAffected", rowsAffected).Int64("lastInsertId", lastInsertId).Dur("duration", duration).Msg("SQL exec completed")

	// Log database operation if we have a current request
	if e.currentReqID != "" {
		dbOp := DatabaseOperation{
			Timestamp:    startTime,
			Type:         "exec",
			SQL:          query,
			Parameters:   flatArgs,
			Result:       fmt.Sprintf("success: %d rows affected", rowsAffected),
			Duration:     duration,
			RowsAffected: rowsAffected,
			LastInsertId: lastInsertId,
		}
		e.reqLogger.AddDatabaseOperation(e.currentReqID, dbOp)
	}

	return map[string]interface{}{
		"success":      true,
		"rowsAffected": rowsAffected,
		"lastInsertId": lastInsertId,
	}
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

	// Also log to request logger if we have a current request
	if e.currentReqID != "" {
		message := fmt.Sprintf("%v", args)
		e.reqLogger.AddLog(e.currentReqID, "log", message, args)
	}
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

	// Also log to request logger if we have a current request
	if e.currentReqID != "" {
		message := fmt.Sprintf("%v", args)
		e.reqLogger.AddLog(e.currentReqID, "error", message, args)
	}
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

	// Also log to request logger if we have a current request
	if e.currentReqID != "" {
		message := fmt.Sprintf("%v", args)
		e.reqLogger.AddLog(e.currentReqID, "info", message, args)
	}
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

	// Also log to request logger if we have a current request
	if e.currentReqID != "" {
		message := fmt.Sprintf("%v", args)
		e.reqLogger.AddLog(e.currentReqID, "warn", message, args)
	}
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

	// Also log to request logger if we have a current request
	if e.currentReqID != "" {
		message := fmt.Sprintf("%v", args)
		e.reqLogger.AddLog(e.currentReqID, "debug", message, args)
	}
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
	if err := e.rt.Set("console", map[string]interface{}{
		"log":   func(args ...interface{}) { e.captureConsoleOutput(result, "log", args...) },
		"error": func(args ...interface{}) { e.captureConsoleOutput(result, "error", args...) },
		"info":  func(args ...interface{}) { e.captureConsoleOutput(result, "info", args...) },
		"warn":  func(args ...interface{}) { e.captureConsoleOutput(result, "warn", args...) },
		"debug": func(args ...interface{}) { e.captureConsoleOutput(result, "debug", args...) },
	}); err != nil {
		log.Error().Err(err).Msg("Failed to set console capture binding")
	}

	return original
}

// restoreConsole restores original console functions
func (e *Engine) restoreConsole(original *ConsoleCapture) {
	if err := e.rt.Set("console", map[string]interface{}{
		"log":   original.Log,
		"error": original.Error,
		"info":  original.Info,
		"warn":  original.Warn,
		"debug": original.Debug,
	}); err != nil {
		log.Error().Err(err).Msg("Failed to restore console binding")
	}
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

// setupGeppettoAPIs configures Geppetto JavaScript APIs (Conversation, Embeddings, Steps, ChatStepFactory)
func (e *Engine) setupGeppettoAPIs() {
	log.Debug().Msg("Setting up Geppetto JavaScript APIs")

	// Register Conversation API
	if err := js.RegisterConversation(e.rt); err != nil {
		log.Error().Err(err).Msg("Failed to register Conversation API")
	} else {
		log.Debug().Msg("Conversation API registered")
	}

	// Register ChatStepFactory
	if err := js.RegisterFactory(e.rt, e.loop, e.stepSettings); err != nil {
		log.Error().Err(err).Msg("Failed to register ChatStepFactory")
	} else {
		log.Debug().Msg("ChatStepFactory registered")
	}

	// TODO: Register Embeddings API when available
	// This requires an embeddings provider which we don't have configured yet
	// if err := js.RegisterEmbeddings(e.rt, "embeddings", embeddingsProvider, e.loop); err != nil {
	//     log.Error().Err(err).Msg("Failed to register Embeddings API")
	// } else {
	//     log.Debug().Msg("Embeddings API registered")
	// }

	// TODO: Register additional Steps API when needed
	// Steps are typically registered on a per-use basis

	log.Debug().Msg("Geppetto JavaScript APIs setup complete")
}

// setupGeppettoBindings sets up only the Geppetto API bindings
func (e *Engine) setupGeppettoBindings() error {
	log.Debug().Msg("Setting up Geppetto JavaScript APIs")

	// Register Conversation API
	if err := js.RegisterConversation(e.rt); err != nil {
		log.Error().Err(err).Msg("Failed to register Conversation API")
		return err
	}
	log.Debug().Msg("Conversation API registered")

	// Register ChatStepFactory
	if err := js.RegisterFactory(e.rt, e.loop, e.stepSettings); err != nil {
		log.Error().Err(err).Msg("Failed to register ChatStepFactory")
		return err
	}
	log.Debug().Msg("ChatStepFactory registered")

	log.Debug().Msg("Geppetto JavaScript APIs setup complete")
	return nil
}
