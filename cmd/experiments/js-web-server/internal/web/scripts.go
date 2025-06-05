package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/internal/engine"
	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/internal/repository"
	"github.com/rs/zerolog/log"
)

// ScriptsHandler creates a handler for the script viewer page
func ScriptsHandler(jsEngine *engine.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			serveScriptsPage(w, r, jsEngine)
		case "POST":
			serveScriptsAPI(w, r, jsEngine)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func serveScriptsPage(w http.ResponseWriter, r *http.Request, jsEngine *engine.Engine) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Script Executions - JS Playground</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/themes/prism.min.css">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/themes/prism-okaidia.min.css">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/plugins/line-numbers/prism-line-numbers.min.css">
    <style>
        .code-snippet {
            max-height: 150px;
            overflow-y: auto;
            background: #2d3748;
            border: 1px solid #4a5568;
            border-radius: 0.375rem;
            padding: 0;
            font-family: 'Fira Code', 'Courier New', monospace;
            font-size: 0.875rem;
            position: relative;
        }
        .code-snippet.expanded {
            max-height: none;
        }
        .code-snippet pre {
            margin: 0;
            padding: 0.75rem;
            background: transparent !important;
            border: none;
            border-radius: 0.375rem;
        }
        .code-snippet pre code {
            background: transparent !important;
            padding: 0 !important;
        }
        .expand-btn {
            position: absolute;
            top: 0.5rem;
            right: 0.5rem;
            z-index: 10;
            background: rgba(0,0,0,0.7);
            border: none;
            color: #fff;
            padding: 0.25rem 0.5rem;
            border-radius: 0.25rem;
            font-size: 0.75rem;
            cursor: pointer;
        }
        .expand-btn:hover {
            background: rgba(0,0,0,0.9);
        }
        .console-log {
            max-height: 100px;
            overflow-y: auto;
            background: #1a202c;
            color: #f7fafc;
            border: 1px solid #4a5568;
            border-radius: 0.375rem;
            padding: 0.75rem;
            font-family: 'Fira Code', 'Courier New', monospace;
            font-size: 0.875rem;
            position: relative;
        }
        .console-log.expanded {
            max-height: none;
        }
        .result-snippet {
            max-height: 120px;
            overflow-y: auto;
            background: #1a365d;
            border: 1px solid #2b6cb0;
            border-radius: 0.375rem;
            padding: 0;
            font-family: 'Fira Code', 'Courier New', monospace;
            font-size: 0.875rem;
            position: relative;
        }
        .result-snippet.expanded {
            max-height: none;
        }
        .result-snippet pre {
            margin: 0;
            padding: 0.75rem;
            background: transparent !important;
            border: none;
            color: #e2e8f0;
        }
        .error-text {
            color: #fc8181;
            font-family: 'Fira Code', 'Courier New', monospace;
            background: #2d1b1b;
            border: 1px solid #e53e3e;
            border-radius: 0.375rem;
            padding: 0.75rem;
            position: relative;
        }
        .error-text.expanded {
            max-height: none;
        }
        .timestamp {
            font-size: 0.875rem;
            color: #6c757d;
        }
        .session-id {
            font-family: 'Fira Code', 'Courier New', monospace;
            font-size: 0.875rem;
            background: #e9ecef;
            padding: 0.25rem 0.5rem;
            border-radius: 0.25rem;
        }
        .action-btn {
            position: absolute;
            top: 0.5rem;
            z-index: 10;
            background: rgba(0,0,0,0.7);
            border: none;
            color: #fff;
            padding: 0.25rem 0.5rem;
            border-radius: 0.25rem;
            font-size: 0.75rem;
            cursor: pointer;
            margin-left: 0.25rem;
        }
        .action-btn:hover {
            background: rgba(0,0,0,0.9);
        }
        .copy-btn {
            right: 5.5rem;
        }
        .download-btn {
            right: 3.5rem;
        }
    </style>
</head>
<body>
    <div class="container-fluid mt-4">
        <div class="row">
            <div class="col-12">
                <h1 class="mb-4">JavaScript Script Executions</h1>
                
                <!-- Search and Filter Form -->
                <div class="card mb-4">
                    <div class="card-body">
                        <form id="searchForm" class="row g-3">
                            <div class="col-md-4">
                                <label for="search" class="form-label">Search</label>
                                <input type="text" class="form-control" id="search" name="search" 
                                       placeholder="Search in code, results, or console output">
                            </div>
                            <div class="col-md-4">
                                <label for="sessionId" class="form-label">Session ID</label>
                                <input type="text" class="form-control" id="sessionId" name="sessionId" 
                                       placeholder="Filter by session ID">
                            </div>
                            <div class="col-md-2">
                                <label for="limit" class="form-label">Per Page</label>
                                <select class="form-control" id="limit" name="limit">
                                    <option value="10">10</option>
                                    <option value="25" selected>25</option>
                                    <option value="50">50</option>
                                    <option value="100">100</option>
                                </select>
                            </div>
                            <div class="col-md-2">
                                <label>&nbsp;</label>
                                <div>
                                    <button type="submit" class="btn btn-primary">Search</button>
                                    <button type="button" class="btn btn-secondary" onclick="clearForm()">Clear</button>
                                </div>
                            </div>
                        </form>
                    </div>
                </div>

                <!-- Results Container -->
                <div id="resultsContainer">
                    <div class="d-flex justify-content-center">
                        <div class="spinner-border" role="status">
                            <span class="visually-hidden">Loading...</span>
                        </div>
                    </div>
                </div>

                <!-- Pagination Container -->
                <div id="paginationContainer" class="mt-4"></div>
            </div>
        </div>
    </div>

    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/js/bootstrap.bundle.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/prism.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/prism-javascript.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/prism-json.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/plugins/line-numbers/prism-line-numbers.min.js"></script>
    <script>
        let currentPage = 1;
        let totalPages = 1;

        // Load initial data
        document.addEventListener('DOMContentLoaded', function() {
            loadScripts();
        });

        // Handle form submission
        document.getElementById('searchForm').addEventListener('submit', function(e) {
            e.preventDefault();
            currentPage = 1;
            loadScripts();
        });

        function clearForm() {
            document.getElementById('search').value = '';
            document.getElementById('sessionId').value = '';
            currentPage = 1;
            loadScripts();
        }

        function loadScripts(page = 1) {
            currentPage = page;
            const formData = new FormData(document.getElementById('searchForm'));
            formData.append('page', page);

            fetch('/admin/scripts', {
                method: 'POST',
                body: formData
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    renderResults(data.executions);
                    renderPagination(data.total, data.limit, page);
                } else {
                    document.getElementById('resultsContainer').innerHTML = 
                        '<div class="alert alert-danger">Error: ' + data.error + '</div>';
                }
            })
            .catch(error => {
                console.error('Error:', error);
                document.getElementById('resultsContainer').innerHTML = 
                    '<div class="alert alert-danger">Failed to load scripts</div>';
            });
        }

        function renderResults(executions) {
            const container = document.getElementById('resultsContainer');
            
            if (!executions || executions.length === 0) {
                container.innerHTML = '<div class="alert alert-info">No script executions found</div>';
                return;
            }

            let html = '';
            executions.forEach(exec => {
                const timestamp = new Date(exec.timestamp).toLocaleString();
                const hasError = exec.error && exec.error.trim();
                const hasResult = exec.result && exec.result.trim();
                const hasConsoleLog = exec.console_log && exec.console_log.trim();

                html += '<div class="card mb-3">';
                html += '<div class="card-header d-flex justify-content-between align-items-center">';
                html += '<div>';
                html += '<span class="session-id">' + escapeHtml(exec.session_id) + '</span>';
                html += '<span class="badge bg-secondary ms-2">' + escapeHtml(exec.source) + '</span>';
                html += '</div>';
                html += '<div class="timestamp">' + timestamp + '</div>';
                html += '</div>';
                html += '<div class="card-body">';
                
                // Store raw content in a global object for easy access
                const codeKey = 'code-' + exec.id;
                const resultKey = 'result-' + exec.id;
                const consoleKey = 'console-' + exec.id;
                const errorKey = 'error-' + exec.id;
                
                // Add to global content store
                if (!window.rawContentStore) window.rawContentStore = {};
                window.rawContentStore[codeKey] = exec.code;
                if (hasResult) window.rawContentStore[resultKey] = exec.result;
                if (hasConsoleLog) window.rawContentStore[consoleKey] = exec.console_log;
                if (hasError) window.rawContentStore[errorKey] = exec.error;
                
                // Code
                html += '<h6>Code:</h6>';
                html += '<div class="code-snippet" id="code-' + exec.id + '">';
                html += '<button class="expand-btn" onclick="toggleExpand(\'code-' + exec.id + '\')">⛶</button>';
                html += '<button class="action-btn copy-btn" onclick="copyRawContent(\'' + codeKey + '\')">📋</button>';
                html += '<button class="action-btn download-btn" onclick="downloadRawContent(\'' + codeKey + '\', \'script-' + exec.session_id + '.js\', \'text/javascript\')">💾</button>';
                html += '<pre class="line-numbers"><code class="language-javascript">' + escapeHtml(exec.code) + '</code></pre>';
                html += '</div>';
                
                // Result
                if (hasResult) {
                    html += '<h6 class="mt-3">Result:</h6>';
                    html += '<div class="result-snippet" id="result-' + exec.id + '">';
                    html += '<button class="expand-btn" onclick="toggleExpand(\'result-' + exec.id + '\')">⛶</button>';
                    html += '<button class="action-btn copy-btn" onclick="copyRawContent(\'' + resultKey + '\')">📋</button>';
                    html += '<button class="action-btn download-btn" onclick="downloadRawContent(\'' + resultKey + '\', \'result-' + exec.session_id + '.json\', \'application/json\')">💾</button>';
                    html += '<pre class="line-numbers"><code class="language-json">' + formatJson(exec.result) + '</code></pre>';
                    html += '</div>';
                }
                
                // Console Log
                if (hasConsoleLog) {
                    html += '<h6 class="mt-3">Console Output:</h6>';
                    html += '<div class="console-log" id="console-' + exec.id + '">';
                    html += '<button class="expand-btn" onclick="toggleExpand(\'console-' + exec.id + '\')">⛶</button>';
                    html += '<button class="action-btn copy-btn" onclick="copyRawContent(\'' + consoleKey + '\')">📋</button>';
                    html += '<button class="action-btn download-btn" onclick="downloadRawContent(\'' + consoleKey + '\', \'console-' + exec.session_id + '.log\', \'text/plain\')">💾</button>';
                    html += '<pre class="line-numbers"><code class="language-none">' + escapeHtml(exec.console_log) + '</code></pre>';
                    html += '</div>';
                }
                
                // Error
                if (hasError) {
                    html += '<h6 class="mt-3">Error:</h6>';
                    html += '<div class="error-text" id="error-' + exec.id + '">';
                    html += '<button class="expand-btn" onclick="toggleExpand(\'error-' + exec.id + '\')">⛶</button>';
                    html += '<button class="action-btn copy-btn" onclick="copyRawContent(\'' + errorKey + '\')">📋</button>';
                    html += '<button class="action-btn download-btn" onclick="downloadRawContent(\'' + errorKey + '\', \'error-' + exec.session_id + '.txt\', \'text/plain\')">💾</button>';
                    html += '<pre class="line-numbers"><code class="language-none">' + escapeHtml(exec.error) + '</code></pre>';
                    html += '</div>';
                }
                
                html += '</div>';
                html += '</div>';
            });

            container.innerHTML = html;
            
            // Apply syntax highlighting
            Prism.highlightAll();
        }

        function renderPagination(total, limit, currentPage) {
            const container = document.getElementById('paginationContainer');
            totalPages = Math.ceil(total / limit);
            
            if (totalPages <= 1) {
                container.innerHTML = '';
                return;
            }

            let html = '<nav aria-label="Script executions pagination">';
            html += '<ul class="pagination justify-content-center">';
            
            // Previous button
            html += '<li class="page-item' + (currentPage === 1 ? ' disabled' : '') + '">';
            html += '<a class="page-link" href="#" onclick="loadScripts(' + (currentPage - 1) + ')">Previous</a>';
            html += '</li>';
            
            // Page numbers
            for (let i = Math.max(1, currentPage - 2); i <= Math.min(totalPages, currentPage + 2); i++) {
                html += '<li class="page-item' + (i === currentPage ? ' active' : '') + '">';
                html += '<a class="page-link" href="#" onclick="loadScripts(' + i + ')">' + i + '</a>';
                html += '</li>';
            }
            
            // Next button
            html += '<li class="page-item' + (currentPage === totalPages ? ' disabled' : '') + '">';
            html += '<a class="page-link" href="#" onclick="loadScripts(' + (currentPage + 1) + ')">Next</a>';
            html += '</li>';
            
            html += '</ul>';
            html += '</nav>';
            
            // Show total count
            html += '<div class="text-center text-muted">';
            html += 'Showing ' + ((currentPage - 1) * limit + 1) + '-' + Math.min(currentPage * limit, total) + ' of ' + total + ' executions';
            html += '</div>';

            container.innerHTML = html;
        }

        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }

        function toggleExpand(elementId) {
            const element = document.getElementById(elementId);
            if (element) {
                element.classList.toggle('expanded');
                const btn = element.querySelector('.expand-btn');
                if (btn) {
                    btn.textContent = element.classList.contains('expanded') ? '⛷' : '⛶';
                }
            }
        }

        function copyRawContent(contentKey) {
            if (window.rawContentStore && window.rawContentStore[contentKey]) {
                const rawContent = window.rawContentStore[contentKey];
                navigator.clipboard.writeText(rawContent).then(function() {
                    // Show temporary feedback
                    const btn = event.target;
                    const originalText = btn.textContent;
                    btn.textContent = '✓';
                    setTimeout(() => {
                        btn.textContent = originalText;
                    }, 1000);
                }).catch(function(err) {
                    console.error('Failed to copy text: ', err);
                });
            } else {
                console.error('Content not found for key:', contentKey);
            }
        }

        function downloadRawContent(contentKey, filename, mimeType) {
            if (window.rawContentStore && window.rawContentStore[contentKey]) {
                const rawContent = window.rawContentStore[contentKey];
                
                // Create a blob with the content
                const blob = new Blob([rawContent], { type: mimeType });
                
                // Create a temporary URL for the blob
                const url = window.URL.createObjectURL(blob);
                
                // Create a temporary anchor element and trigger download
                const a = document.createElement('a');
                a.style.display = 'none';
                a.href = url;
                a.download = filename;
                
                // Add to DOM, click, and remove
                document.body.appendChild(a);
                a.click();
                document.body.removeChild(a);
                
                // Clean up the URL
                window.URL.revokeObjectURL(url);
                
                // Show temporary feedback
                const btn = event.target;
                const originalText = btn.textContent;
                btn.textContent = '✓';
                setTimeout(() => {
                    btn.textContent = originalText;
                }, 1000);
            } else {
                console.error('Content not found for key:', contentKey);
            }
        }

        function formatJson(jsonString) {
            try {
                const parsed = JSON.parse(jsonString);
                return escapeHtml(JSON.stringify(parsed, null, 2));
            } catch (e) {
                return escapeHtml(jsonString);
            }
        }
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write([]byte(html)); err != nil {
		log.Error().Err(err).Msg("Failed to write HTML response")
	}
}

