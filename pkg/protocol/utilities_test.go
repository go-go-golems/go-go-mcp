package protocol

import (
	"encoding/json"
	"testing"
)

func TestNewCancellationNotification(t *testing.T) {
	requestID := "test-123"
	reason := "User requested cancellation"

	notif := NewCancellationNotification(requestID, reason)

	if notif.JSONRPC != "2.0" {
		t.Errorf("expected JSONRPC '2.0', got '%s'", notif.JSONRPC)
	}

	if notif.Method != "notifications/cancelled" {
		t.Errorf("expected method 'notifications/cancelled', got '%s'", notif.Method)
	}

	// Parse the params to verify they're correct
	var params CancellationParams
	if err := json.Unmarshal(notif.Params, &params); err != nil {
		t.Errorf("failed to unmarshal params: %v", err)
	}

	if params.RequestID != requestID {
		t.Errorf("expected RequestID '%s', got '%s'", requestID, params.RequestID)
	}

	if params.Reason != reason {
		t.Errorf("expected Reason '%s', got '%s'", reason, params.Reason)
	}
}
