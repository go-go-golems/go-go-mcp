package web

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/pkg/doc"
)

// CodeExample represents a JavaScript code example extracted from docs
type CodeExample struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Code        string `json:"code"`
	Source      string `json:"source"`   // Which file it came from
	Category    string `json:"category"` // Type of example
}

// DocsAPIHandler handles requests for documentation and code examples
func DocsAPIHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		action := r.URL.Query().Get("action")

		switch action {
		case "examples":
			handleExamples(w, r)
		case "list":
			handleDocsList(w, r)
		case "content":
			handleDocContent(w, r)
		default:
			http.Error(w, "Invalid action. Use: examples, list, or content", http.StatusBadRequest)
		}
	}
}

// handleExamples extracts and returns JavaScript code examples from docs
func handleExamples(w http.ResponseWriter, r *http.Request) {
	docsFS, err := doc.GetJSWebServerDocsFS()
	if err != nil {
		http.Error(w, "Failed to access docs filesystem", http.StatusInternalServerError)
		return
	}

	examples, err := extractCodeExamples(docsFS)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to extract examples: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(examples)
}

// handleDocsList returns a list of available documentation files
func handleDocsList(w http.ResponseWriter, r *http.Request) {
	docsFS, err := doc.GetJSWebServerDocsFS()
	if err != nil {
		http.Error(w, "Failed to access docs filesystem", http.StatusInternalServerError)
		return
	}

	files, err := fs.ReadDir(docsFS, ".")
	if err != nil {
		http.Error(w, "Failed to read docs directory", http.StatusInternalServerError)
		return
	}

	var docs []map[string]interface{}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".md") {
			docs = append(docs, map[string]interface{}{
				"name":  file.Name(),
				"title": strings.TrimSuffix(strings.ReplaceAll(file.Name(), "-", " "), ".md"),
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(docs)
}

// handleDocContent returns the content of a specific documentation file
func handleDocContent(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("file")
	if filename == "" {
		http.Error(w, "Missing file parameter", http.StatusBadRequest)
		return
	}

	docsFS, err := doc.GetJSWebServerDocsFS()
	if err != nil {
		http.Error(w, "Failed to access docs filesystem", http.StatusInternalServerError)
		return
	}

	content, err := fs.ReadFile(docsFS, filename)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"filename": filename,
		"content":  string(content),
	})
}

// extractCodeExamples extracts JavaScript code blocks from markdown files
func extractCodeExamples(docsFS fs.FS) ([]CodeExample, error) {
	var examples []CodeExample

	// Regular expression to match JavaScript code blocks
	jsCodeBlockRe := regexp.MustCompile(`(?s)` + "```javascript\n(.*?)\n```")

	files, err := fs.ReadDir(docsFS, ".")
	if err != nil {
		return nil, fmt.Errorf("failed to read docs directory: %v", err)
	}

	exampleID := 1
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".md") {
			continue
		}

		content, err := fs.ReadFile(docsFS, file.Name())
		if err != nil {
			continue
		}

		fileContent := string(content)
		category := getCategoryFromFilename(file.Name())

		// Find all JavaScript code blocks
		matches := jsCodeBlockRe.FindAllStringSubmatch(fileContent, -1)

		for _, match := range matches {
			if len(match) > 1 {
				code := strings.TrimSpace(match[1])
				if len(code) == 0 {
					continue
				}

				// Generate a name and description based on the code content
				name, description := generateExampleMetadata(code, category)

				example := CodeExample{
					ID:          fmt.Sprintf("doc-example-%d", exampleID),
					Name:        name,
					Description: description,
					Code:        code,
					Source:      file.Name(),
					Category:    category,
				}

				examples = append(examples, example)
				exampleID++
			}
		}
	}

	return examples, nil
}

// getCategoryFromFilename determines the category based on the filename
func getCategoryFromFilename(filename string) string {
	switch {
	case strings.Contains(filename, "javascript-developer"):
		return "Developer Guide"
	case strings.Contains(filename, "server-architecture"):
		return "Architecture"
	case strings.Contains(filename, "repository"):
		return "Repository Pattern"
	case strings.Contains(filename, "templ"):
		return "Template Integration"
	default:
		return "General"
	}
}

// generateExampleMetadata creates a name and description for a code example
func generateExampleMetadata(code string, category string) (string, string) {
	lines := strings.Split(code, "\n")

	// Look for comments that might indicate what the example does
	var firstComment string
	var isRouteExample bool
	var isDbExample bool
	var isMiddlewareExample bool

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for route definitions
		if strings.Contains(trimmed, "app.get") || strings.Contains(trimmed, "app.post") ||
			strings.Contains(trimmed, "app.put") || strings.Contains(trimmed, "app.delete") {
			isRouteExample = true
		}

		// Check for database operations
		if strings.Contains(trimmed, "db.query") || strings.Contains(trimmed, "db.execute") ||
			strings.Contains(trimmed, "db.get") || strings.Contains(trimmed, "db.all") {
			isDbExample = true
		}

		// Check for middleware
		if strings.Contains(trimmed, "app.use") || strings.Contains(trimmed, "next()") {
			isMiddlewareExample = true
		}

		// Capture first comment as potential description
		if strings.HasPrefix(trimmed, "//") && firstComment == "" {
			firstComment = strings.TrimSpace(strings.TrimPrefix(trimmed, "//"))
		}
	}

	// Generate name based on code content
	var name string
	var description string

	if isRouteExample {
		name = "API Routes"
		description = "Express.js route handling example"
	} else if isDbExample {
		name = "Database Operations"
		description = "SQLite database interaction example"
	} else if isMiddlewareExample {
		name = "Middleware"
		description = "Express.js middleware example"
	} else {
		name = "JavaScript Example"
		description = "General JavaScript code example"
	}

	// Use first comment as description if available and more descriptive
	if firstComment != "" && len(firstComment) > 10 {
		description = firstComment
	}

	// Add category context
	if category != "General" {
		name = fmt.Sprintf("%s - %s", category, name)
	}

	return name, description
}
