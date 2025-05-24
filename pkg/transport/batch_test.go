package transport

import (
	"testing"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
)

func TestParseMessage(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectType  string
		shouldError bool
	}{
		{
			name:       "single request",
			input:      `{"jsonrpc":"2.0","id":"1","method":"test"}`,
			expectType: "*protocol.Request",
		},
		{
			name:       "batch request",
			input:      `[{"jsonrpc":"2.0","id":"1","method":"test1"},{"jsonrpc":"2.0","id":"2","method":"test2"}]`,
			expectType: "protocol.BatchRequest",
		},
		{
			name:        "invalid JSON",
			input:       `{invalid json}`,
			shouldError: true,
		},
		{
			name:        "empty batch",
			input:       `[]`,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseMessage([]byte(tt.input))

			if tt.shouldError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			switch tt.expectType {
			case "*protocol.Request":
				if _, ok := result.(*protocol.Request); !ok {
					t.Errorf("expected *protocol.Request, got %T", result)
				}
			case "protocol.BatchRequest":
				if _, ok := result.(protocol.BatchRequest); !ok {
					t.Errorf("expected protocol.BatchRequest, got %T", result)
				}
			}
		})
	}
}

func TestIsBatchMessage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "single request",
			input:    `{"jsonrpc":"2.0","id":"1","method":"test"}`,
			expected: false,
		},
		{
			name:     "batch request",
			input:    `[{"jsonrpc":"2.0","id":"1","method":"test"}]`,
			expected: true,
		},
		{
			name:     "batch with whitespace",
			input:    `  [{"jsonrpc":"2.0","id":"1","method":"test"}]  `,
			expected: true,
		},
		{
			name:     "empty",
			input:    ``,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsBatchMessage([]byte(tt.input))
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
