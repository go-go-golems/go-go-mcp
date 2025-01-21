package cursor

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/tools"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/yaml.v3"
)

func RegisterGetFileReferencesTool(registry *tools.Registry) error {
	schemaJson := `{
		"type": "object",
		"properties": {
			"conversation_id": {
				"type": "string",
				"description": "The conversation ID to get file references for"
			}
		},
		"required": ["conversation_id"]
	}`

	tool, err := tools.NewToolImpl("cursor-get-file-references", "Get all files referenced in a conversation", json.RawMessage(schemaJson))
	if err != nil {
		return err
	}

	registry.RegisterToolWithHandler(
		tool,
		func(ctx context.Context, tool tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
			conversationID, ok := arguments["conversation_id"].(string)
			if !ok {
				return protocol.NewToolResult(
					protocol.WithError("conversation_id argument must be a string"),
				), nil
			}

			db, err := sql.Open("sqlite3", defaultDBPath)
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error opening database: %v", err)),
				), nil
			}
			defer db.Close()

			query := `
				SELECT key, 
					json_extract(value, '$.conversation[0].context.fileSelections') as files
				FROM cursorDiskKV
				WHERE key = ? AND json_valid(value)
			`

			var key string
			var files string
			err = db.QueryRowContext(ctx, query, conversationID).Scan(&key, &files)
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error executing query: %v", err)),
				), nil
			}

			var fileList []interface{}
			if err := json.Unmarshal([]byte(files), &fileList); err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error parsing JSON: %v", err)),
				), nil
			}

			yamlData, err := yaml.Marshal(map[string]interface{}{
				"key":   key,
				"files": fileList,
			})
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error converting to YAML: %v", err)),
				), nil
			}

			return protocol.NewToolResult(
				protocol.WithText(string(yamlData)),
			), nil
		})

	return nil
}

func RegisterGetConversationContextTool(registry *tools.Registry) error {
	schemaJson := `{
		"type": "object",
		"properties": {
			"conversation_id": {
				"type": "string",
				"description": "The conversation ID to get context for"
			}
		},
		"required": ["conversation_id"]
	}`

	tool, err := tools.NewToolImpl("cursor-get-conversation-context", "Get full context including files and mentions", json.RawMessage(schemaJson))
	if err != nil {
		return err
	}

	registry.RegisterToolWithHandler(
		tool,
		func(ctx context.Context, tool tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
			conversationID, ok := arguments["conversation_id"].(string)
			if !ok {
				return protocol.NewToolResult(
					protocol.WithError("conversation_id argument must be a string"),
				), nil
			}

			db, err := sql.Open("sqlite3", defaultDBPath)
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error opening database: %v", err)),
				), nil
			}
			defer db.Close()

			query := `
				SELECT key, 
					json_extract(value, '$.conversation[0].context') as context
				FROM cursorDiskKV
				WHERE key = ? AND json_valid(value)
			`

			var key string
			var contextData string
			err = db.QueryRowContext(ctx, query, conversationID).Scan(&key, &contextData)
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error executing query: %v", err)),
				), nil
			}

			var context interface{}
			if err := json.Unmarshal([]byte(contextData), &context); err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error parsing JSON: %v", err)),
				), nil
			}

			yamlData, err := yaml.Marshal(map[string]interface{}{
				"key":     key,
				"context": context,
			})
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error converting to YAML: %v", err)),
				), nil
			}

			return protocol.NewToolResult(
				protocol.WithText(string(yamlData)),
			), nil
		})

	return nil
}
