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

func RegisterExtractCodeBlocksTool(registry *tools.Registry) error {
	schemaJson := `{
		"type": "object",
		"properties": {
			"filepath": {
				"type": "string",
				"description": "The filepath to extract code blocks for"
			}
		},
		"required": ["filepath"]
	}`

	tool, err := tools.NewToolImpl("cursor-extract-code-blocks", "Get all code blocks for a file", json.RawMessage(schemaJson))
	if err != nil {
		return err
	}

	registry.RegisterToolWithHandler(
		tool,
		func(ctx context.Context, tool tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
			filepath, ok := arguments["filepath"].(string)
			if !ok {
				return protocol.NewToolResult(
					protocol.WithError("filepath argument must be a string"),
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
					json_extract(value, '$.conversation[1].codeBlocks') as code_blocks
				FROM cursorDiskKV
				WHERE value LIKE ?
			`

			rows, err := db.QueryContext(ctx, query, "%"+filepath+"%")
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error executing query: %v", err)),
				), nil
			}
			defer rows.Close()

			var results []map[string]interface{}
			for rows.Next() {
				var key string
				var codeBlocks string
				if err := rows.Scan(&key, &codeBlocks); err != nil {
					return protocol.NewToolResult(
						protocol.WithError(fmt.Sprintf("error scanning row: %v", err)),
					), nil
				}

				var blocks []interface{}
				if err := json.Unmarshal([]byte(codeBlocks), &blocks); err != nil {
					continue // Skip invalid JSON
				}

				results = append(results, map[string]interface{}{
					"key":         key,
					"code_blocks": blocks,
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

func RegisterTrackFileModificationsTool(registry *tools.Registry) error {
	schemaJson := `{
		"type": "object",
		"properties": {
			"filepath": {
				"type": "string",
				"description": "The filepath to track modifications for"
			}
		},
		"required": ["filepath"]
	}`

	tool, err := tools.NewToolImpl("cursor-track-file-modifications", "Track changes to a file over time", json.RawMessage(schemaJson))
	if err != nil {
		return err
	}

	registry.RegisterToolWithHandler(
		tool,
		func(ctx context.Context, tool tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
			filepath, ok := arguments["filepath"].(string)
			if !ok {
				return protocol.NewToolResult(
					protocol.WithError("filepath argument must be a string"),
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
				WITH RECURSIVE 
				file_mods AS (
					SELECT key,
						json_extract(value, '$.createdAt') as created_at,
						json_extract(value, '$.conversation[1].codeBlocks') as code_blocks
					FROM cursorDiskKV
					WHERE value LIKE ?
					ORDER BY created_at ASC
				)
				SELECT * FROM file_mods
			`

			rows, err := db.QueryContext(ctx, query, "%"+filepath+"%")
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error executing query: %v", err)),
				), nil
			}
			defer rows.Close()

			var results []map[string]interface{}
			for rows.Next() {
				var key string
				var createdAt string
				var codeBlocks string
				if err := rows.Scan(&key, &createdAt, &codeBlocks); err != nil {
					return protocol.NewToolResult(
						protocol.WithError(fmt.Sprintf("error scanning row: %v", err)),
					), nil
				}

				var blocks []interface{}
				if err := json.Unmarshal([]byte(codeBlocks), &blocks); err != nil {
					continue // Skip invalid JSON
				}

				results = append(results, map[string]interface{}{
					"key":         key,
					"created_at":  createdAt,
					"code_blocks": blocks,
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
