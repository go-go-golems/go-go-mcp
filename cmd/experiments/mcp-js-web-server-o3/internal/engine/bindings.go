package engine

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/dop251/goja"
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
	})

	// Basic utilities
	e.rt.Set("JSON", map[string]interface{}{
		"stringify": e.jsonStringify,
		"parse":     e.jsonParse,
	})
}

// jsQuery executes SQL queries and returns results as JavaScript objects
func (e *Engine) jsQuery(query string, args ...interface{}) []map[string]interface{} {
	rows, err := e.db.Query(query, args...)
	if err != nil {
		log.Printf("SQL query error: %v", err)
		return nil
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		log.Printf("SQL columns error: %v", err)
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
			log.Printf("SQL scan error: %v", err)
			continue
		}

		rec := make(map[string]interface{})
		for i, col := range cols {
			rec[col] = vals[i]
		}
		result = append(result, rec)
	}

	return result
}

// registerHandler registers an HTTP handler function
func (e *Engine) registerHandler(method, path string, handler goja.Value) {
	callable, ok := goja.AssertFunction(handler)
	if !ok {
		panic(e.rt.NewTypeError("Handler must be a function"))
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if e.handlers[path] == nil {
		e.handlers[path] = make(map[string]goja.Callable)
	}
	e.handlers[path][method] = callable

	log.Printf("Registered handler: %s %s", method, path)
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
	log.Printf("Registered file handler: %s", path)
}

// consoleLog provides console.log functionality
func (e *Engine) consoleLog(args ...interface{}) {
	fmt.Print("[JS] ")
	for i, arg := range args {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Print(arg)
	}
	fmt.Println()
}

// consoleError provides console.error functionality
func (e *Engine) consoleError(args ...interface{}) {
	fmt.Print("[JS ERROR] ")
	for i, arg := range args {
		if i > 0 {
			fmt.Print(" ")
		}
		fmt.Print(arg)
	}
	fmt.Println()
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