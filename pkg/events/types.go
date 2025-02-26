package events

import (
	"context"
	"encoding/json"
	"fmt"
)

// UIEvent represents an event in the UI system that can trigger updates
type UIEvent struct {
	Type      string      `json:"type"`      // e.g. "component-update", "page-reload"
	PageID    string      `json:"pageId"`    // Which page is being updated
	Component interface{} `json:"component"` // The updated component data
}

// EventManager defines the interface for managing UI events
type EventManager interface {
	// Subscribe returns a channel that receives events for the specified page
	Subscribe(ctx context.Context, pageID string) (<-chan UIEvent, error)
	// Publish sends an event for the specified page
	Publish(pageID string, event UIEvent) error
	// Close cleans up resources used by the event manager
	Close() error
}

// NewPageReloadEvent creates a new event for reloading a page
func NewPageReloadEvent(pageID string, pageDef interface{}) UIEvent {
	return UIEvent{
		Type:      "page-reload",
		PageID:    pageID,
		Component: map[string]interface{}{"data": pageDef},
	}
}

// Validate checks if the event is valid
func (e UIEvent) Validate() error {
	if e.Type == "" {
		return fmt.Errorf("event type cannot be empty")
	}
	if e.PageID == "" {
		return fmt.Errorf("page ID cannot be empty")
	}
	return nil
}

// ToJSON converts the event to JSON bytes
func (e UIEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON creates an event from JSON bytes
func FromJSON(data []byte) (UIEvent, error) {
	var event UIEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return UIEvent{}, fmt.Errorf("failed to unmarshal event: %w", err)
	}
	return event, nil
}
