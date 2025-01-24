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

const defaultDBPath = "/home/manuel/.config/Cursor/User/globalStorage/state.vscdb"

func RegisterGetConversationTool(registry *tools.Registry) error {
	schemaJson := `{
		"type": "object",
		"properties": {
			"composer_ids": {
				"type": "array",
				"items": {
					"type": "string"
				},
				"description": "List of composer IDs or search terms to find conversations. Each term can be a specific composer ID or a search term. Terms containing spaces should be wrapped in quotes. Example: ['abc123', 'def456', 'search term']. Multiple unquoted words are treated as separate search terms."
			}
		},
		"required": ["composer_ids"]
	}`

	tool, err := tools.NewToolImpl(
		"cursor-get-conversation",
		"Retrieve the complete conversation history between users and AI assistants in the Cursor IDE. Returns the full conversations including all messages, code snippets, and metadata. Accepts space-separated terms (use quotes for terms containing spaces) for flexible searching. Example: abc123 'search term' def456",
		json.RawMessage(schemaJson))
	if err != nil {
		return err
	}

	registry.RegisterToolWithHandler(
		tool,
		func(ctx context.Context, tool tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
			composerIDs, ok := arguments["composer_ids"].([]interface{})
			if !ok {
				return protocol.NewToolResult(
					protocol.WithError("composer_ids argument must be an array of strings"),
				), nil
			}

			// Process each composer ID/search term
			var searchTerms []string
			for _, id := range composerIDs {
				idStr := fmt.Sprintf("%v", id)
				terms := splitSearchTerms(idStr)
				searchTerms = append(searchTerms, terms...)
			}

			db, err := sql.Open("sqlite3", defaultDBPath)
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error opening database: %v", err)),
				), nil
			}
			defer db.Close()

			// Build conditions for each search term
			conditions := make([]string, len(searchTerms))
			args := make([]interface{}, len(searchTerms))
			for i, term := range searchTerms {
				conditions[i] = "(json_extract(value, '$.composerId') = ? OR value LIKE ?)"
				args[i*2] = term
				args[i*2+1] = "%" + term + "%"
			}

			query := fmt.Sprintf(`
				SELECT value 
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

			// Group results by matching search term
			results := make(map[string][]interface{})
			for rows.Next() {
				var value string
				if err := rows.Scan(&value); err != nil {
					return protocol.NewToolResult(
						protocol.WithError(fmt.Sprintf("error scanning row: %v", err)),
					), nil
				}

				var jsonData interface{}
				if err := json.Unmarshal([]byte(value), &jsonData); err != nil {
					continue // Skip invalid JSON
				}

				// Match against each search term
				for _, term := range searchTerms {
					if strings.Contains(value, term) {
						if _, ok := results[term]; !ok {
							results[term] = []interface{}{}
						}
						results[term] = append(results[term], jsonData)
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

func RegisterFindConversationsTool(registry *tools.Registry) error {
	schemaJson := `{
		"type": "object",
		"properties": {
			"search_terms": {
				"type": "array",
				"items": {
					"type": "string"
				},
				"description": "List of terms to search for in conversations. Each term can be a simple word or a quoted phrase. Terms containing spaces should be wrapped in quotes. Example: ['error', 'file.go', 'database query']. Multiple unquoted words are treated as separate search terms."
			}
		},
		"required": ["search_terms"]
	}`

	tool, err := tools.NewToolImpl(
		"cursor-find-conversations",
		"Search through all AI conversations stored in the Cursor IDE database. Returns matching conversations with their composer IDs, initial user messages, and AI responses. Accepts space-separated terms (use quotes for terms containing spaces) for flexible searching. Example: error 'database query' file.go",
		json.RawMessage(schemaJson))
	if err != nil {
		return err
	}

	registry.RegisterToolWithHandler(
		tool,
		func(ctx context.Context, tool tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
			searchTermsInput, ok := arguments["search_terms"].([]interface{})
			if !ok {
				return protocol.NewToolResult(
					protocol.WithError("search_terms argument must be an array of strings"),
				), nil
			}

			// Process each search term
			var searchTerms []string
			for _, term := range searchTermsInput {
				termStr := fmt.Sprintf("%v", term)
				terms := splitSearchTerms(termStr)
				searchTerms = append(searchTerms, terms...)
			}

			db, err := sql.Open("sqlite3", defaultDBPath)
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error opening database: %v", err)),
				), nil
			}
			defer db.Close()

			// Build conditions for each search term
			conditions := make([]string, len(searchTerms))
			args := make([]interface{}, len(searchTerms))
			for i, term := range searchTerms {
				conditions[i] = "value LIKE ?"
				args[i] = "%" + term + "%"
			}

			query := fmt.Sprintf(`
				SELECT key, 
					json_extract(value, '$.conversation[0].text') as user_message,
					json_extract(value, '$.conversation[1].text') as assistant_response,
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

			// Group results by matching search term
			results := make(map[string][]map[string]interface{})
			for rows.Next() {
				var key, userMsg, assistantMsg, fullValue string
				if err := rows.Scan(&key, &userMsg, &assistantMsg, &fullValue); err != nil {
					return protocol.NewToolResult(
						protocol.WithError(fmt.Sprintf("error scanning row: %v", err)),
					), nil
				}

				// Match against each search term
				for _, term := range searchTerms {
					if strings.Contains(fullValue, term) {
						if _, ok := results[term]; !ok {
							results[term] = []map[string]interface{}{}
						}
						results[term] = append(results[term], map[string]interface{}{
							"key":                key,
							"user_message":       userMsg,
							"assistant_response": assistantMsg,
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

func RegisterGetFileReferencesTool(registry *tools.Registry) error {
	schemaJson := `{
		"type": "object",
		"properties": {
			"conversation_ids": {
				"type": "array",
				"items": {
					"type": "string"
				},
				"description": "List of conversation IDs or search terms. Each term can be a specific conversation ID or a search term. Terms containing spaces should be wrapped in quotes. Example: ['conv123', 'other456', 'search term']. Multiple unquoted words are treated as separate search terms."
			}
		},
		"required": ["conversation_ids"]
	}`

	tool, err := tools.NewToolImpl(
		"cursor-get-file-references",
		"Retrieve all files that were referenced, viewed, or modified during specific Cursor AI conversations. This includes files that were open in the editor, files that were searched, and files that were modified. Accepts space-separated terms (use quotes for terms containing spaces) for flexible searching. Example: conv123 'search term' other456",
		json.RawMessage(schemaJson))
	if err != nil {
		return err
	}

	registry.RegisterToolWithHandler(
		tool,
		func(ctx context.Context, tool tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
			conversationIDs, ok := arguments["conversation_ids"].([]interface{})
			if !ok {
				return protocol.NewToolResult(
					protocol.WithError("conversation_ids argument must be an array of strings"),
				), nil
			}

			// Process each conversation ID/search term
			var searchTerms []string
			for _, id := range conversationIDs {
				idStr := fmt.Sprintf("%v", id)
				terms := splitSearchTerms(idStr)
				searchTerms = append(searchTerms, terms...)
			}

			db, err := sql.Open("sqlite3", defaultDBPath)
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error opening database: %v", err)),
				), nil
			}
			defer db.Close()

			// Build conditions for each search term
			conditions := make([]string, len(searchTerms))
			args := make([]interface{}, len(searchTerms)*2)
			for i, term := range searchTerms {
				conditions[i] = "(key = ? OR value LIKE ?)"
				args[i*2] = term
				args[i*2+1] = "%" + term + "%"
			}

			query := fmt.Sprintf(`
				SELECT key, 
					json_extract(value, '$.conversation[0].context.fileSelections') as files
				FROM cursorDiskKV
				WHERE (%s) AND json_valid(value)
			`, strings.Join(conditions, " OR "))

			rows, err := db.QueryContext(ctx, query, args...)
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error executing query: %v", err)),
				), nil
			}
			defer rows.Close()

			results := make(map[string]interface{})
			for rows.Next() {
				var key string
				var files string
				if err := rows.Scan(&key, &files); err != nil {
					return protocol.NewToolResult(
						protocol.WithError(fmt.Sprintf("error scanning row: %v", err)),
					), nil
				}

				var fileList []interface{}
				if err := json.Unmarshal([]byte(files), &fileList); err != nil {
					continue // Skip invalid JSON
				}

				results[key] = fileList
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

func RegisterGetConversationContextTool(registry *tools.Registry) error {
	schemaJson := `{
		"type": "object",
		"properties": {
			"conversation_ids": {
				"type": "array",
				"items": {
					"type": "string"
				},
				"description": "List of conversation IDs or search terms. Each term can be a specific conversation ID or a search term. Terms containing spaces should be wrapped in quotes. Example: ['conv123', 'other456', 'search term']. Multiple unquoted words are treated as separate search terms."
			}
		},
		"required": ["conversation_ids"]
	}`

	tool, err := tools.NewToolImpl(
		"cursor-get-conversation-context",
		"Retrieve the complete context that was available to the AI during specific conversations in the Cursor IDE. This includes open files, selected text regions, cursor positions, and other IDE state information. Accepts space-separated terms (use quotes for terms containing spaces) for flexible searching. Example: conv123 'search term' other456",
		json.RawMessage(schemaJson))
	if err != nil {
		return err
	}

	registry.RegisterToolWithHandler(
		tool,
		func(ctx context.Context, tool tools.Tool, arguments map[string]interface{}) (*protocol.ToolResult, error) {
			conversationIDs, ok := arguments["conversation_ids"].([]interface{})
			if !ok {
				return protocol.NewToolResult(
					protocol.WithError("conversation_ids argument must be an array of strings"),
				), nil
			}

			// Process each conversation ID/search term
			var searchTerms []string
			for _, id := range conversationIDs {
				idStr := fmt.Sprintf("%v", id)
				terms := splitSearchTerms(idStr)
				searchTerms = append(searchTerms, terms...)
			}

			db, err := sql.Open("sqlite3", defaultDBPath)
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error opening database: %v", err)),
				), nil
			}
			defer db.Close()

			// Build conditions for each search term
			conditions := make([]string, len(searchTerms))
			args := make([]interface{}, len(searchTerms)*2)
			for i, term := range searchTerms {
				conditions[i] = "(key = ? OR value LIKE ?)"
				args[i*2] = term
				args[i*2+1] = "%" + term + "%"
			}

			query := fmt.Sprintf(`
				SELECT key, 
					json_extract(value, '$.conversation[0].context') as context
				FROM cursorDiskKV
				WHERE (%s) AND json_valid(value)
			`, strings.Join(conditions, " OR "))

			rows, err := db.QueryContext(ctx, query, args...)
			if err != nil {
				return protocol.NewToolResult(
					protocol.WithError(fmt.Sprintf("error executing query: %v", err)),
				), nil
			}
			defer rows.Close()

			results := make(map[string]interface{})
			for rows.Next() {
				var key string
				var contextData string
				if err := rows.Scan(&key, &contextData); err != nil {
					return protocol.NewToolResult(
						protocol.WithError(fmt.Sprintf("error scanning row: %v", err)),
					), nil
				}

				var context interface{}
				if err := json.Unmarshal([]byte(contextData), &context); err != nil {
					continue // Skip invalid JSON
				}

				results[key] = context
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
