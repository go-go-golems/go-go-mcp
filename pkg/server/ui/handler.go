package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-go-golems/go-go-mcp/pkg/events"
	"github.com/go-go-golems/go-go-mcp/pkg/server/sse"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
)

// UIDefinition represents a UI page definition
type UIDefinition struct {
	Components []map[string]interface{} `yaml:"components"`
	RequestID  string                   `yaml:"requestID,omitempty"`
}

// UIHandler manages all UI-related functionality
type UIHandler struct {
	waitRegistry *WaitRegistry
	events       events.EventManager
	sseHandler   *sse.SSEHandler
	logger       *zerolog.Logger

	// Configuration
	timeout time.Duration
}

// NewUIHandler creates a new UI handler with the given dependencies
func NewUIHandler(events events.EventManager, sseHandler *sse.SSEHandler, logger *zerolog.Logger) *UIHandler {
	h := &UIHandler{
		waitRegistry: NewWaitRegistry(30 * time.Second), // 30 second default timeout
		events:       events,
		sseHandler:   sseHandler,
		logger:       logger,
		timeout:      30 * time.Second,
	}

	// Start background cleanup for orphaned requests
	go h.cleanupOrphanedRequests(context.Background())

	return h
}

// RegisterHandlers registers all UI-related HTTP handlers with the given mux
func (h *UIHandler) RegisterHandlers(mux *http.ServeMux) {
	mux.Handle("/api/ui-update", h.handleUIUpdate())
	mux.Handle("/api/ui-action", h.handleUIAction())
}

// cleanupOrphanedRequests periodically cleans up stale requests
func (h *UIHandler) cleanupOrphanedRequests(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Perform cleanup
			count := h.waitRegistry.CleanupStale()
			if count > 0 {
				h.logger.Debug().Int("count", count).Msg("Cleaned up orphaned requests")
			}

		case <-ctx.Done():
			return
		}
	}
}

// Helper method for sending error responses
func (h *UIHandler) sendErrorResponse(w http.ResponseWriter, status int, errorType, message string, details map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	errorResponse := map[string]interface{}{
		"status": "error",
		"error": map[string]interface{}{
			"type":    errorType,
			"message": message,
		},
	}

	// Add any additional details if provided
	for k, v := range details {
		errorResponse["error"].(map[string]interface{})[k] = v
	}

	_ = json.NewEncoder(w).Encode(errorResponse)
}

// Placeholder methods that will be implemented later
func (h *UIHandler) handleUIUpdate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only accept POST requests
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse JSON body
		var jsonData map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&jsonData); err != nil {
			// Return detailed JSON error response
			h.sendErrorResponse(w, http.StatusBadRequest, "json_parse_error", "Invalid JSON: "+err.Error(), nil)
			return
		}

		// Generate a unique request ID
		requestID := uuid.New().String()

		// Add the request ID to the UI definition
		// This will be used by the frontend to associate actions with this request
		jsonData["requestID"] = requestID

		// Convert to YAML for storage
		yamlData, err := yaml.Marshal(jsonData)
		if err != nil {
			h.sendErrorResponse(w, http.StatusInternalServerError, "yaml_conversion_error", "Failed to convert to YAML: "+err.Error(), nil)
			return
		}

		// Parse into UIDefinition
		var def UIDefinition
		if err := yaml.Unmarshal(yamlData, &def); err != nil {
			h.sendErrorResponse(w, http.StatusBadRequest, "ui_definition_error", "Invalid UI definition: "+err.Error(), map[string]interface{}{
				"yaml": string(yamlData),
			})
			return
		}

		// Validate the UI definition
		validationErrors := h.validateUIDefinition(def)
		if len(validationErrors) > 0 {
			h.sendErrorResponse(w, http.StatusBadRequest, "ui_validation_error", "UI definition validation failed", map[string]interface{}{
				"details": validationErrors,
			})
			return
		}

		// Register this request in the wait registry
		responseChan := h.waitRegistry.Register(requestID)

		// Create and publish refresh-ui event with the request ID
		event := events.UIEvent{
			Type:      "refresh-ui",
			PageID:    "ui-update",
			Component: def,
		}

		// Add the request ID to the event data if possible
		if compData, ok := event.Component.(UIDefinition); ok {
			compData.RequestID = requestID
			event.Component = compData
		}

		if err := h.events.Publish("ui-update", event); err != nil {
			// Clean up the registry entry
			h.waitRegistry.Cleanup(requestID)
			h.sendErrorResponse(w, http.StatusInternalServerError, "event_publish_error", "Failed to publish event: "+err.Error(), nil)
			return
		}

		// Log that we're waiting for a response
		h.logger.Debug().
			Str("requestId", requestID).
			Msg("Waiting for UI action response")

		// Wait for the response or timeout
		select {
		case response := <-responseChan:
			// Request was resolved with a user action
			if response.Error != nil {
				// There was an error processing the action
				h.sendErrorResponse(w, http.StatusInternalServerError, "action_processing_error", response.Error.Error(), nil)
				return
			}

			// Log the successful response
			h.logger.Info().
				Str("requestId", requestID).
				Str("action", response.Action).
				Str("componentId", response.ComponentID).
				Msg("UI action received, completing request")

			// Return success with the action data
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status":      "success",
				"action":      response.Action,
				"componentId": response.ComponentID,
				"data":        response.Data,
			})

		case <-time.After(h.timeout):
			// Request timed out
			h.waitRegistry.Cleanup(requestID)

			h.logger.Warn().
				Str("requestId", requestID).
				Dur("timeout", h.timeout).
				Msg("Request timed out waiting for UI action")

			h.sendErrorResponse(w, http.StatusRequestTimeout, "timeout", "No user action received within the timeout period", nil)
		}
	}
}

