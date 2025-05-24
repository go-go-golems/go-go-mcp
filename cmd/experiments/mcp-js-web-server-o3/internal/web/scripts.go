package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-go-golems/go-go-mcp/cmd/experiments/mcp-js-web-server-o3/internal/engine"
	"github.com/rs/zerolog/log"
)

// ScriptsHandler creates a handler for the script viewer page
func ScriptsHandler(jsEngine *engine.Engine) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			serveScriptsPage(w, r, jsEngine)
		} else if r.Method == "POST" {
			serveScriptsAPI(w, r, jsEngine)
		} else {
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
    <style>
        .code-snippet {
            max-height: 150px;
            overflow-y: auto;
            background: #f8f9fa;
            border: 1px solid #dee2e6;
            border-radius: 0.375rem;
            padding: 0.5rem;
            font-family: 'Courier New', monospace;
            font-size: 0.875rem;
        }
        .console-log {
            max-height: 100px;
            overflow-y: auto;
            background: #1e1e1e;
            color: #f8f8f2;
            border-radius: 0.375rem;
            padding: 0.5rem;
            font-family: 'Courier New', monospace;
            font-size: 0.875rem;
        }
        .error-text {
            color: #dc3545;
            font-family: 'Courier New', monospace;
        }
        .timestamp {
            font-size: 0.875rem;
            color: #6c757d;
        }
        .session-id {
            font-family: 'Courier New', monospace;
            font-size: 0.875rem;
            background: #e9ecef;
            padding: 0.25rem 0.5rem;
            border-radius: 0.25rem;
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
                
                // Code
                html += '<h6>Code:</h6>';
                html += '<div class="code-snippet"><pre>' + escapeHtml(exec.code) + '</pre></div>';
                
                // Result
                if (hasResult) {
                    html += '<h6 class="mt-3">Result:</h6>';
                    html += '<div class="code-snippet"><pre>' + escapeHtml(exec.result) + '</pre></div>';
                }
                
                // Console Log
                if (hasConsoleLog) {
                    html += '<h6 class="mt-3">Console Output:</h6>';
                    html += '<div class="console-log"><pre>' + escapeHtml(exec.console_log) + '</pre></div>';
                }
                
                // Error
                if (hasError) {
                    html += '<h6 class="mt-3">Error:</h6>';
                    html += '<div class="error-text"><pre>' + escapeHtml(exec.error) + '</pre></div>';
                }
                
                html += '</div>';
                html += '</div>';
            });

            container.innerHTML = html;
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
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func serveScriptsAPI(w http.ResponseWriter, r *http.Request, jsEngine *engine.Engine) {
	// Parse form data
	if err := r.ParseForm(); err != nil {
		log.Error().Err(err).Msg("Failed to parse form data")
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
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

	log.Debug().
		Str("search", search).
		Str("sessionID", sessionID).
		Int("limit", limit).
		Int("offset", offset).
		Msg("Scripts API request")

	// Query the database
	executions, total, err := jsEngine.GetScriptExecutions(search, sessionID, limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get script executions")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Database error",
		})
		return
	}

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
	json.NewEncoder(w).Encode(response)
}