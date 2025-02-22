package examples

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/tools"
	"github.com/go-go-golems/go-go-mcp/pkg/tools/providers/tool-registry"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/yaml.v3"
)

const defaultDBPath = "/home/manuel/.config/Cursor/User/globalStorage/state.vscdb"

func RegisterSQLiteTool(registry *tool_registry.Registry) error {
	schemaJson := `{
		"type": "object",
		"properties": {
			"query": {
				"type": "string",
				"description": "An SQL query to execute against Cursor's SQLite database. The database contains AI conversation history, IDE state, and other Cursor-related data. Common tables include 'cursorDiskKV' for conversation storage. Note: SQLite dot commands like .tables are not supported as they are CLI-specific features. Use standard SQL queries instead."
			}
		},
		"required": ["query"]
	}`

	tool, err := tools.NewToolImpl(
		"sqlite",
		"Execute SQL queries against the Cursor IDE's SQLite database and output results as YAML. This tool provides direct access to Cursor's underlying data storage, allowing complex queries for conversation analysis, usage patterns, and IDE state. Useful for advanced data analysis or when the higher-level conversation tools are insufficient.",
		json.RawMessage(schemaJson))
	if err != nil {
		return err
	}

	registry.RegisterToolWithHandler(
		tool,
		func(ctx context.Context, tool tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
			query, ok := arguments["query"].(string)
			if !ok {
				return protocol.NewToolResult(
					protocol.WithError("query argument must be a string"),
				), nil
			}

			db, err := sql.Open("sqlite3", defaultDBPath)
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error opening database: %v", err)),
				), nil
			}
			defer db.Close()

			rows, err := db.QueryContext(ctx, query)
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error executing query: %v", err)),
				), nil
			}
			defer rows.Close()

			// Get column names
			columns, err := rows.Columns()
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error getting columns: %v", err)),
				), nil
			}

			// Prepare result storage
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range columns {
				valuePtrs[i] = &values[i]
			}

			// Build results
			var results []map[string]interface{}
			for rows.Next() {
				err := rows.Scan(valuePtrs...)
				if err != nil {
					return protocol.NewToolResult(
						protocol.WithError(fmt.Sprintf("error scanning row: %v", err)),
					), nil
				}

				row := make(map[string]interface{})
				for i, col := range columns {
					val := values[i]
					b, ok := val.([]byte)
					if ok {
						row[col] = string(b)
					} else {
						row[col] = val
					}
				}
				results = append(results, row)
			}

			if err = rows.Err(); err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error iterating rows: %v", err)),
				), nil
			}

			// Convert results to YAML
			yamlData, err := yaml.Marshal(results)
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error converting results to YAML: %v", err)),
				), nil
			}

			return protocol.NewToolResult(
				protocol.WithText(string(yamlData)),
			), nil
		})

	return nil
}
