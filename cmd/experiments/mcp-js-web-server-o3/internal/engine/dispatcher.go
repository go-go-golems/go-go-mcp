package engine

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/dop251/goja"
	"github.com/rs/zerolog/log"
)

// StartDispatcher starts the single-threaded JavaScript dispatcher goroutine
func (e *Engine) StartDispatcher() {
	log.Info().Msg("JavaScript dispatcher started")

	for job := range e.jobs {
		e.processJob(job)
	}

	log.Info().Msg("JavaScript dispatcher stopped")
}

// processJob handles a single evaluation job
func (e *Engine) processJob(job EvalJob) {
	defer func() {
		if r := recover(); r != nil {
			log.Error().Interface("panic", r).Msg("JavaScript panic during job execution")
			if job.W != nil {
				job.W.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(job.W, "JavaScript error: %v", r)
			}
			if job.Done != nil {
				job.Done <- fmt.Errorf("panic: %v", r)
			}
		}
	}()

	var err error

	if job.Handler != nil {
		log.Debug().Msg("Executing registered handler function")
		err = e.executeHandler(job)
	} else if job.Code != "" {
		log.Debug().Str("code", job.Code).Msg("Executing raw JavaScript code")
		var result *EvalResult
		var execErr error

		if job.Result != nil {
			// Execute with result capture
			result, execErr = e.executeCodeWithResult(job.Code)
			err = execErr
			if job.Result != nil {
				job.Result <- result
			}
		} else {
			// Execute without result capture (legacy mode)
			err = e.executeCode(job.Code)
		}

		// Store execution in database
		e.storeExecution(job, result, execErr)

		if job.W != nil {
			job.W.WriteHeader(http.StatusAccepted)
			job.W.Write([]byte("JavaScript executed"))
		}
	}

	if err != nil {
		log.Error().Err(err).Msg("Job execution failed")
	} else {
		log.Debug().Msg("Job execution completed successfully")
	}

	if job.Done != nil {
		job.Done <- err
	}
}

// executeHandler executes a registered JavaScript handler
func (e *Engine) executeHandler(job EvalJob) error {
	log.Debug().Str("method", job.R.Method).Str("path", job.R.URL.Path).Msg("Executing JavaScript handler")

	// Create enhanced request object for JavaScript
	reqObj := e.createEnhancedRequestObject(job.R)

	// Call the JavaScript function
	result, err := job.Handler.Fn(goja.Undefined(), e.rt.ToValue(reqObj))
	if err != nil {
		log.Error().Err(err).Str("method", job.R.Method).Str("path", job.R.URL.Path).Msg("Handler execution error")
		if job.W != nil {
			job.W.WriteHeader(http.StatusInternalServerError)
			job.W.Write([]byte(err.Error()))
		}
		return err
	}

	log.Debug().Str("method", job.R.Method).Str("path", job.R.URL.Path).Msg("Handler executed successfully")

	// Process the result with enhanced response handling
	return e.writeEnhancedResponse(job.W, result, job.Handler.ContentType)
}

// storeExecution stores the script execution results in the database
func (e *Engine) storeExecution(job EvalJob, result *EvalResult, execErr error) {
	if job.Code == "" {
		return // Don't store handler executions, only code executions
	}

	sessionID := job.SessionID
	if sessionID == "" {
		sessionID = "unknown"
	}

	source := job.Source
	if source == "" {
		source = "api"
	}

	var resultStr, consoleLogStr, errorStr string

	if result != nil {
		if result.Value != nil {
			if resultBytes, err := json.Marshal(result.Value); err == nil {
				resultStr = string(resultBytes)
			}
		}
		consoleLogStr = strings.Join(result.ConsoleLog, "\n")
	}

	if execErr != nil {
		errorStr = execErr.Error()
	}

	if err := e.StoreScriptExecution(sessionID, job.Code, resultStr, consoleLogStr, errorStr, source); err != nil {
		log.Warn().Err(err).Msg("Failed to store script execution")
	} else {
		log.Debug().Str("sessionID", sessionID).Str("source", source).Msg("Stored script execution")
	}
}
