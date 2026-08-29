package embeddable

import (
	"encoding/json"
	"testing"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
)

func TestMapToolResultToMCPForwardsMeta(t *testing.T) {
	result := protocol.NewErrorToolResult(protocol.NewTextContent("authentication required"))
	result.Meta = map[string]any{
		"mcp/www_authenticate": []string{`Bearer error="insufficient_scope"`},
	}
	mapped, err := mapToolResultToOfficial(result)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(mapped)
	if err != nil {
		t.Fatal(err)
	}
	var body map[string]any
	if err := json.Unmarshal(encoded, &body); err != nil {
		t.Fatal(err)
	}
	meta, ok := body["_meta"].(map[string]any)
	if !ok || meta["mcp/www_authenticate"] == nil {
		t.Fatalf("missing forwarded challenge metadata: %s", encoded)
	}
}
