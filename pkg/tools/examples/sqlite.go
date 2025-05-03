package examples

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/session"
	"github.com/go-go-golems/go-go-mcp/pkg/tools"
	tool_registry "github.com/go-go-golems/go-go-mcp/pkg/tools/providers/tool-registry"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

const (
	defaultDBPath          = "/home/manuel/.config/Cursor/User/globalStorage/state.vscdb"
	sessionDBConnectionKey = "sqlite_db_connection"
	sessionDBPathKey       = "sqlite_db_path"
)

// RegisterSQLiteSessionTools registers three tools: sqlite_open, sqlite_query, and sqlite_close.
// sqlite_open: Opens a SQLite database and stores the connection in the session.
// sqlite_query: Executes queries against the session's open database or a default one.
// sqlite_close: Closes the SQLite database connection stored in the session.
func RegisterSQLiteSessionTools(registry *tool_registry.Registry) error {
	if err := registerSQLiteOpenTool(registry); err != nil {
		return errors.Wrap(err, "failed to register sqlite_open tool")
	}
	if err := registerSQLiteQueryTool(registry); err != nil {
		return errors.Wrap(err, "failed to register sqlite_query tool")
	}
	return nil
}

// registerSQLiteOpenTool registers the tool to open a database connection.
func registerSQLiteOpenTool(registry *tool_registry.Registry) error {
	schemaJson := `{
		"type": "object",
		"properties": {
			"file_path": {
				"type": "string",
				"description": "Path to the SQLite database file to open. If not provided, the default Cursor database path will be used."
			}
		},
		"required": []
	}`

	tool, err := tools.NewToolImpl(
		"sqlite_open",
		"Opens a SQLite database connection and stores it in the current session. Subsequent 'sqlite_query' calls will use this connection.",
		json.RawMessage(schemaJson))
	if err != nil {
		return err
	}

	registry.RegisterToolWithHandler(
		tool,
		func(ctx context.Context, _ tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
			s, ok := session.GetSessionFromContext(ctx)
			if !ok {
				return protocol.NewToolResult(protocol.WithError("no session found in context")), nil
			}

			// Close existing connection if any
			closeSessionDB(s) // Best effort closing

			dbPath := defaultDBPath
			if filePath, pathOK := arguments["file_path"].(string); pathOK && filePath != "" {
				dbPath = filePath
			}

			db, err := sql.Open("sqlite3", dbPath)
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error opening database '%s': %v", dbPath, err)),
				), nil
			}

			// Test connection
			if err := db.PingContext(ctx); err != nil {
				_ = db.Close()
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error pinging database '%s': %v", dbPath, err)),
				), nil
			}

			// Store connection and path in session
			s.SetData(sessionDBConnectionKey, db)
			s.SetData(sessionDBPathKey, dbPath)

			log.Info().Str("sessionID", string(s.ID)).Str("dbPath", dbPath).Msg("SQLite database opened and stored in session")

			return protocol.NewToolResult(
				protocol.WithText(fmt.Sprintf("Successfully opened database: %s", dbPath)),
			), nil
		})

	return nil
}

// closeSessionDB closes the database connection stored in the session, if it exists.
// It returns the path of the closed DB and true if a connection was found and closed, otherwise empty string and false.
func closeSessionDB(s *session.Session) (string, bool) {
	dbPath := ""
	pathVal, pathOk := s.GetData(sessionDBPathKey)
	if pathOk {
		dbPath = pathVal.(string)
	}

	connVal, ok := s.GetData(sessionDBConnectionKey)
	if !ok {
		log.Debug().Str("sessionID", string(s.ID)).Msg("No DB connection found in session to close")
		return "", false // No connection to close
	}

	db, ok := connVal.(*sql.DB)
	if !ok {
		log.Warn().Str("sessionID", string(s.ID)).Msg("Session data for DB connection is not *sql.DB")
		// Attempt to remove potentially corrupted data
		s.DeleteData(sessionDBConnectionKey)
		s.DeleteData(sessionDBPathKey)
		return "", false
	}

	err := db.Close()
	s.DeleteData(sessionDBConnectionKey) // Remove even if close failed
	s.DeleteData(sessionDBPathKey)

	if err != nil {
		log.Error().Err(err).Str("sessionID", string(s.ID)).Str("dbPath", dbPath).Msg("Error closing session SQLite database")
		// We still report closed=true because we removed the entry
	} else {
		log.Debug().Str("sessionID", string(s.ID)).Str("dbPath", dbPath).Msg("Closed session SQLite database")
	}
	return dbPath, true
}

