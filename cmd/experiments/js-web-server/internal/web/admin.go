package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/internal/engine"
	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/internal/repository"
	"github.com/rs/zerolog/log"
)

// AdminHandler provides admin endpoints for managing the server
type AdminHandler struct {
	logger   *engine.RequestLogger
	repos    repository.RepositoryManager
	jsEngine *engine.Engine

	// SSE support
	sseClients map[string]chan string
	sseMutex   sync.RWMutex
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(logger *engine.RequestLogger, repos repository.RepositoryManager, jsEngine *engine.Engine) *AdminHandler {
	ah := &AdminHandler{
		logger:     logger,
		repos:      repos,
		jsEngine:   jsEngine,
		sseClients: make(map[string]chan string),
	}

	// Start monitoring for new requests
	go ah.monitorNewRequests()

	return ah
}

// HandleAdminLogs serves the admin logs interface
func (ah *AdminHandler) HandleAdminLogs(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/admin/logs" {
		ah.serveLogsInterface(w, r)
		return
	}

	if r.URL.Path == "/admin/logs/api" {
		ah.serveLogsAPI(w, r)
		return
	}

	if strings.HasPrefix(r.URL.Path, "/admin/logs/api/") {
		ah.serveLogsAPI(w, r)
		return
	}

	if r.URL.Path == "/admin/logs/events" {
		ah.serveSSE(w, r)
		return
	}

	http.NotFound(w, r)
}

