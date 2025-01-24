package cursor

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/tools"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/yaml.v3"
)

func RegisterExtractCodeBlocksTool(registry *tools.Registry) error {
	schemaJson := `{
		"type": "object",
		"properties": {
			"filepaths": {
				"type": "array",
				"items": {
					"type": "string"
				},
				"description": "List of file paths or search terms to extract code blocks for. Each term can be a file path (relative or absolute) or a search term. Terms containing spaces should be wrapped in quotes. Example: ['file.go', 'other.go', 'some search term']. Multiple unquoted words are treated as separate search terms."
			}
		},
		"required": ["filepaths"]
	}`

	tool, err := tools.NewToolImpl(
		"cursor-extract-code-blocks",
		"Extract all code blocks that were generated or modified by the AI for multiple files or search terms. Returns a structured YAML output containing code blocks organized by file and conversation. Accepts space-separated terms (use quotes for terms containing spaces) for flexible searching. Example: file.go 'quoted term' other.go",
		json.RawMessage(schemaJson))
	if err != nil {
		return err
	}

	registry.RegisterToolWithHandler(
		tool,
		func(ctx context.Context, tool tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
			filepaths, ok := arguments["filepaths"].([]interface{})
			if !ok {
				return protocol.NewToolResult(
					protocol.WithError("filepaths argument must be an array of strings"),
				), nil
			}

			// Process each filepath/search term
			var searchTerms []string
			for _, path := range filepaths {
				pathStr := fmt.Sprintf("%v", path)
				terms := splitSearchTerms(pathStr)
				searchTerms = append(searchTerms, terms...)
			}

			db, err := sql.Open("sqlite3", defaultDBPath)
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error opening database: %v", err)),
				), nil
			}
			defer db.Close()

			// Build LIKE conditions for each search term
			conditions := make([]string, len(searchTerms))
			args := make([]interface{}, len(searchTerms))
			for i, term := range searchTerms {
				conditions[i] = "value LIKE ?"
				args[i] = "%" + term + "%"
			}

			query := fmt.Sprintf(`
				SELECT key, 
					json_extract(value, '$.conversation[1].codeBlocks') as code_blocks,
					value as full_value
				FROM cursorDiskKV
				WHERE %s
			`, strings.Join(conditions, " OR "))

			rows, err := db.QueryContext(ctx, query, args...)
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error executing query: %v", err)),
				), nil
			}
			defer rows.Close()

			results := make(map[string][]map[string]interface{})
			for rows.Next() {
				var key string
				var codeBlocks string
				var fullValue string
				if err := rows.Scan(&key, &codeBlocks, &fullValue); err != nil {
					return protocol.NewToolResult(
						protocol.WithError(fmt.Sprintf("error scanning row: %v", err)),
					), nil
				}

				var blocks []interface{}
				if err := json.Unmarshal([]byte(codeBlocks), &blocks); err != nil {
					continue // Skip invalid JSON
				}

				// Match against each search term
				for _, term := range searchTerms {
					if strings.Contains(fullValue, term) {
						if _, ok := results[term]; !ok {
							results[term] = []map[string]interface{}{}
						}
						results[term] = append(results[term], map[string]interface{}{
							"key":         key,
							"code_blocks": blocks,
						})
					}
				}
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
			"filepaths": {
				"type": "array",
				"items": {
					"type": "string"
				},
				"description": "List of file paths or search terms to track modifications for. Each term can be a file path (relative or absolute) or a search term. Terms containing spaces should be wrapped in quotes. Example: ['file.go', 'other.go', 'some search term']. Multiple unquoted words are treated as separate search terms."
			}
		},
		"required": ["filepaths"]
	}`

	tool, err := tools.NewToolImpl(
		"cursor-track-file-modifications",
		"Track and analyze all modifications made to files matching the specified search terms through Cursor AI interactions over time. Returns a chronological history of changes for each matching file or term. Accepts space-separated terms (use quotes for terms containing spaces) for flexible searching. Example: file.go 'quoted term' other.go",
		json.RawMessage(schemaJson))
	if err != nil {
		return err
	}

	registry.RegisterToolWithHandler(
		tool,
		func(ctx context.Context, tool tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
			filepaths, ok := arguments["filepaths"].([]interface{})
			if !ok {
				return protocol.NewToolResult(
					protocol.WithError("filepaths argument must be an array of strings"),
				), nil
			}

			// Process each filepath/search term
			var searchTerms []string
			for _, path := range filepaths {
				pathStr := fmt.Sprintf("%v", path)
				terms := splitSearchTerms(pathStr)
				searchTerms = append(searchTerms, terms...)
			}

			db, err := sql.Open("sqlite3", defaultDBPath)
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error opening database: %v", err)),
				), nil
			}
			defer db.Close()

			// Build LIKE conditions for each search term
			conditions := make([]string, len(searchTerms))
			args := make([]interface{}, len(searchTerms))
			for i, term := range searchTerms {
				conditions[i] = "value LIKE ?"
				args[i] = "%" + term + "%"
			}

			query := fmt.Sprintf(`
				WITH RECURSIVE 
				file_mods AS (
					SELECT key,
						json_extract(value, '$.createdAt') as created_at,
						json_extract(value, '$.conversation[1].codeBlocks') as code_blocks,
						value as full_value
					FROM cursorDiskKV
					WHERE %s
					ORDER BY created_at ASC
				)
				SELECT * FROM file_mods
			`, strings.Join(conditions, " OR "))

			rows, err := db.QueryContext(ctx, query, args...)
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error executing query: %v", err)),
				), nil
			}
			defer rows.Close()

			results := make(map[string][]map[string]interface{})
			for rows.Next() {
				var key string
				var createdAt string
				var codeBlocks string
				var fullValue string
				if err := rows.Scan(&key, &createdAt, &codeBlocks, &fullValue); err != nil {
					return protocol.NewToolResult(
						protocol.WithError(fmt.Sprintf("error scanning row: %v", err)),
					), nil
				}

				var blocks []interface{}
				if err := json.Unmarshal([]byte(codeBlocks), &blocks); err != nil {
					continue // Skip invalid JSON
				}

				// Match against each search term
				for _, term := range searchTerms {
					if strings.Contains(fullValue, term) {
						if _, ok := results[term]; !ok {
							results[term] = []map[string]interface{}{}
						}
						results[term] = append(results[term], map[string]interface{}{
							"key":         key,
							"created_at":  createdAt,
							"code_blocks": blocks,
						})
					}
				}
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
