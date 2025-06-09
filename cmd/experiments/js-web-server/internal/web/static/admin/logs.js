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