func (h *UIHandler) handleUIAction() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only accept POST requests
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse JSON body
		var action struct {
			ComponentID string                 `json:"componentId"`
			Action      string                 `json:"action"`
			Data        map[string]interface{} `json:"data"`
			RequestID   string                 `json:"requestId"` // Added field for request ID
		}

		if err := json.NewDecoder(r.Body).Decode(&action); err != nil {
			http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Determine if this is an important event to log at INFO level
		isImportantEvent := false
		switch action.Action {
		case "clicked", "changed", "submitted":
			isImportantEvent = true
		}

		// Log the action at appropriate level
		logger := h.logger.Debug()
		if isImportantEvent {
			logger = h.logger.Info()
		}

		// Create log entry with component and action info
		logger = logger.
			Str("componentId", action.ComponentID).
			Str("action", action.Action)

		// Add request ID to log if present
		if action.RequestID != "" {
			logger = logger.Str("requestId", action.RequestID)
		}

		// Add data to log if it exists and is relevant
		if len(action.Data) > 0 {
			// For form submissions, log the form data in detail
			if action.Action == "submitted" && action.Data["formData"] != nil {
				logger = logger.Interface("formData", action.Data["formData"])
			} else if action.Action == "changed" {
				// For changed events, log the new value
				logger = logger.Interface("data", action.Data)
			} else {
				// For other events, just log that data exists
				logger = logger.Bool("hasData", true)
			}
		}

		// Output the log message
		logger.Msg("UI action received")

		// Check if this action is associated with a waiting request
		if action.RequestID != "" && (action.Action == "clicked" || action.Action == "submitted") {
			// Try to resolve the waiting request
			resolved := h.waitRegistry.Resolve(action.RequestID, UIActionResponse{
				Action:      action.Action,
				ComponentID: action.ComponentID,
				Data:        action.Data,
				Error:       nil,
				Timestamp:   time.Now(),
			})

			if resolved {
				logger.Bool("waitingRequestResolved", true).Msg("Resolved waiting request")
			}
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(map[string]string{"status": "success"})
		if err != nil {
			http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		}
	}
}

// validateUIDefinition checks a UI definition for common errors
func (h *UIHandler) validateUIDefinition(def UIDefinition) []map[string]interface{} {
	var errors []map[string]interface{}

	// Check if components exist
	if len(def.Components) == 0 {
		errors = append(errors, map[string]interface{}{
			"path":    "components",
			"message": "No components defined",
		})
		return errors
	}

	// Validate each component
	for i, comp := range def.Components {
		for typ, props := range comp {
			propsMap, ok := props.(map[string]interface{})
			if !ok {
				errors = append(errors, map[string]interface{}{
					"path":    fmt.Sprintf("components[%d].%s", i, typ),
					"message": "Component properties must be a map",
				})
				continue
			}

			// Check for required ID
			if _, hasID := propsMap["id"]; !hasID && h.requiresID(typ) {
				errors = append(errors, map[string]interface{}{
					"path":    fmt.Sprintf("components[%d].%s", i, typ),
					"message": "Component is missing required 'id' property",
				})
			}

			// Validate nested components in forms
			if typ == "form" {
				if formComps, hasComps := propsMap["components"]; hasComps {
					if formCompsList, ok := formComps.([]interface{}); ok {
						for j, formComp := range formCompsList {
							if formCompMap, ok := formComp.(map[string]interface{}); ok {
								for formCompType, formCompProps := range formCompMap {
									if _, ok := formCompProps.(map[string]interface{}); !ok {
										errors = append(errors, map[string]interface{}{
											"path":    fmt.Sprintf("components[%d].%s.components[%d].%s", i, typ, j, formCompType),
											"message": "Form component properties must be a map",
										})
									}
								}
							} else {
								errors = append(errors, map[string]interface{}{
									"path":    fmt.Sprintf("components[%d].%s.components[%d]", i, typ, j),
									"message": "Form component must be a map",
								})
							}
						}
					} else {
						errors = append(errors, map[string]interface{}{
							"path":    fmt.Sprintf("components[%d].%s.components", i, typ),
							"message": "Form components must be an array",
						})
					}
				}
			}

			// Validate list items
			if typ == "list" {
				if items, hasItems := propsMap["items"]; hasItems {
					if itemsList, ok := items.([]interface{}); ok {
						for j, item := range itemsList {
							if _, ok := item.(map[string]interface{}); !ok {
								errors = append(errors, map[string]interface{}{
									"path":    fmt.Sprintf("components[%d].%s.items[%d]", i, typ, j),
									"message": "List item must be a map",
								})
							}
						}
					} else {
						errors = append(errors, map[string]interface{}{
							"path":    fmt.Sprintf("components[%d].%s.items", i, typ),
							"message": "List items must be an array",
						})
					}
				}
			}
		}
	}

	return errors
}

// requiresID returns true if the component type requires an ID
func (h *UIHandler) requiresID(componentType string) bool {
	switch componentType {
	case "text", "title":
		// These can optionally have IDs
		return false
	default:
		// All other components require IDs
		return true
	}
}
