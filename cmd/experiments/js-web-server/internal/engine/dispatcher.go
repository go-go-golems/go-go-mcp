package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/dop251/goja"
	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/internal/repository"
	"github.com/rs/zerolog/log"
)

// StartDispatcher starts the job processing dispatcher
func (e *Engine) StartDispatcher() {
	log.Debug().Msg("Starting JavaScript dispatcher")
	go e.dispatcher()
}

// dispatcher processes jobs from the job queue
func (e *Engine) dispatcher() {
	for job := range e.jobs {
		e.processJob(job)
	}
}

// processJob processes a single evaluation job
func (e *Engine) processJob(job EvalJob) {
	defer func() {
		if r := recover(); r != nil {
			log.Error().Interface("panic", r).Msg("Panic in JavaScript execution")
			if job.Done != nil {
				job.Done <- fmt.Errorf("panic in JavaScript execution: %v", r)
			}
		}
	}()

	// Start request logging if this is an HTTP request
	var requestLog *RequestLog
	if job.R != nil {
		requestLog = e.reqLogger.StartRequest(job.R)
		e.currentReqID = requestLog.ID
		defer func() {
			e.currentReqID = ""
		}()
	}

	var err error

	if job.Handler != nil {
		// Execute pre-registered handler
		err = e.executeHandler(job)
	} else {
		// Execute code directly
		err = e.executeDirectCode(job)
	}

	// Finish request logging
	if requestLog != nil {
		status := 200
		response := ""
		if responseRecorder, ok := job.W.(*ResponseRecorder); ok {
			status = responseRecorder.status
			if len(responseRecorder.body) < 1024 {
				response = string(responseRecorder.body)
			}
		}
		e.reqLogger.FinishRequest(requestLog.ID, status, response, err)
	}

	if job.Done != nil {
		job.Done <- err
	}
}

// executeHandler executes a pre-registered JavaScript handler function
func (e *Engine) executeHandler(job EvalJob) error {
	if job.Handler == nil || job.Handler.Fn == nil {
		return fmt.Errorf("no handler function provided")
	}

	log.Debug().Str("path", job.R.URL.Path).Str("method", job.R.Method).Msg("Creating Express.js request/response objects")

	// Create Express.js compatible request and response objects
	reqObj := e.createExpressRequestObject(job.R)
	resObj := e.createExpressResponseObject(job.W)

	log.Debug().
		Interface("reqObj", map[string]interface{}{
			"method":   reqObj.Method,
			"path":     reqObj.Path,
			"url":      reqObj.URL,
			"protocol": reqObj.Protocol,
			"hostname": reqObj.Hostname,
			"ip":       reqObj.IP,
		}).
		Interface("resObj", map[string]interface{}{
			"statusCode": resObj.StatusCode,
			"sent":       resObj.sent,
		}).
		Msg("Express.js objects created")

	// Add path parameters if available
	if job.Handler.Options != nil {
		if pathPattern, ok := job.Handler.Options["pathPattern"].(string); ok {
			reqObj.Params = parsePathParams(pathPattern, job.R.URL.Path)
			log.Debug().Str("pathPattern", pathPattern).Interface("params", reqObj.Params).Msg("Path parameters parsed")
		}
	}

	// Convert to Goja values and log their types
	reqValue := e.rt.ToValue(reqObj)
	resValue := e.rt.ToValue(resObj)

	// Use JavaScript JSON.stringify to get proper string representation
	reqJSON := e.stringifyJSValue(reqValue)
	resJSON := e.stringifyJSValue(resValue)

	log.Debug().
		Str("reqJSON", reqJSON).
		Str("resJSON", resJSON).
		Msg("Converted to Goja values")

	// Call the JavaScript handler function with Express.js style (req, res)
	log.Debug().Msg("Calling JavaScript handler function")
	v, err := job.Handler.Fn(goja.Undefined(), reqValue, resValue)
	log.Debug().Interface("v", v.Export()).Msg("Handler execution result")
	if err != nil {
		log.Error().Err(err).Str("path", job.R.URL.Path).Msg("Handler execution error")

		// Send error response if not already sent
		if !resObj.sent {
			log.Debug().Msg("Sending error response via http.Error")
			http.Error(job.W, "Internal Server Error", http.StatusInternalServerError)
		} else {
			log.Debug().Msg("Response already sent, not sending error response")
		}
		return err
	}

	// If the response wasn't sent by the handler, send a default response
	if !resObj.sent {
		log.Debug().Msg("Response not sent by handler, sending default 200 response")
		if err := resObj.Status(200).End(); err != nil {
			log.Error().Err(err).Msg("Failed to send default response")
		}
	} else {
		log.Debug().Msg("Response was sent by handler")
	}

	return nil
}

// executeDirectCode executes JavaScript code directly and captures results
func (e *Engine) executeDirectCode(job EvalJob) error {
	result, err := e.executeCodeWithResult(job.Code)
	if err != nil {
		log.Error().Err(err).Str("code", job.Code).Msg("Code execution error")
	}

	// Store execution result if we have session tracking
	if job.SessionID != "" {
		var resultStr, consoleLogStr, errorStr *string

		if result.Value != nil {
			if data, marshalErr := json.Marshal(result.Value); marshalErr == nil {
				s := string(data)
				resultStr = &s
			}
		}

		if len(result.ConsoleLog) > 0 {
			s := strings.Join(result.ConsoleLog, "\n")
			consoleLogStr = &s
		}

		if result.Error != nil {
			s := result.Error.Error()
			errorStr = &s
		}

		req := repository.CreateExecutionRequest{
			SessionID:  job.SessionID,
			Code:       job.Code,
			Result:     resultStr,
			ConsoleLog: consoleLogStr,
			Error:      errorStr,
			Source:     job.Source,
		}

		if _, storeErr := e.repos.Executions().CreateExecution(context.Background(), req); storeErr != nil {
			log.Error().Err(storeErr).Msg("Failed to store script execution")
		} else {
			log.Debug().Str("sessionID", job.SessionID).Msg("Script execution stored via repository")
		}
	}

	// Send result to channel if provided
	if job.Result != nil {
		job.Result <- result
	}

	return err
}