func serveScriptsAPI(w http.ResponseWriter, r *http.Request, jsEngine *engine.Engine) {
	// Parse form data (handles both application/x-www-form-urlencoded and multipart/form-data)
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		// If multipart parsing fails, try regular form parsing
		if err := r.ParseForm(); err != nil {
			log.Error().Err(err).Msg("Failed to parse form data")
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}
	}

	// Get parameters
	search := strings.TrimSpace(r.FormValue("search"))
	sessionID := strings.TrimSpace(r.FormValue("sessionId"))
	limitStr := r.FormValue("limit")
	pageStr := r.FormValue("page")

	// Parse pagination parameters
	limit := 25 // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	page := 1 // default
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	offset := (page - 1) * limit

	log.Info().
		Str("search", search).
		Str("sessionID", sessionID).
		Str("limitStr", limitStr).
		Str("pageStr", pageStr).
		Int("limit", limit).
		Int("offset", offset).
		Interface("form", r.Form).
		Msg("Scripts API request")

	// Query via repository
	filter := repository.ExecutionFilter{
		Search:    search,
		SessionID: sessionID,
	}
	pagination := repository.PaginationOptions{
		Limit:  limit,
		Offset: offset,
	}
	
	result, err := jsEngine.GetRepositoryManager().Executions().ListExecutions(r.Context(), filter, pagination)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get script executions")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if encodeErr := json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Repository error",
		}); encodeErr != nil {
			log.Error().Err(encodeErr).Msg("Failed to encode error response")
		}
		return
	}
	
	executions := result.Executions
	total := result.Total

	// Return JSON response
	response := map[string]interface{}{
		"success":    true,
		"executions": executions,
		"total":      total,
		"limit":      limit,
		"page":       page,
		"totalPages": (total + limit - 1) / limit, // ceiling division
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error().Err(err).Msg("Failed to encode response")
	}
}
