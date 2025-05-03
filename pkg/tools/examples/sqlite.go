package examples

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

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

			dbPath := defaultDBPath
			if filePath, pathOK := arguments["file_path"].(string); pathOK && filePath != "" {
				dbPath = filePath
			}

			// Check if this DB is already open in the session
			if existingPathVal, pathOk := s.GetData(sessionDBPathKey); pathOk {
				if existingPath, ok := existingPathVal.(string); ok && existingPath == dbPath {
					if _, connOk := s.GetData(sessionDBConnectionKey); connOk {
						log.Info().Str("sessionID", string(s.ID)).Str("dbPath", dbPath).Msg("SQLite database already open in session")
						return protocol.NewToolResult(
							protocol.WithText(fmt.Sprintf("Database already open: %s", dbPath)),
						), nil
					}
				}
			}

			// If not already open or different path requested, close existing connection if any
			closeSessionDB(s) // Best effort closing

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

				trimmedQuery := strings.TrimSpace(query)
				upperQuery := strings.ToUpper(trimmedQuery)

				// Check if this is a comment-only query
				if trimmedQuery == "" || strings.HasPrefix(trimmedQuery, "--") {
					processResult := map[string]interface{}{
						"query":  query,
						"status": "comment or empty query - skipped",
					}
					queryLogger.Debug().Msg("Skipping comment or empty query")
					allResults = append(allResults, processResult)
					continue
				}

				// Determine query type
				isSelect := strings.HasPrefix(upperQuery, "SELECT") ||
					strings.HasPrefix(upperQuery, "WITH")
				isDML := strings.HasPrefix(upperQuery, "UPDATE") ||
					strings.HasPrefix(upperQuery, "INSERT") ||
					strings.HasPrefix(upperQuery, "DELETE") ||
					strings.HasPrefix(upperQuery, "REPLACE")
				isAdminCommand := strings.HasPrefix(upperQuery, "CREATE") ||
					strings.HasPrefix(upperQuery, "DROP") ||
					strings.HasPrefix(upperQuery, "PRAGMA") ||
					strings.HasPrefix(upperQuery, "ALTER") ||
					strings.HasPrefix(upperQuery, "TRUNCATE") ||
					strings.HasPrefix(upperQuery, "BEGIN") ||
					strings.HasPrefix(upperQuery, "START TRANSACTION") ||
					strings.HasPrefix(upperQuery, "COMMIT") ||
					strings.HasPrefix(upperQuery, "ROLLBACK") ||
					strings.HasPrefix(upperQuery, "SAVEPOINT") ||
					strings.HasPrefix(upperQuery, "RELEASE") ||
					strings.HasPrefix(upperQuery, "VACUUM") ||
					strings.HasPrefix(upperQuery, "ANALYZE") ||
					strings.HasPrefix(upperQuery, "ATTACH") ||
					strings.HasPrefix(upperQuery, "DETACH") ||
					strings.HasPrefix(upperQuery, "REINDEX") ||
					strings.HasPrefix(upperQuery, "EXPLAIN")

				var processResult map[string]interface{}
				if isSelect {
					// Use QueryContext for SELECT statements
					rows, err := db.QueryContext(ctx, query)
					if err != nil {
						queryErr = errors.Wrapf(err, "error executing SELECT query %d", i)
						queryLogger.Error().
							Err(err).
							Msg("Failed to execute SQLite SELECT query")

						processResult = map[string]interface{}{
							"query": query,
							"error": err.Error(),
						}
					} else {
						queryLogger.Debug().Msg("Successfully executed query, processing results")

						var processErr error
						processResult, processErr = processQueryResults(rows, query)
						_ = rows.Close() // Close rows immediately after processing
						if processErr != nil {
							queryErr = errors.Wrapf(processErr, "error processing results for query %d", i)
							queryLogger.Error().
								Err(processErr).
								Msg("Failed to process query results")

							// Add error information to this query's result map
							processResult["error"] = processErr.Error()
						}
					}
				} else if isAdminCommand || isDML {
					// Use ExecContext for DML and admin commands
					res, err := db.ExecContext(ctx, query)
					if err != nil {
						queryErr = errors.Wrapf(err, "error executing query %d", i)
						queryLogger.Error().
							Err(err).
							Msg("Failed to execute SQLite query")

						processResult = map[string]interface{}{
							"query": query,
							"error": err.Error(),
						}
					} else {
						processResult = map[string]interface{}{
							"query":  query,
							"status": "executed successfully",
						}

						// For DML statements, always try to get rows affected
						if isDML {
							rowsAffected, err := res.RowsAffected()
							if err != nil {
								queryLogger.Error().
									Err(err).
									Msg("Failed to get rows affected")
								processResult["error"] = fmt.Sprintf("query succeeded but failed to get rows affected: %v", err)
							} else {
								processResult["rows_affected"] = rowsAffected
								if rowsAffected == 0 && !strings.Contains(upperQuery, "WHERE 1=0") {
									// If no rows were affected and this wasn't explicitly intended (WHERE 1=0),
									// add a warning
									processResult["warning"] = "query succeeded but no rows were affected"
								}
							}
						}

						lastID, err := res.LastInsertId()
						if err == nil && lastID > 0 {
							processResult["last_insert_id"] = lastID
						}

						queryLogger.Debug().
							Interface("result", processResult).
							Msg("Successfully executed query")
					}
				} else {
					processResult = map[string]interface{}{
						"query":  query,
						"status": "unknown query type",
					}
				}

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
	logger := log.Logger.With().Str("query", query).Logger()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		logger.Error().Err(err).Msg("Failed to get column names")
		return map[string]interface{}{"query": query}, errors.Wrap(err, "error getting columns")
	}

	logger.Debug().Strs("columns", columns).Msg("Retrieved column names")

	// Prepare result storage
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range columns {
		valuePtrs[i] = &values[i]
	}
	if len(columns) == 0 {
		return map[string]interface{}{"query": query}, nil
	}

	// Build results for this query
	var queryResults []map[string]interface{}
	rowNum := 0
	totalBytes := 0

	queryResultMap := map[string]interface{}{
		"query":   query,
		"results": queryResults,
	}

	// Calculate initial size with just the query
	initialBytes, err := yaml.Marshal(queryResultMap)
	if err != nil {
		return queryResultMap, errors.Wrap(err, "error calculating initial size")
	}
	totalBytes = len(initialBytes)

	for rows.Next() {
		rowNum++
		if rowNum > 1000 {
			queryResults = append(queryResults, map[string]interface{}{
				"warning": "Stopping after 1000 rows",
			})
			break
		}
		rowLogger := logger.With().Int("row", rowNum).Logger()

		err := rows.Scan(valuePtrs...)
		if err != nil {
			rowLogger.Error().Err(err).Msg("Failed to scan row")
			return map[string]interface{}{"query": query}, errors.Wrap(err, "error scanning row")
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			// Handle BLOBs/bytes as strings for YAML output
			b, ok := val.([]byte)
			if ok {
				row[col] = string(b)
			} else {
				if val == nil {
					row[col] = nil
				} else {
					row[col] = val
				}
			}
		}

		queryResults = append(queryResults, row)
		queryResultMap["results"] = queryResults

		// Check size after adding new row
		bytes, err := yaml.Marshal(queryResultMap)
		if err != nil {
			rowLogger.Error().Err(err).Msg("Failed to marshal results")
			continue
		}

		totalBytes = len(bytes)
		if totalBytes > 10000 {
			// Remove last row that put us over the limit
			queryResults = queryResults[:len(queryResults)-1]
			queryResults = append(queryResults, map[string]interface{}{
				"warning": "Results truncated due to size limit of 10000 bytes",
			})
			break
		}

		rowLogger.Debug().Interface("row_data", row).Msg("Processed row")
	}

	if err = rows.Err(); err != nil {
		logger.Error().Err(err).Msg("Error occurred while iterating rows")
		return map[string]interface{}{"query": query}, errors.Wrap(err, "error iterating rows")
	}

	logger.Debug().
		Int("total_rows", rowNum).
		Int("total_bytes", totalBytes).
		Msg("Finished processing rows")

	queryResultMap["results"] = queryResults
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
