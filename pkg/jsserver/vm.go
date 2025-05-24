package jsserver

import (
	"context"
	"fmt"
	"time"

	"github.com/dop251/goja"
	"github.com/pkg/errors"
)

func (s *JSWebServer) initJavaScript() error {
	s.vm = goja.New()
	
	// Set up timeout for JavaScript execution
	s.vm.SetMaxCallStackSize(1000)
	
	// Initialize JavaScript context with our APIs
	if err := s.setupJSBindings(); err != nil {
		return errors.Wrap(err, "failed to setup JavaScript bindings")
	}

	return nil
}

func (s *JSWebServer) executeJavaScript(req ExecuteRequest) ExecuteResponse {
	start := time.Now()
	executionID := fmt.Sprintf("%s-%d", start.Format("2006-01-02T15:04:05Z"), start.UnixNano())
	
	response := ExecuteResponse{
		ExecutionID: executionID,
	}

	// Create a new VM instance for this execution to avoid state pollution
	vm := goja.New()
	vm.SetMaxCallStackSize(1000)
	
	// Setup bindings for this execution
	if err := s.setupJSBindingsForVM(vm); err != nil {
		response.Error = fmt.Sprintf("Failed to setup JavaScript bindings: %v", err)
		return response
	}

	// Execute with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	done := make(chan struct{})
	var result goja.Value
	var execErr error

	go func() {
		defer func() {
			if r := recover(); r != nil {
				execErr = fmt.Errorf("JavaScript panic: %v", r)
			}
			close(done)
		}()

		result, execErr = vm.RunString(req.Code)
	}()

	select {
	case <-ctx.Done():
		response.Error = "JavaScript execution timeout"
		return response
	case <-done:
		if execErr != nil {
			response.Error = fmt.Sprintf("JavaScript execution error: %v", execErr)
			return response
		}
	}

	// Archive code if requested
	if req.Persist {
		archivedFile, err := s.archiveCode(req.Code, req.Name, executionID)
		if err != nil {
			response.Error = fmt.Sprintf("Failed to archive code: %v", err)
			return response
		}
		response.ArchivedFile = archivedFile
	}

	// Store execution in database
	if err := s.storeExecution(executionID, req.Code, result.String(), true, response.ArchivedFile); err != nil {
		// Log error but don't fail the response
		fmt.Printf("Failed to store execution: %v\n", err)
	}

	response.Success = true
	if result != nil {
		response.Result = result.String()
	}

	return response
}

func (s *JSWebServer) setupJSBindingsForVM(vm *goja.Runtime) error {
	// Database bindings
	dbObj := vm.NewObject()
	dbObj.Set("exec", s.createDBExecFunc(vm))
	dbObj.Set("query", s.createDBQueryFunc(vm))
	dbObj.Set("prepare", s.createDBPrepareFunc(vm))
	dbObj.Set("transaction", s.createDBTransactionFunc(vm))
	vm.Set("db", dbObj)

	// Server bindings
	serverObj := vm.NewObject()
	serverObj.Set("get", s.createServerRouteFunc(vm, "GET"))
	serverObj.Set("post", s.createServerRouteFunc(vm, "POST"))
	serverObj.Set("put", s.createServerRouteFunc(vm, "PUT"))
	serverObj.Set("delete", s.createServerRouteFunc(vm, "DELETE"))
	serverObj.Set("any", s.createServerRouteFunc(vm, "*"))
	serverObj.Set("file", s.createServerFileFunc(vm))
	serverObj.Set("static", s.createServerStaticFunc(vm))
	vm.Set("server", serverObj)

	// State bindings
	stateObj := vm.NewObject()
	stateObj.Set("get", s.createStateGetFunc(vm))
	stateObj.Set("set", s.createStateSetFunc(vm))
	stateObj.Set("delete", s.createStateDeleteFunc(vm))
	stateObj.Set("clear", s.createStateClearFunc(vm))
	stateObj.Set("keys", s.createStateKeysFunc(vm))
	vm.Set("state", stateObj)

	// Utils bindings
	utilsObj := vm.NewObject()
	utilsObj.Set("log", s.createUtilsLogFunc(vm))
	utilsObj.Set("sleep", s.createUtilsSleepFunc(vm))
	utilsObj.Set("fetch", s.createUtilsFetchFunc(vm))
	utilsObj.Set("uuid", s.createUtilsUUIDFunc(vm))
	utilsObj.Set("now", s.createUtilsNowFunc(vm))
	utilsObj.Set("env", s.createUtilsEnvFunc(vm))
	vm.Set("utils", utilsObj)

	return nil
}

func (s *JSWebServer) setupJSBindings() error {
	return s.setupJSBindingsForVM(s.vm)
}