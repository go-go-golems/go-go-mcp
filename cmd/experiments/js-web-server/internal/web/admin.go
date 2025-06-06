package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/internal/engine"
	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/internal/repository"
	"github.com/rs/zerolog/log"
)

// AdminHandler provides admin endpoints for managing the server
type AdminHandler struct {
	logger *engine.RequestLogger
	repos  repository.RepositoryManager
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(logger *engine.RequestLogger, repos repository.RepositoryManager) *AdminHandler {
	return &AdminHandler{
		logger: logger,
		repos:  repos,
	}
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

	http.NotFound(w, r)
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
                autoRefreshInterval = setInterval(refreshLogs, 5000);
            } else {
                if (autoRefreshInterval) {
                    clearInterval(autoRefreshInterval);
                    autoRefreshInterval = null;
                }
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
        }
        
        // Initial load
        refreshLogs();
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
