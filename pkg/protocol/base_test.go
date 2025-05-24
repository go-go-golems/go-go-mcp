package protocol

import (
	"encoding/json"
	"testing"
)

func TestBatchRequestValidation(t *testing.T) {
	tests := []struct {
		name      string
		batch     BatchRequest
		shouldErr bool
	}{
		{
			name:      "empty batch should error",
			batch:     BatchRequest{},
			shouldErr: true,
		},
		{
			name: "valid batch should pass",
			batch: BatchRequest{
				{JSONRPC: "2.0", Method: "test1", ID: json.RawMessage(`"1"`)},
				{JSONRPC: "2.0", Method: "test2", ID: json.RawMessage(`"2"`)},
			},
			shouldErr: false,
		},
		{
			name: "invalid JSONRPC version should error",
			batch: BatchRequest{
				{JSONRPC: "1.0", Method: "test1", ID: json.RawMessage(`"1"`)},
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.batch.Validate()
			if tt.shouldErr && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestBatchRequestGetByID(t *testing.T) {
	batch := BatchRequest{
		{JSONRPC: "2.0", Method: "test1", ID: json.RawMessage(`"1"`)},
		{JSONRPC: "2.0", Method: "test2", ID: json.RawMessage(`"2"`)},
	}

	// Test finding existing request
	req := batch.GetRequestByID(json.RawMessage(`"1"`))
	if req == nil {
		t.Error("expected to find request with ID '1'")
	}
	if req != nil && req.Method != "test1" {
		t.Errorf("expected method 'test1', got '%s'", req.Method)
	}

	// Test not finding non-existent request
	req = batch.GetRequestByID(json.RawMessage(`"3"`))
	if req != nil {
		t.Error("expected not to find request with ID '3'")
	}
}