// registerSQLiteQueryTool registers the tool to execute SQL queries.
func registerSQLiteQueryTool(registry *tool_registry.Registry) error {
	schemaJson := `{
		"type": "object",
		"properties": {
			"queries": {
				"type": "array",
				"items": { "type": "string" },
				"description": "A list of SQL queries to execute. Uses the connection opened by 'sqlite_open' if available, otherwise uses the default Cursor DB."
			},
			"file_path": {
				"type": "string",
				"description": "Path to the SQLite database file. Only used if no connection is open in the session. If neither session connection nor file_path is provided, the default Cursor database path is used."
			}
		},
		"required": ["queries"]
	}`

	tool, err := tools.NewToolImpl(
		"sqlite_query",
		"Execute SQL queries against the session's open SQLite database or a specified/default one, outputting results as YAML.",
		json.RawMessage(schemaJson))
	if err != nil {
		return err
	}

	registry.RegisterToolWithHandler(
		tool,
		func(ctx context.Context, _ tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
			queriesArg, ok := arguments["queries"]
			if !ok {
				return protocol.NewToolResult(protocol.WithError("missing required argument 'queries'")), nil
			}
			queriesInterfaces, ok := queriesArg.([]interface{})
			if !ok {
				return protocol.NewToolResult(protocol.WithError("'queries' argument must be an array of strings")), nil
			}
			if len(queriesInterfaces) == 0 {
				return protocol.NewToolResult(protocol.WithError("'queries' array must not be empty")), nil
			}

			queries := make([]string, len(queriesInterfaces))
			for i, q := range queriesInterfaces {
				queryStr, ok := q.(string)
				if !ok {
					return protocol.NewToolResult(protocol.WithError(fmt.Sprintf("query at index %d is not a string", i))), nil
				}
				queries[i] = queryStr
			}

			var db *sql.DB
			var dbPath string
			var closeDbFunc func() // Function to close the DB if opened temporarily
			var sessionUsed bool

			s, sessionOk := session.GetSessionFromContext(ctx)
			if sessionOk {
				if connVal, connOk := s.GetData(sessionDBConnectionKey); connOk {
					var dbOk bool
					db, dbOk = connVal.(*sql.DB)
					if !dbOk {
						log.Warn().Str("sessionID", string(s.ID)).Msg("Session data for DB connection is not *sql.DB, ignoring")
						// Proceed as if no session connection exists
					} else {
						sessionUsed = true
						if pathVal, pathOk := s.GetData(sessionDBPathKey); pathOk {
							dbPath = pathVal.(string)
						} else {
							dbPath = "[unknown session path]"
						}
						log.Debug().Str("sessionID", string(s.ID)).Str("dbPath", dbPath).Msg("Using SQLite connection from session")
					}
				}
			}

			// If no valid session DB, try opening one based on file_path or default
			if !sessionUsed {
				dbPath = defaultDBPath
				if filePathArg, ok := arguments["file_path"].(string); ok && filePathArg != "" {
					dbPath = filePathArg
				}
				log.Debug().Str("dbPath", dbPath).Msg("Opening temporary SQLite connection")

				var openErr error
				db, openErr = sql.Open("sqlite3", dbPath)
				if openErr != nil {
					return protocol.NewToolResult(
						protocol.WithError(fmt.Sprintf("error opening database '%s': %v", dbPath, openErr)),
					), nil
				}
				closeDbFunc = func() {
					if err := db.Close(); err != nil {
						log.Error().Err(err).Str("dbPath", dbPath).Msg("Error closing temporary SQLite database")
					} else {
						log.Debug().Str("dbPath", dbPath).Msg("Closed temporary SQLite database")
					}
				}
				// Test connection
				if err := db.PingContext(ctx); err != nil {
					closeDbFunc()
					return protocol.NewToolResult(
						protocol.WithError(fmt.Sprintf("error pinging database '%s': %v", dbPath, err)),
					), nil
				}
			}

			// Ensure temporary DB is closed if opened
			if closeDbFunc != nil {
				defer closeDbFunc()
			}

			// Execute queries
			allResults := make([]map[string]interface{}, 0, len(queries))
			var queryErr error // To store the first error encountered

			for i, query := range queries {
				queryLogger := log.With().Int("query_index", i).Str("query", query).Logger()

				queryLogger.Debug().Msg("Executing SQLite query")

				if queryErr != nil {
					queryLogger.Warn().
						Str("previous_error", queryErr.Error()).
						Msg("Skipping query due to previous error")

					allResults = append(allResults, map[string]interface{}{
						"query": query,
						"error": "Skipped due to previous error",
					})
					continue
				}

				rows, err := db.QueryContext(ctx, query)
				if err != nil {
					queryErr = errors.Wrapf(err, "error executing query %d", i)
					queryLogger.Error().
						Err(err).
						Msg("Failed to execute SQLite query")

					allResults = append(allResults, map[string]interface{}{
						"query": query,
						"error": err.Error(),
					})
					continue // Move to the next query
				}

				queryLogger.Debug().Msg("Successfully executed query, processing results")

				processResult, err := processQueryResults(rows, query)
				_ = rows.Close() // Close rows immediately after processing
				if err != nil {
					queryErr = errors.Wrapf(err, "error processing results for query %d", i)
					queryLogger.Error().
						Err(err).
						Msg("Failed to process query results")

					// Add error information to this query's result map
					processResult["error"] = err.Error()
					allResults = append(allResults, processResult)
					continue // Move to the next query
				}

				queryLogger.Debug().
					Interface("result", processResult).
					Msg("Successfully processed query results")

				allResults = append(allResults, processResult)
			}

			// Convert results to YAML
			yamlData, err := yaml.Marshal(allResults)
			if err != nil {
				// This error happens after query execution, return results error
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error converting results to YAML: %v. Query execution status: %v", err, queryErr)),
				), nil
			}

			// If there was a query error, return it in the protocol error field
			if queryErr != nil {
				return protocol.NewToolResult(
					protocol.WithText(string(yamlData)),
					protocol.WithError(queryErr.Error()),
				), nil
			}

			return protocol.NewToolResult(
				protocol.WithText(string(yamlData)),
			), nil
		})

	return nil
}

