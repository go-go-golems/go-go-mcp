package engine

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dop251/goja"
	"github.com/rs/zerolog/log"
)

// StartDispatcher starts the job processing dispatcher
func (e *Engine) StartDispatcher() {
	log.Info().Msg("Starting JavaScript dispatcher")
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

	// Create enhanced request object
	reqObj := e.createEnhancedRequestObject(job.R)

	// Add path parameters if available
	if job.Handler.Options != nil {
		if pathPattern, ok := job.Handler.Options["pathPattern"].(string); ok {
			reqObj.Params = parsePathParams(pathPattern, job.R.URL.Path)
		}
	}

	// Call the JavaScript handler function
	result, err := job.Handler.Fn(goja.Undefined(), e.rt.ToValue(reqObj))
	if err != nil {
		log.Error().Err(err).Str("path", job.R.URL.Path).Msg("Handler execution error")
		http.Error(job.W, "Internal Server Error", http.StatusInternalServerError)
		return err
	}

	// Write response
	contentTypeOverrides := []string{}
	if job.Handler.ContentType != "" {
		contentTypeOverrides = append(contentTypeOverrides, job.Handler.ContentType)
	}

	if err := e.writeEnhancedResponse(job.W, result, contentTypeOverrides...); err != nil {
		log.Error().Err(err).Msg("Failed to write response")
		return err
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
		resultJSON := ""
		consoleLogJSON := ""
		errorStr := ""

		if result.Value != nil {
			if data, marshalErr := json.Marshal(result.Value); marshalErr == nil {
				resultJSON = string(data)
			}
		}

		if len(result.ConsoleLog) > 0 {
			if data, marshalErr := json.Marshal(result.ConsoleLog); marshalErr == nil {
				consoleLogJSON = string(data)
			}
		}

		if result.Error != nil {
			errorStr = result.Error.Error()
		}

		if storeErr := e.StoreScriptExecution(job.SessionID, job.Code, resultJSON, consoleLogJSON, errorStr, job.Source); storeErr != nil {
			log.Error().Err(storeErr).Msg("Failed to store script execution")
		}
	}

	// Send result to channel if provided
	if job.Result != nil {
		job.Result <- result
	}

	return err
}
