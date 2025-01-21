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

const defaultDBPath = "/home/manuel/.config/Cursor/User/globalStorage/state.vscdb"

func RegisterGetConversationTool(registry *tools.Registry) error {
	schemaJson := `{
		"type": "object",
		"properties": {
			"composer_id": {
				"type": "string",
				"description": "The composer ID to retrieve the conversation for"
			}
		},
		"required": ["composer_id"]
	}`

	tool, err := tools.NewToolImpl("cursor-get-conversation", "Retrieve full conversation by composer ID", json.RawMessage(schemaJson))
	if err != nil {
		return err
	}

	registry.RegisterToolWithHandler(
		tool,
		func(ctx context.Context, tool tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
			composerID, ok := arguments["composer_id"].(string)
			if !ok {
				return protocol.NewToolResult(
					protocol.WithError("composer_id argument must be a string"),
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
				SELECT value 
				FROM cursorDiskKV 
				WHERE json_extract(value, '$.composerId') = ?
			`

			var value string
			err = db.QueryRowContext(ctx, query, composerID).Scan(&value)
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error executing query: %v", err)),
				), nil
			}

			// Convert to YAML for better readability
			var jsonData interface{}
			if err := json.Unmarshal([]byte(value), &jsonData); err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error parsing JSON: %v", err)),
				), nil
			}

			yamlData, err := yaml.Marshal(jsonData)
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

func RegisterFindConversationsTool(registry *tools.Registry) error {
	schemaJson := `{
		"type": "object",
		"properties": {
			"search_term": {
				"type": "string",
				"description": "Term to search for in conversations"
			}
		},
		"required": ["search_term"]
	}`

	tool, err := tools.NewToolImpl("cursor-find-conversations", "Search conversations by content", json.RawMessage(schemaJson))
	if err != nil {
		return err
	}

	registry.RegisterToolWithHandler(
		tool,
		func(ctx context.Context, tool tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
			searchTerm, ok := arguments["search_term"].(string)
			if !ok {
				return protocol.NewToolResult(
					protocol.WithError("search_term argument must be a string"),
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
					json_extract(value, '$.conversation[0].text') as user_message,
					json_extract(value, '$.conversation[1].text') as assistant_response
				FROM cursorDiskKV 
				WHERE value LIKE ?
			`

			rows, err := db.QueryContext(ctx, query, "%"+searchTerm+"%")
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error executing query: %v", err)),
				), nil
			}
			defer rows.Close()

			var results []map[string]interface{}
			for rows.Next() {
				var key, userMsg, assistantMsg string
				if err := rows.Scan(&key, &userMsg, &assistantMsg); err != nil {
					return protocol.NewToolResult(
						protocol.WithError(fmt.Sprintf("error scanning row: %v", err)),
					), nil
				}

				results = append(results, map[string]interface{}{
					"key":                key,
					"user_message":       userMsg,
					"assistant_response": assistantMsg,
				})
			}

			yamlData, err := yaml.Marshal(results)
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
