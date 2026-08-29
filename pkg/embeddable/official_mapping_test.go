package embeddable

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
)

func TestMapToolToOfficialPreservesDescriptor(t *testing.T) {
	readOnly := true
	tool := protocol.Tool{
		Name:         "lookup",
		Title:        "Lookup",
		Description:  "Look up one item",
		InputSchema:  json.RawMessage(`{"type":"object"}`),
		OutputSchema: json.RawMessage(`{"type":"object","required":["id"]}`),
		Annotations: &protocol.ToolAnnotations{
			Title:           "Lookup annotation",
			ReadOnlyHint:    true,
			DestructiveHint: &readOnly,
			IdempotentHint:  true,
			OpenWorldHint:   &readOnly,
		},
		Meta: map[string]any{"securitySchemes": []any{map[string]any{"type": "oauth2"}}},
	}

	mapped := mapToolToOfficial(tool)
	encoded, err := json.Marshal(mapped)
	if err != nil {
		t.Fatal(err)
	}
	var wire map[string]any
	if err := json.Unmarshal(encoded, &wire); err != nil {
		t.Fatal(err)
	}
	if wire["name"] != "lookup" || wire["title"] != "Lookup" {
		t.Fatalf("descriptor identity = %#v", wire)
	}
	if wire["outputSchema"] == nil || wire["annotations"] == nil || wire["_meta"] == nil {
		t.Fatalf("descriptor fields were lost: %s", encoded)
	}
}

func TestMapToolResultToOfficialPreservesAllContent(t *testing.T) {
	image := base64.StdEncoding.EncodeToString([]byte("image"))
	blob := base64.StdEncoding.EncodeToString([]byte("blob"))
	result := &protocol.ToolResult{
		Content: []protocol.ToolContent{
			protocol.NewTextContent("hello"),
			protocol.NewImageContent(image, "image/png"),
			protocol.NewResourceContent(&protocol.ResourceContent{URI: "coinvault://doc/1", MimeType: "text/plain", Blob: blob}),
		},
		StructuredContent: map[string]any{"answer": "hello"},
		Meta:              map[string]any{"test/result-meta": "preserved"},
	}

	mapped, err := mapToolResultToOfficial(result)
	if err != nil {
		t.Fatal(err)
	}
	if len(mapped.Content) != 3 {
		t.Fatalf("content count = %d", len(mapped.Content))
	}
	if mapped.StructuredContent.(map[string]any)["answer"] != "hello" {
		t.Fatalf("structured content = %#v", mapped.StructuredContent)
	}
	if mapped.Meta["test/result-meta"] != "preserved" {
		t.Fatalf("meta = %#v", mapped.Meta)
	}
	encoded, err := json.Marshal(mapped)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(encoded) {
		t.Fatalf("invalid wire result: %s", encoded)
	}
}

func TestMapToolResultToOfficialRejectsInvalidContent(t *testing.T) {
	tests := []protocol.ToolContent{
		{Type: "image", Data: "not base64"},
		{Type: "resource"},
		{Type: "unknown"},
	}
	for _, content := range tests {
		if _, err := mapToolResultToOfficial(&protocol.ToolResult{Content: []protocol.ToolContent{content}}); err == nil {
			t.Fatalf("expected error for %#v", content)
		}
	}
}

func TestWithJSONPopulatesCompatibilityAndStructuredContent(t *testing.T) {
	result := protocol.NewToolResult(protocol.WithJSON(map[string]any{"ok": true}))
	if len(result.Content) != 1 || result.Content[0].Text != `{"ok":true}` {
		t.Fatalf("compatibility content = %#v", result.Content)
	}
	if result.StructuredContent.(map[string]any)["ok"] != true {
		t.Fatalf("structured content = %#v", result.StructuredContent)
	}
}
