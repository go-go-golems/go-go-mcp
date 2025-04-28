package examples

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/tools"
	tool_registry "github.com/go-go-golems/go-go-mcp/pkg/tools/providers/tool-registry"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/yaml.v3"
)

const defaultDBPath = "/home/manuel/.config/Cursor/User/globalStorage/state.vscdb"

func RegisterSQLiteTool(registry *tool_registry.Registry) error {
	schemaJson := `{
		"type": "object",
		"properties": {
			"queries": {
				"type": "array",
				"items": {
					"type": "string"
				},
				"description": "A list of SQL queries to execute sequentially against Cursor's SQLite database. The database contains AI conversation history, IDE state, and other Cursor-related data. Common tables include 'cursorDiskKV' for conversation storage. Note: SQLite dot commands like .tables are not supported as they are CLI-specific features. Use standard SQL queries instead."
			},
			"file_path": {
				"type": "string",
				"description": "Path to the SQLite database file. If not provided, the default Cursor database path will be used."
			}
		},
		"required": ["queries"]
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
			queries, ok := arguments["queries"].([]interface{})
			if !ok {
				return protocol.NewToolResult(
					protocol.WithError("queries argument must be an array of strings"),
				), nil
			}

			dbPath := defaultDBPath
			if filePath, ok := arguments["file_path"].(string); ok && filePath != "" {
				dbPath = filePath
			}

			db, err := sql.Open("sqlite3", dbPath)
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error opening database: %v", err)),
				), nil
			}
			defer func() {
				if closeErr := db.Close(); closeErr != nil {
					if err == nil {
						err = closeErr
					}
				}
			}()

			// Store results for each query
			allResults := make([]map[string]interface{}, 0)

			// Execute each query sequentially
			for i, q := range queries {
				query, ok := q.(string)
				if !ok {
					return protocol.NewToolResult(
						protocol.WithError(fmt.Sprintf("query at index %d is not a string", i)),
					), nil
				}

				rows, err := db.QueryContext(ctx, query)
				if err != nil {
					return protocol.NewToolResult(
						protocol.WithError(fmt.Sprintf("error executing query %d: %v", i, err)),
					), nil
				}

				// Get column names
				columns, err := rows.Columns()
				if err != nil {
					_ = rows.Close()
					return protocol.NewToolResult(
						protocol.WithError(fmt.Sprintf("error getting columns for query %d: %v", i, err)),
					), nil
				}

				// Prepare result storage
				values := make([]interface{}, len(columns))
				valuePtrs := make([]interface{}, len(columns))
				for i := range columns {
					valuePtrs[i] = &values[i]
				}

				// Build results for this query
				var queryResults []map[string]interface{}
				for rows.Next() {
					err := rows.Scan(valuePtrs...)
					if err != nil {
						_ = rows.Close()
						return protocol.NewToolResult(
							protocol.WithError(fmt.Sprintf("error scanning row for query %d: %v", i, err)),
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
					queryResults = append(queryResults, row)
				}

				if err = rows.Err(); err != nil {
					_ = rows.Close()
					return protocol.NewToolResult(
						protocol.WithError(fmt.Sprintf("error iterating rows for query %d: %v", i, err)),
					), nil
				}
				_ = rows.Close()

				// Add query results to the overall results
				queryResult := map[string]interface{}{
					"query":   query,
					"results": queryResults,
				}
				allResults = append(allResults, queryResult)
			}

			// Convert results to YAML
			yamlData, err := yaml.Marshal(allResults)
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