// HandleGlobalState serves the globalState interface and API
func (ah *AdminHandler) HandleGlobalState(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		if r.Header.Get("Accept") == "application/json" {
			// API request - return JSON
			globalState := ah.jsEngine.GetGlobalState()
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(globalState))
		} else {
			// Regular request - serve HTML interface
			ah.serveGlobalStateInterface(w, r)
		}
	case "POST":
		// Update globalState
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		jsonData := r.FormValue("globalState")
		if jsonData == "" {
			http.Error(w, "Missing globalState data", http.StatusBadRequest)
			return
		}

		if err := ah.jsEngine.SetGlobalState(jsonData); err != nil {
			http.Error(w, "Failed to update globalState: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Return success response
		if r.Header.Get("Accept") == "application/json" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"success": true}`))
		} else {
			// Redirect back to the interface
			http.Redirect(w, r, "/admin/globalstate", http.StatusSeeOther)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// serveLogsInterface serves the HTML interface for viewing logs
func (ah *AdminHandler) serveLogsInterface(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Request Logs - Admin Console</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: #f5f5f5;
            color: #333;
        }
        
        .header {
            background: #2c3e50;
            color: white;
            padding: 1rem 2rem;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        
        .header h1 {
            display: inline-block;
            margin-right: 2rem;
        }
        
        .controls {
            display: inline-block;
        }
        
        .controls button {
            background: #3498db;
            color: white;
            border: none;
            padding: 0.5rem 1rem;
            margin: 0 0.5rem;
            border-radius: 4px;
            cursor: pointer;
            font-size: 0.9rem;
        }
        
        .controls button:hover {
            background: #2980b9;
        }
        
        .controls button.danger {
            background: #e74c3c;
        }
        
        .controls button.danger:hover {
            background: #c0392b;
        }
        
        .main-content {
            display: flex;
            height: calc(100vh - 80px);
        }
        
        .sidebar {
            width: 300px;
            background: white;
            border-right: 1px solid #ddd;
            overflow-y: auto;
        }
        
        .stats {
            padding: 1rem;
            border-bottom: 1px solid #eee;
            background: #f8f9fa;
        }
        
        .stats h3 {
            margin-bottom: 0.5rem;
            color: #2c3e50;
        }
        
        .stats .stat-item {
            display: flex;
            justify-content: space-between;
            margin: 0.25rem 0;
            font-size: 0.9rem;
        }
        
        .request-list {
            padding: 1rem;
        }
        
        .request-item {
            padding: 0.75rem;
            border: 1px solid #ddd;
            border-radius: 4px;
            margin-bottom: 0.5rem;
            cursor: pointer;
            transition: all 0.2s;
        }
        
        .request-item:hover {
            background: #f0f0f0;
            border-color: #3498db;
        }
        
        .request-item.selected {
            background: #e3f2fd;
            border-color: #2196f3;
        }
        
        .request-summary {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 0.25rem;
        }
        
        .request-method {
            font-weight: bold;
            padding: 0.25rem 0.5rem;
            border-radius: 3px;
            font-size: 0.8rem;
        }
        
        .method-GET { background: #4caf50; color: white; }
        .method-POST { background: #ff9800; color: white; }
        .method-PUT { background: #2196f3; color: white; }
        .method-DELETE { background: #f44336; color: white; }
        .method-PATCH { background: #9c27b0; color: white; }
        
        .request-path {
            font-family: monospace;
            font-size: 0.9rem;
        }
        
        .request-status {
            font-weight: bold;
            padding: 0.25rem 0.5rem;
            border-radius: 3px;
            font-size: 0.8rem;
        }
        
        .status-2xx { background: #4caf50; color: white; }
        .status-3xx { background: #ff9800; color: white; }
        .status-4xx { background: #ff5722; color: white; }
        .status-5xx { background: #f44336; color: white; }
        
        .request-time {
            font-size: 0.8rem;
            color: #666;
        }
        
        .details-panel {
            flex: 1;
            background: white;
            overflow-y: auto;
        }
        
        .no-selection {
            display: flex;
            align-items: center;
            justify-content: center;
            height: 100%;
            color: #999;
            font-size: 1.2rem;
        }
        
        .request-details {
            padding: 2rem;
        }
        
        .details-header {
            border-bottom: 2px solid #eee;
            padding-bottom: 1rem;
            margin-bottom: 1.5rem;
        }
        
        .details-title {
            display: flex;
            align-items: center;
            gap: 1rem;
            margin-bottom: 0.5rem;
        }
        
        .details-meta {
            color: #666;
            font-size: 0.9rem;
        }
        
        .section {
            margin-bottom: 2rem;
        }
        
        .section h3 {
            margin-bottom: 1rem;
            color: #2c3e50;
            border-bottom: 1px solid #eee;
            padding-bottom: 0.5rem;
        }
        
        .logs-container {
            background: #1e1e1e;
            color: #f0f0f0;
            padding: 1rem;
            border-radius: 4px;
            font-family: 'Courier New', monospace;
            font-size: 0.9rem;
            max-height: 400px;
            overflow-y: auto;
        }
        
        .log-entry {
            margin-bottom: 0.5rem;
            padding: 0.25rem 0;
            border-left: 3px solid transparent;
            padding-left: 0.5rem;
        }
        
        .log-entry.log { border-left-color: #4caf50; }
        .log-entry.info { border-left-color: #2196f3; }
        .log-entry.warn { border-left-color: #ff9800; }
        .log-entry.error { border-left-color: #f44336; }
        .log-entry.debug { border-left-color: #9c27b0; }
        
        .log-time {
            color: #888;
            font-size: 0.8rem;
        }
        
        .log-level {
            font-weight: bold;
            text-transform: uppercase;
            margin: 0 0.5rem;
        }
        
        .db-operations-container {
            background: #f8f9fa;
            border: 1px solid #ddd;
            border-radius: 4px;
            padding: 1rem;
            max-height: 400px;
            overflow-y: auto;
        }
        
        .db-operation {
            margin-bottom: 1rem;
            padding: 0.75rem;
            border-radius: 4px;
            border-left: 4px solid #28a745;
        }
        
        .db-operation.error {
            border-left-color: #dc3545;
            background: #fff5f5;
        }
        
        .db-operation.success {
            border-left-color: #28a745;
            background: #f0fff4;
        }
        
        .db-op-header {
            display: flex;
            align-items: center;
            gap: 1rem;
            margin-bottom: 0.5rem;
            font-size: 0.9rem;
        }
        
        .db-op-type {
            background: #007bff;
            color: white;
            padding: 0.25rem 0.5rem;
            border-radius: 3px;
            font-weight: bold;
            font-size: 0.8rem;
        }
        
        .db-op-time {
            color: #666;
            font-size: 0.8rem;
        }
        
        .db-op-duration {
            background: #6c757d;
            color: white;
            padding: 0.25rem 0.5rem;
            border-radius: 3px;
            font-size: 0.8rem;
        }
        
        .db-op-error {
            background: #dc3545;
            color: white;
            padding: 0.25rem 0.5rem;
            border-radius: 3px;
            font-weight: bold;
            font-size: 0.8rem;
        }
        
        .db-op-sql {
            background: #2d3748;
            color: #e2e8f0;
            padding: 0.5rem;
            border-radius: 4px;
            margin: 0.5rem 0;
            font-family: 'Courier New', monospace;
            font-size: 0.9rem;
            overflow-x: auto;
        }
        
        .db-op-params {
            color: #495057;
            font-size: 0.9rem;
            margin: 0.25rem 0;
        }
        
        .db-op-result {
            color: #28a745;
            font-size: 0.9rem;
            margin: 0.25rem 0;
        }
        
        .db-op-error-msg {
            color: #dc3545;
            font-size: 0.9rem;
            margin: 0.25rem 0;
        }
        
        .json-display {
            background: #f8f9fa;
            border: 1px solid #ddd;
            border-radius: 4px;
            padding: 1rem;
            font-family: monospace;
            font-size: 0.9rem;
            overflow-x: auto;
        }
        
        .key-value {
            display: flex;
            margin-bottom: 0.5rem;
        }
        
        .key {
            font-weight: bold;
            width: 120px;
            color: #2c3e50;
        }
        
        .value {
            flex: 1;
            font-family: monospace;
        }
        
        .auto-refresh {
            margin-left: 1rem;
        }
        
        .auto-refresh input {
            margin-right: 0.5rem;
        }
        
        .tabs {
            display: flex;
            background: #f8f9fa;
            border-bottom: 2px solid #ddd;
            margin: 0;
            padding: 0;
        }
        
        .tab-button {
            background: none;
            border: none;
            padding: 1rem 2rem;
            cursor: pointer;
            font-size: 1rem;
            border-bottom: 3px solid transparent;
            transition: all 0.3s ease;
        }
        
        .tab-button:hover {
            background: #e9ecef;
        }
        
        .tab-button.active {
            background: white;
            border-bottom-color: #007bff;
            font-weight: bold;
        }
        
        .tab-content {
            display: none;
        }
        
        .tab-content.active {
            display: block;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>Admin Logs</h1>
        <div class="controls">
            <button onclick="refreshLogs()">Refresh</button>
            <button onclick="clearLogs()" class="danger">Clear Logs</button>
            <div class="auto-refresh">
                <input type="checkbox" id="autoRefresh" onchange="toggleAutoRefresh()">
                <label for="autoRefresh">Auto-refresh (5s)</label>
            </div>
            <div style="margin-left: auto;">
                <a href="/admin/globalstate" style="color: white; text-decoration: none; padding: 0.5rem 1rem; background: rgba(255,255,255,0.1); border-radius: 4px;">GlobalState Inspector</a>
                <a href="/admin/scripts" style="color: white; text-decoration: none; padding: 0.5rem 1rem; background: rgba(255,255,255,0.1); border-radius: 4px; margin-left: 0.5rem;">Scripts</a>
            </div>
        </div>
    </div>
    
    <div class="tabs">
        <button class="tab-button active" onclick="switchTab('requests')">HTTP Requests</button>
        <button class="tab-button" onclick="switchTab('executions')">Script Executions</button>
    </div>
    
    <div class="main-content">
        <div class="sidebar">
            <div class="tab-content active" id="requests-tab">
                <div class="stats" id="stats">
                    <h3>HTTP Statistics</h3>
                    <div class="stat-item">
                        <span>Total Requests:</span>
                        <span id="totalRequests">0</span>
                    </div>
                </div>
                <div class="request-list" id="requestList">
                    <p>Loading requests...</p>
                </div>
            </div>
            
            <div class="tab-content" id="executions-tab">
                <div class="stats" id="execStats">
                    <h3>Execution Statistics</h3>
                    <div class="stat-item">
                        <span>Total Executions:</span>
                        <span id="totalExecutions">0</span>
                    </div>
                </div>
                <div class="request-list" id="executionList">
                    <p>Loading executions...</p>
                </div>
            </div>
        </div>
        
        <div class="details-panel">
            <div class="no-selection" id="noSelection">
                Select an item to view details
            </div>
            <div class="request-details" id="requestDetails" style="display: none;">
                <!-- Request details will be populated here -->
            </div>
            <div class="execution-details" id="executionDetails" style="display: none;">
                <!-- Execution details will be populated here -->
            </div>
        </div>
    </div>

    <script>
        let autoRefreshInterval = null;
        let selectedRequestId = null;
        let eventSource = null;
        let isRealTimeEnabled = false;
        
        async function loadStats() {
            try {
                const response = await fetch('/admin/logs/api/stats');
                const stats = await response.json();
                
                let statsHTML = '<h3>Statistics</h3>';
                statsHTML += '<div class="stat-item"><span>Total Requests:</span><span>' + stats.totalRequests + '</span></div>';
                statsHTML += '<div class="stat-item"><span>Max Logs:</span><span>' + stats.maxLogs + '</span></div>';
                
                if (stats.averageDuration) {
                    const avgMs = Math.round(stats.averageDuration / 1000000); // Convert nanoseconds to milliseconds
                    statsHTML += '<div class="stat-item"><span>Avg Duration:</span><span>' + avgMs + 'ms</span></div>';
                }
                
                if (stats.methodCounts) {
                    statsHTML += '<h4 style="margin-top: 1rem; margin-bottom: 0.5rem;">Methods</h4>';
                    for (const [method, count] of Object.entries(stats.methodCounts)) {
                        statsHTML += '<div class="stat-item"><span>' + method + ':</span><span>' + count + '</span></div>';
                    }
                }
                
                if (stats.statusCounts) {
                    statsHTML += '<h4 style="margin-top: 1rem; margin-bottom: 0.5rem;">Status Codes</h4>';
                    for (const [status, count] of Object.entries(stats.statusCounts)) {
                        statsHTML += '<div class="stat-item"><span>' + status + ':</span><span>' + count + '</span></div>';
                    }
                }
                
                document.getElementById('stats').innerHTML = statsHTML;
            } catch (error) {
                console.error('Failed to load stats:', error);
            }
        }
        
        async function loadRequests() {
            try {
                const response = await fetch('/admin/logs/api/requests?limit=50');
                const requests = await response.json();
                
                const requestList = document.getElementById('requestList');
                if (requests.length === 0) {
                    requestList.innerHTML = '<p>No requests logged yet</p>';
                    return;
                }
                
                let html = '';
                requests.forEach(request => {
                    const time = new Date(request.startTime).toLocaleTimeString();
                    const statusClass = 'status-' + Math.floor(request.status / 100) + 'xx';
                    const methodClass = 'method-' + request.method;
                    const duration = Math.round(request.duration / 1000000); // Convert to milliseconds
                    
                    html += '<div class="request-item" onclick="selectRequest(\'' + request.id + '\')" data-id="' + request.id + '">';
                    html += '  <div class="request-summary">';
                    html += '    <span class="request-method ' + methodClass + '">' + request.method + '</span>';
                    html += '    <span class="request-status ' + statusClass + '">' + request.status + '</span>';
                    html += '  </div>';
                    html += '  <div class="request-path">' + request.path + '</div>';
                    html += '  <div class="request-time">' + time + ' (' + duration + 'ms)</div>';
                    html += '</div>';
                });
                
                requestList.innerHTML = html;
                
                // Restore selection if it exists
                if (selectedRequestId) {
                    const element = document.querySelector('[data-id="' + selectedRequestId + '"]');
                    if (element) {
                        element.classList.add('selected');
                    }
                }
            } catch (error) {
                console.error('Failed to load requests:', error);
                document.getElementById('requestList').innerHTML = '<p>Error loading requests</p>';
            }
        }
        
        async function selectRequest(requestId) {
            // Update UI selection
            document.querySelectorAll('.request-item').forEach(item => {
                item.classList.remove('selected');
            });
            document.querySelector('[data-id="' + requestId + '"]').classList.add('selected');
            selectedRequestId = requestId;
            
            try {
                const response = await fetch('/admin/logs/api/requests/' + requestId);
                const request = await response.json();
                
                const noSelection = document.getElementById('noSelection');
                const requestDetails = document.getElementById('requestDetails');
                
                noSelection.style.display = 'none';
                requestDetails.style.display = 'block';
                
                // Build details HTML
                const startTime = new Date(request.startTime).toLocaleString();
                const endTime = request.endTime ? new Date(request.endTime).toLocaleString() : 'N/A';
                const duration = request.duration ? Math.round(request.duration / 1000000) + 'ms' : 'N/A';
                const statusClass = 'status-' + Math.floor(request.status / 100) + 'xx';
                const methodClass = 'method-' + request.method;
                
                let html = '<div class="details-header">';
                html += '  <div class="details-title">';
                html += '    <span class="request-method ' + methodClass + '">' + request.method + '</span>';
                html += '    <span class="request-path">' + request.path + '</span>';
                html += '    <span class="request-status ' + statusClass + '">' + request.status + '</span>';
                html += '  </div>';
                html += '  <div class="details-meta">';
                html += '    <div>Started: ' + startTime + '</div>';
                html += '    <div>Duration: ' + duration + '</div>';
                html += '    <div>Remote IP: ' + (request.remoteIP || 'N/A') + '</div>';
                html += '  </div>';
                html += '</div>';
                
                // Request details
                if (request.query && Object.keys(request.query).length > 0) {
                    html += '<div class="section">';
                    html += '  <h3>Query Parameters</h3>';
                    html += '  <div class="json-display">' + JSON.stringify(request.query, null, 2) + '</div>';
                    html += '</div>';
                }
                
                if (request.headers && Object.keys(request.headers).length > 0) {
                    html += '<div class="section">';
                    html += '  <h3>Request Headers</h3>';
                    html += '  <div class="json-display">' + JSON.stringify(request.headers, null, 2) + '</div>';
                    html += '</div>';
                }
                
                if (request.body) {
                    html += '<div class="section">';
                    html += '  <h3>Request Body</h3>';
                    html += '  <div class="json-display">' + request.body + '</div>';
                    html += '</div>';
                }
                
                if (request.response) {
                    html += '<div class="section">';
                    html += '  <h3>Response</h3>';
                    html += '  <div class="json-display">' + request.response + '</div>';
                    html += '</div>';
                }
                
                // Database Operations
                if (request.databaseOps && request.databaseOps.length > 0) {
                    html += '<div class="section">';
                    html += '  <h3>Database Operations (' + request.databaseOps.length + ')</h3>';
                    html += '  <div class="db-operations-container">';
                    request.databaseOps.forEach(op => {
                        const time = new Date(op.timestamp).toLocaleTimeString();
                        const durationMs = Math.round(op.duration / 1000000); // Convert nanoseconds to milliseconds
                        const statusClass = op.error ? 'error' : 'success';
                        
                        html += '<div class="db-operation ' + statusClass + '">';
                        html += '  <div class="db-op-header">';
                        html += '    <span class="db-op-type">' + op.type.toUpperCase() + '</span>';
                        html += '    <span class="db-op-time">' + time + '</span>';
                        html += '    <span class="db-op-duration">' + durationMs + 'ms</span>';
                        if (op.error) {
                            html += '    <span class="db-op-error">ERROR</span>';
                        }
                        html += '  </div>';
                        html += '  <div class="db-op-sql"><code>' + op.sql + '</code></div>';
                        if (op.parameters && op.parameters.length > 0) {
                            html += '  <div class="db-op-params">Parameters: <code>' + JSON.stringify(op.parameters) + '</code></div>';
                        }
                        if (op.error) {
                            html += '  <div class="db-op-error-msg">Error: ' + op.error + '</div>';
                        } else if (op.result) {
                            html += '  <div class="db-op-result">Result: ' + op.result + '</div>';
                        }
                        html += '</div>';
                    });
                    html += '  </div>';
                    html += '</div>';
                }

                // Console Logs
                if (request.logs && request.logs.length > 0) {
                    html += '<div class="section">';
                    html += '  <h3>Console Logs (' + request.logs.length + ')</h3>';
                    html += '  <div class="logs-container">';
                    request.logs.forEach(log => {
                        const time = new Date(log.timestamp).toLocaleTimeString();
                        html += '<div class="log-entry ' + log.level + '">';
                        html += '  <span class="log-time">' + time + '</span>';
                        html += '  <span class="log-level">' + log.level + '</span>';
                        html += '  <span class="log-message">' + log.message + '</span>';
                        if (log.data) {
                            html += '<br><span style="margin-left: 2rem; color: #ccc;">' + JSON.stringify(log.data) + '</span>';
                        }
                        html += '</div>';
                    });
                    html += '  </div>';
                    html += '</div>';
                }
                
                if (request.error) {
                    html += '<div class="section">';
                    html += '  <h3>Error</h3>';
                    html += '  <div class="json-display" style="color: #e74c3c;">' + request.error + '</div>';
                    html += '</div>';
                }
                
                requestDetails.innerHTML = html;
            } catch (error) {
                console.error('Failed to load request details:', error);
            }
        }
        
        async function refreshLogs() {
            const activeTab = document.querySelector('.tab-button.active').onclick.toString().match(/switchTab\('([^']+)'\)/)[1];
            if (activeTab === 'requests') {
                await Promise.all([loadStats(), loadRequests()]);
            } else if (activeTab === 'executions') {
                await Promise.all([loadExecutionStats(), loadExecutions()]);
            }
        }
        
        async function clearLogs() {
            if (confirm('Are you sure you want to clear all logs?')) {
                try {
                    await fetch('/admin/logs/api/clear', { method: 'POST' });
                    selectedRequestId = null;
                    document.getElementById('noSelection').style.display = 'flex';
                    document.getElementById('requestDetails').style.display = 'none';
                    await refreshLogs();
                } catch (error) {
                    console.error('Failed to clear logs:', error);
                    alert('Failed to clear logs');
                }
            }
        }
        
        function toggleAutoRefresh() {
            const checkbox = document.getElementById('autoRefresh');
            if (checkbox.checked) {
                startRealTimeUpdates();
            } else {
                stopRealTimeUpdates();
            }
        }
        
        function startRealTimeUpdates() {
            if (isRealTimeEnabled) return;
            
            try {
                // Try Server-Sent Events first
                eventSource = new EventSource('/admin/logs/events');
                
                eventSource.onopen = function() {
                    console.log('Real-time updates connected');
                    isRealTimeEnabled = true;
                    updateConnectionStatus(true);
                };
                
                eventSource.onmessage = function(event) {
                    try {
                        const data = JSON.parse(event.data);
                        handleRealTimeUpdate(data);
                    } catch (e) {
                        console.error('Failed to parse SSE message:', e);
                    }
                };
                
                eventSource.onerror = function(event) {
                    console.warn('SSE connection failed, falling back to polling');
                    if (eventSource) {
                        eventSource.close();
                        eventSource = null;
                    }
                    fallbackToPolling();
                };
                
            } catch (e) {
                console.warn('SSE not supported, using polling');
                fallbackToPolling();
            }
        }
        
        function stopRealTimeUpdates() {
            if (eventSource) {
                eventSource.close();
                eventSource = null;
            }
            if (autoRefreshInterval) {
                clearInterval(autoRefreshInterval);
                autoRefreshInterval = null;
            }
            isRealTimeEnabled = false;
            updateConnectionStatus(false);
        }
        
        function fallbackToPolling() {
            if (!autoRefreshInterval) {
                autoRefreshInterval = setInterval(refreshLogs, 5000);
                isRealTimeEnabled = true;
                updateConnectionStatus(true);
            }
        }
        
        function handleRealTimeUpdate(data) {
            switch (data.type) {
                case 'connected':
                    console.log('SSE connected with client ID:', data.clientId);
                    break;
                case 'newRequest':
                    const activeTab = document.querySelector('.tab-button.active').onclick.toString().match(/switchTab\('([^']+)'\)/)[1];
                    if (activeTab === 'requests') {
                        loadStats();
                        loadRequests();
                    }
                    break;
                case 'newExecution':
                    const currentTab = document.querySelector('.tab-button.active').onclick.toString().match(/switchTab\('([^']+)'\)/)[1];
                    if (currentTab === 'executions') {
                        loadExecutionStats();
                        loadExecutions();
                    }
                    break;
            }
        }
        
        function updateConnectionStatus(connected) {
            const label = document.querySelector('label[for="autoRefresh"]');
            if (connected) {
                label.textContent = 'Real-time updates';
                label.style.color = '#28a745';
            } else {
                label.textContent = 'Auto-refresh (5s)';
                label.style.color = '';
            }
        }
        
        async function loadExecutionStats() {
            try {
                const response = await fetch('/admin/logs/api/executions?limit=0');
                const result = await response.json();
                
                let statsHTML = '<h3>Execution Statistics</h3>';
                statsHTML += '<div class="stat-item"><span>Total Executions:</span><span>' + (result.total || 0) + '</span></div>';
                
                document.getElementById('execStats').innerHTML = statsHTML;
            } catch (error) {
                console.error('Failed to load execution stats:', error);
            }
        }
        
        async function loadExecutions() {
            try {
                const response = await fetch('/admin/logs/api/executions?limit=50');
                const result = await response.json();
                const executions = result.executions || [];
                
                const executionList = document.getElementById('executionList');
                if (executions.length === 0) {
                    executionList.innerHTML = '<p>No script executions logged yet</p>';
                    return;
                }
                
                let html = '';
                executions.forEach(execution => {
                    const time = new Date(execution.timestamp).toLocaleTimeString();
                    const statusClass = execution.error ? 'error' : 'success';
                    const shortCode = execution.code ? execution.code.substring(0, 50) + (execution.code.length > 50 ? '...' : '') : '';
                    
                    html += '<div class="request-item ' + statusClass + '" onclick="loadExecutionDetails(' + execution.id + ')">';
                    html += '  <div class="request-time">' + time + '</div>';
                    html += '  <div class="request-method">' + (execution.source || 'EXEC') + '</div>';
                    html += '  <div class="request-path">' + shortCode + '</div>';
                    if (execution.error) {
                        html += '  <div class="request-status error">ERROR</div>';
                    } else {
                        html += '  <div class="request-status success">SUCCESS</div>';
                    }
                    html += '</div>';
                });
                
                executionList.innerHTML = html;
            } catch (error) {
                console.error('Failed to load executions:', error);
            }
        }
        
        async function loadExecutionDetails(executionId) {
            try {
                selectedRequestId = executionId;
                const response = await fetch('/admin/logs/api/executions/' + executionId);
                const execution = await response.json();
                
                document.getElementById('noSelection').style.display = 'none';
                document.getElementById('requestDetails').style.display = 'none';
                document.getElementById('executionDetails').style.display = 'block';
                
                const executionDetails = document.getElementById('executionDetails');
                
                let html = '<div class="details-header">';
                html += '  <div class="details-title">';
                html += '    <h2>Execution #' + execution.id + '</h2>';
                if (execution.error) {
                    html += '    <span class="status error">ERROR</span>';
                } else {
                    html += '    <span class="status success">SUCCESS</span>';
                }
                html += '  </div>';
                html += '  <div class="details-meta">';
                html += '    <span>Source: ' + (execution.source || 'unknown') + '</span>';
                html += '    <span>Time: ' + new Date(execution.timestamp).toLocaleString() + '</span>';
                if (execution.session_id) {
                    html += '    <span>Session: ' + execution.session_id + '</span>';
                }
                html += '  </div>';
                html += '</div>';
                
                // Code section
                html += '<div class="section">';
                html += '  <h3>JavaScript Code</h3>';
                html += '  <div class="logs-container">';
                html += '    <pre style="color: #f8f8f2; margin: 0;">' + (execution.code || 'No code') + '</pre>';
                html += '  </div>';
                html += '</div>';
                
                // Result section
                if (execution.result) {
                    html += '<div class="section">';
                    html += '  <h3>Result</h3>';
                    html += '  <div class="json-display">' + execution.result + '</div>';
                    html += '</div>';
                }
                
                // Console logs
                if (execution.console_log) {
                    html += '<div class="section">';
                    html += '  <h3>Console Output</h3>';
                    html += '  <div class="logs-container">';
                    html += '    <pre style="color: #f8f8f2; margin: 0;">' + execution.console_log + '</pre>';
                    html += '  </div>';
                    html += '</div>';
                }
                
                // Error section
                if (execution.error) {
                    html += '<div class="section">';
                    html += '  <h3>Error</h3>';
                    html += '  <div class="json-display" style="color: #e74c3c;">' + execution.error + '</div>';
                    html += '</div>';
                }
                
                executionDetails.innerHTML = html;
            } catch (error) {
                console.error('Failed to load execution details:', error);
            }
        }
        
        function switchTab(tabName) {
            // Update tab buttons
            document.querySelectorAll('.tab-button').forEach(btn => btn.classList.remove('active'));
            document.querySelector('[onclick="switchTab(\'' + tabName + '\')"]').classList.add('active');
            
            // Update tab content
            document.querySelectorAll('.tab-content').forEach(tab => tab.classList.remove('active'));
            document.getElementById(tabName + '-tab').classList.add('active');
            
            // Hide details panel
            selectedRequestId = null;
            document.getElementById('noSelection').style.display = 'flex';
            document.getElementById('requestDetails').style.display = 'none';
            document.getElementById('executionDetails').style.display = 'none';
            
            // Load appropriate data
            if (tabName === 'requests') {
                Promise.all([loadStats(), loadRequests()]);
            } else if (tabName === 'executions') {
                Promise.all([loadExecutionStats(), loadExecutions()]);
            }
            
            // Note: Real-time updates will automatically update the active tab based on the data type
        }
        
        // Initial load
        refreshLogs();
        
        // Clean up on page unload
        window.addEventListener('beforeunload', function() {
            stopRealTimeUpdates();
        });
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprint(w, html); err != nil {
		log.Error().Err(err).Msg("Failed to write HTML response")
	}
}

// serveLogsAPI handles API endpoints for log data
func (ah *AdminHandler) serveLogsAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch {
	case r.URL.Path == "/admin/logs/api/stats":
		ah.handleStatsAPI(w, r)
	case r.URL.Path == "/admin/logs/api/requests":
		ah.handleRequestsAPI(w, r)
	case strings.HasPrefix(r.URL.Path, "/admin/logs/api/requests/"):
		requestID := strings.TrimPrefix(r.URL.Path, "/admin/logs/api/requests/")
		ah.handleRequestDetailsAPI(w, r, requestID)
	case r.URL.Path == "/admin/logs/api/executions":
		ah.handleExecutionsAPI(w, r)
	case strings.HasPrefix(r.URL.Path, "/admin/logs/api/executions/"):
		executionID := strings.TrimPrefix(r.URL.Path, "/admin/logs/api/executions/")
		ah.handleExecutionDetailsAPI(w, r, executionID)
	case r.URL.Path == "/admin/logs/api/clear":
		ah.handleClearLogsAPI(w, r)
	default:
		http.NotFound(w, r)
	}
}

// handleStatsAPI returns logging statistics
func (ah *AdminHandler) handleStatsAPI(w http.ResponseWriter, r *http.Request) {
	stats := ah.logger.GetStats()
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		log.Error().Err(err).Msg("Failed to encode stats response")
	}
}

// handleRequestsAPI returns request logs
func (ah *AdminHandler) handleRequestsAPI(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 50 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	requests := ah.logger.GetRecentRequests(limit)
	if err := json.NewEncoder(w).Encode(requests); err != nil {
		log.Error().Err(err).Msg("Failed to encode requests response")
	}
}

// handleRequestDetailsAPI returns details for a specific request
func (ah *AdminHandler) handleRequestDetailsAPI(w http.ResponseWriter, r *http.Request, requestID string) {
	if request, exists := ah.logger.GetRequestByID(requestID); exists {
		if err := json.NewEncoder(w).Encode(request); err != nil {
			log.Error().Err(err).Msg("Failed to encode request details response")
		}
	} else {
		http.NotFound(w, r)
	}
}

// handleClearLogsAPI clears all logs
func (ah *AdminHandler) handleClearLogsAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ah.logger.ClearLogs()
	log.Info().Msg("Request logs cleared via admin interface")

	response := map[string]interface{}{
		"success": true,
		"message": "Logs cleared successfully",
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error().Err(err).Msg("Failed to encode clear logs response")
	}
}

// handleExecutionsAPI returns script execution history
func (ah *AdminHandler) handleExecutionsAPI(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 50 // default
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offsetStr := r.URL.Query().Get("offset")
	offset := 0 // default
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	filter := repository.ExecutionFilter{
		Search: r.URL.Query().Get("search"),
	}

	pagination := repository.PaginationOptions{
		Limit:  limit,
		Offset: offset,
	}

	result, err := ah.repos.Executions().ListExecutions(context.Background(), filter, pagination)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch script executions")
		http.Error(w, "Failed to fetch executions", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Error().Err(err).Msg("Failed to encode executions response")
	}
}

// handleExecutionDetailsAPI returns details for a specific script execution
func (ah *AdminHandler) handleExecutionDetailsAPI(w http.ResponseWriter, r *http.Request, executionIDStr string) {
	executionID, err := strconv.Atoi(executionIDStr)
	if err != nil {
		http.Error(w, "Invalid execution ID", http.StatusBadRequest)
		return
	}

	execution, err := ah.repos.Executions().GetExecution(context.Background(), executionID)
	if err != nil {
		log.Error().Err(err).Int("executionID", executionID).Msg("Failed to fetch script execution")
		http.NotFound(w, r)
		return
	}

	if err := json.NewEncoder(w).Encode(execution); err != nil {
		log.Error().Err(err).Msg("Failed to encode execution details response")
	}
}

// serveSSE handles Server-Sent Events for real-time updates
func (ah *AdminHandler) serveSSE(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create unique client ID
	clientID := fmt.Sprintf("client_%d", time.Now().UnixNano())
	clientChan := make(chan string, 10)

	// Register client
	ah.sseMutex.Lock()
	ah.sseClients[clientID] = clientChan
	ah.sseMutex.Unlock()

	// Clean up on disconnect
	defer func() {
		ah.sseMutex.Lock()
		delete(ah.sseClients, clientID)
		close(clientChan)
		ah.sseMutex.Unlock()
		log.Debug().Str("clientID", clientID).Msg("SSE client disconnected")
	}()

	log.Debug().Str("clientID", clientID).Msg("SSE client connected")

	// Send initial ping
	fmt.Fprintf(w, "data: {\"type\":\"connected\",\"clientId\":\"%s\"}\n\n", clientID)
	w.(http.Flusher).Flush()

	// Listen for messages or client disconnect
	for {
		select {
		case message, ok := <-clientChan:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", message)
			w.(http.Flusher).Flush()
		case <-r.Context().Done():
			return
		}
	}
}

// broadcastSSE sends a message to all connected SSE clients
func (ah *AdminHandler) broadcastSSE(message string) {
	ah.sseMutex.RLock()
	defer ah.sseMutex.RUnlock()

	for clientID, clientChan := range ah.sseClients {
		select {
		case clientChan <- message:
		default:
			// Channel is full, skip this client
			log.Warn().Str("clientID", clientID).Msg("SSE client channel full, skipping message")
		}
	}
}

// monitorNewRequests watches for new HTTP requests and broadcasts updates
func (ah *AdminHandler) monitorNewRequests() {
	lastRequestCount := 0
	lastExecutionCount := 0

	ticker := time.NewTicker(1 * time.Second) // Check every second
	defer ticker.Stop()

	for range ticker.C {
		// Check for new HTTP requests
		stats := ah.logger.GetStats()
		if totalRequests, ok := stats["totalRequests"].(int); ok && totalRequests > lastRequestCount {
			message := fmt.Sprintf("{\"type\":\"newRequest\",\"count\":%d}", totalRequests)
			ah.broadcastSSE(message)
			lastRequestCount = totalRequests
		}

		// Check for new script executions
		if result, err := ah.repos.Executions().ListExecutions(context.Background(), repository.ExecutionFilter{}, repository.PaginationOptions{Limit: 1, Offset: 0}); err == nil {
			if result.Total > lastExecutionCount {
				message := fmt.Sprintf("{\"type\":\"newExecution\",\"count\":%d}", result.Total)
				ah.broadcastSSE(message)
				lastExecutionCount = result.Total
			}
		}
	}
}

// serveGlobalStateInterface serves the HTML interface for inspecting and editing globalState
func (ah *AdminHandler) serveGlobalStateInterface(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>GlobalState Inspector - Admin Console</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: #f5f5f5;
            color: #333;
        }
        
        .header {
            background: #2c3e50;
            color: white;
            padding: 1rem 2rem;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        
        .header h1 {
            margin: 0;
        }
        
        .nav-links {
            display: flex;
            gap: 1rem;
        }
        
        .nav-links a {
            color: white;
            text-decoration: none;
            padding: 0.5rem 1rem;
            border-radius: 4px;
            background: rgba(255,255,255,0.1);
        }
        
        .nav-links a:hover {
            background: rgba(255,255,255,0.2);
        }
        
        .controls {
            background: white;
            padding: 1rem 2rem;
            border-bottom: 1px solid #ddd;
            display: flex;
            gap: 1rem;
            align-items: center;
        }
        
        .controls button {
            background: #3498db;
            color: white;
            border: none;
            padding: 0.5rem 1rem;
            border-radius: 4px;
            cursor: pointer;
            font-size: 0.9rem;
        }
        
        .controls button:hover {
            background: #2980b9;
        }
        
        .controls button.success {
            background: #27ae60;
        }
        
        .controls button.success:hover {
            background: #229954;
        }
        
        .controls button.danger {
            background: #e74c3c;
        }
        
        .controls button.danger:hover {
            background: #c0392b;
        }
        
        .main-content {
            padding: 2rem;
            max-width: 1200px;
            margin: 0 auto;
        }
        
        .editor-container {
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            overflow: hidden;
        }
        
        .editor-header {
            background: #f8f9fa;
            padding: 1rem;
            border-bottom: 1px solid #ddd;
            font-weight: bold;
            color: #2c3e50;
        }
        
        .editor {
            position: relative;
        }
        
        #globalStateEditor {
            width: 100%;
            height: 400px;
            border: none;
            padding: 1rem;
            font-family: 'Courier New', monospace;
            font-size: 14px;
            resize: vertical;
            background: #1e1e1e;
            color: #f8f8f2;
        }
        
        .status-bar {
            background: #f8f9fa;
            padding: 0.5rem 1rem;
            border-top: 1px solid #ddd;
            font-size: 0.9rem;
            color: #666;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        
        .validation-status {
            font-weight: bold;
        }
        
        .validation-status.valid {
            color: #27ae60;
        }
        
        .validation-status.invalid {
            color: #e74c3c;
        }
        
        .auto-refresh {
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }
        
        .help-panel {
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            margin-top: 2rem;
            overflow: hidden;
        }
        
        .help-header {
            background: #f8f9fa;
            padding: 1rem;
            border-bottom: 1px solid #ddd;
            font-weight: bold;
            color: #2c3e50;
        }
        
        .help-content {
            padding: 1rem;
        }
        
        .help-content ul {
            margin-left: 1.5rem;
        }
        
        .help-content li {
            margin-bottom: 0.5rem;
        }
        
        .help-content code {
            background: #f8f9fa;
            padding: 0.2rem 0.4rem;
            border-radius: 3px;
            font-family: 'Courier New', monospace;
        }
        
        .notification {
            position: fixed;
            top: 2rem;
            right: 2rem;
            padding: 1rem 1.5rem;
            border-radius: 4px;
            font-weight: bold;
            z-index: 1000;
            transform: translateX(400px);
            transition: transform 0.3s ease;
        }
        
        .notification.show {
            transform: translateX(0);
        }
        
        .notification.success {
            background: #27ae60;
            color: white;
        }
        
        .notification.error {
            background: #e74c3c;
            color: white;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>GlobalState Inspector</h1>
        <div class="nav-links">
            <a href="/admin/logs">Request Logs</a>
            <a href="/admin/scripts">Scripts</a>
            <a href="/">Playground</a>
        </div>
    </div>
    
    <div class="controls">
        <button onclick="refreshGlobalState()">Refresh</button>
        <button onclick="saveGlobalState()" class="success">Save Changes</button>
        <button onclick="resetGlobalState()" class="danger">Reset to {}</button>
        <div class="auto-refresh">
            <input type="checkbox" id="autoRefresh" onchange="toggleAutoRefresh()">
            <label for="autoRefresh">Auto-refresh (5s)</label>
        </div>
    </div>
    
    <div class="main-content">
        <div class="editor-container">
            <div class="editor-header">
                GlobalState JSON Editor
            </div>
            <div class="editor">
                <textarea id="globalStateEditor" placeholder="Loading globalState..."></textarea>
            </div>
            <div class="status-bar">
                <div>
                    <span class="validation-status" id="validationStatus">Valid JSON</span>
                    <span id="errorMessage"></span>
                </div>
                <div>
                    Last updated: <span id="lastUpdated">Never</span>
                </div>
            </div>
        </div>
        
        <div class="help-panel">
            <div class="help-header">
                Help & Usage
            </div>
            <div class="help-content">
                <p>The <code>globalState</code> object persists data across JavaScript executions in the playground. Use it to:</p>
                <ul>
                    <li>Store configuration values</li>
                    <li>Maintain counters or statistics</li>
                    <li>Share data between different endpoints</li>
                    <li>Cache computed results</li>
                </ul>
                <p><strong>Example usage in JavaScript:</strong></p>
                <code>
                    globalState.counter = (globalState.counter || 0) + 1;<br>
                    globalState.users = globalState.users || [];<br>
                    globalState.config = { theme: 'dark', version: '1.0' };
                </code>
            </div>
        </div>
    </div>
    
    <div class="notification" id="notification"></div>

    <script>
        let autoRefreshInterval = null;
        let lastGlobalStateValue = '';
        
        async function refreshGlobalState() {
            try {
                const response = await fetch('/admin/globalstate', {
                    headers: { 'Accept': 'application/json' }
                });
                const data = await response.text();
                
                document.getElementById('globalStateEditor').value = data;
                lastGlobalStateValue = data;
                
                validateJSON();
                updateLastUpdated();
                
            } catch (error) {
                console.error('Failed to refresh globalState:', error);
                showNotification('Failed to refresh globalState', 'error');
            }
        }
        
        async function saveGlobalState() {
            const editor = document.getElementById('globalStateEditor');
            const jsonData = editor.value;
            
            if (!validateJSON()) {
                showNotification('Cannot save invalid JSON', 'error');
                return;
            }
            
            try {
                const formData = new FormData();
                formData.append('globalState', jsonData);
                
                const response = await fetch('/admin/globalstate', {
                    method: 'POST',
                    body: formData
                });
                
                if (response.ok) {
                    lastGlobalStateValue = jsonData;
                    showNotification('GlobalState saved successfully', 'success');
                    updateLastUpdated();
                } else {
                    const error = await response.text();
                    showNotification('Failed to save: ' + error, 'error');
                }
            } catch (error) {
                console.error('Failed to save globalState:', error);
                showNotification('Failed to save globalState', 'error');
            }
        }
        
        async function resetGlobalState() {
            if (confirm('Are you sure you want to reset globalState to an empty object? This cannot be undone.')) {
                document.getElementById('globalStateEditor').value = '{}';
                await saveGlobalState();
            }
        }
        
        function validateJSON() {
            const editor = document.getElementById('globalStateEditor');
            const status = document.getElementById('validationStatus');
            const errorMessage = document.getElementById('errorMessage');
            
            try {
                JSON.parse(editor.value);
                status.textContent = 'Valid JSON';
                status.className = 'validation-status valid';
                errorMessage.textContent = '';
                return true;
            } catch (error) {
                status.textContent = 'Invalid JSON';
                status.className = 'validation-status invalid';
                errorMessage.textContent = ' - ' + error.message;
                return false;
            }
        }
        
        function toggleAutoRefresh() {
            const checkbox = document.getElementById('autoRefresh');
            if (checkbox.checked) {
                autoRefreshInterval = setInterval(refreshGlobalState, 5000);
            } else {
                if (autoRefreshInterval) {
                    clearInterval(autoRefreshInterval);
                    autoRefreshInterval = null;
                }
            }
        }
        
        function updateLastUpdated() {
            document.getElementById('lastUpdated').textContent = new Date().toLocaleTimeString();
        }
        
        function showNotification(message, type) {
            const notification = document.getElementById('notification');
            notification.textContent = message;
            notification.className = 'notification ' + type + ' show';
            
            setTimeout(() => {
                notification.classList.remove('show');
            }, 3000);
        }
        
        // Validate JSON as user types
        document.getElementById('globalStateEditor').addEventListener('input', validateJSON);
        
        // Check for unsaved changes before leaving
        window.addEventListener('beforeunload', (e) => {
            const currentValue = document.getElementById('globalStateEditor').value;
            if (currentValue !== lastGlobalStateValue) {
                e.preventDefault();
                e.returnValue = '';
            }
        });
        
        // Load initial data
        refreshGlobalState();
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}