// processQueryResults converts sql.Rows into a map structure suitable for YAML marshalling.
func processQueryResults(rows *sql.Rows, query string) (map[string]interface{}, error) {
	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return map[string]interface{}{"query": query}, errors.Wrap(err, "error getting columns")
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
			return map[string]interface{}{"query": query}, errors.Wrap(err, "error scanning row")
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			// Handle BLOBs/bytes as strings for YAML output, adjust if needed
			b, ok := val.([]byte)
			if ok {
				row[col] = string(b)
			} else {
				// Handle potential nil values explicitly if necessary
				if val == nil {
					row[col] = nil // Or "" or some placeholder
				} else {
					row[col] = val
				}
			}
		}
		queryResults = append(queryResults, row)
	}

	if err = rows.Err(); err != nil {
		return map[string]interface{}{"query": query}, errors.Wrap(err, "error iterating rows")
	}

	// Add query results to the overall results
	queryResultMap := map[string]interface{}{
		"query":   query,
		"results": queryResults,
	}
	return queryResultMap, nil
}

// TODO: Remove the old registration function or rename it if needed elsewhere
/*
func RegisterSQLiteTool(registry *tool_registry.Registry) error {
	// ... old implementation ...
}
*/

// Ensure Tool interfaces are implemented (optional, but good practice)
var _ tools.Tool = (*tools.ToolImpl)(nil) // Assuming ToolImpl is the concrete type returned by NewToolImpl
// We don't have direct access to the handler type, so this check is less direct
// var _ tool_registry.ToolHandler = func(ctx context.Context, tool tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) { return nil, nil }
