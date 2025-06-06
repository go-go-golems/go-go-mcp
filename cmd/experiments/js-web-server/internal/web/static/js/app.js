// JavaScript Playground App
class JSPlaygroundApp {
    constructor() {
        this.editor = null;
        this.replHistory = [];
        this.replHistoryIndex = -1;
        this.vimMode = true;
        this.init();
    }

    init() {
        // Initialize based on current page
        if (document.getElementById('editor')) {
            this.initPlayground();
        }
        if (document.getElementById('replConsole')) {
            this.initREPL();
        }
        
        // Initialize common components
        this.initToasts();
        this.loadFromLocalStorage();
    }

    // Playground functionality
    initPlayground() {
        const editorElement = document.getElementById('editor');
        if (!editorElement) return;

        // Initialize CodeMirror
        this.editor = CodeMirror.fromTextArea(editorElement, {
            mode: 'javascript',
            theme: 'darcula',
            lineNumbers: true,
            matchBrackets: true,
            autoCloseBrackets: true,
            indentUnit: 2,
            tabSize: 2,
            keyMap: this.vimMode ? 'vim' : 'default',
            extraKeys: {
                'Ctrl-Enter': () => this.runCode(),
                'Cmd-Enter': () => this.runCode(),
                'Ctrl-S': () => this.executeAndStore(),
                'Cmd-S': () => this.executeAndStore()
            }
        });

        // Bind events
        document.getElementById('runBtn').addEventListener('click', () => this.runCode());
        document.getElementById('executeBtn').addEventListener('click', () => this.executeAndStore());
        document.getElementById('clearBtn').addEventListener('click', () => this.clearEditor());
        document.getElementById('clearOutputBtn').addEventListener('click', () => this.clearOutput());
        document.getElementById('vimModeToggle').addEventListener('change', (e) => this.toggleVimMode(e.target.checked));
        document.getElementById('fontSizeRange').addEventListener('input', (e) => this.changeFontSize(e.target.value));

        // Load presets dropdown
        this.loadPresetsMenu();
        
        // Load saved code or set default
        const savedCode = localStorage.getItem('playgroundCode');
        if (savedCode) {
            this.editor.setValue(savedCode);
            // Don't remove immediately - let it persist for better UX
            setTimeout(() => localStorage.removeItem('playgroundCode'), 1000);
        } else if (editorElement.hasAttribute('data-default-code')) {
            const defaultCode = `// Welcome to the JavaScript Playground!
// This editor supports Vim keybindings and JavaScript syntax highlighting.

// Example: Create an API endpoint
app.get("/hello", (req, res) => {
    res.json({ message: "Hello, World!", timestamp: new Date() });
});

// Example: Database query
const users = db.query("SELECT COUNT(*) as count FROM script_executions");
console.log("Total executions:", users[0].count);

// Try running this code with the "Run" button for local execution
// or "Execute & Store" to save it to the database`;
            this.editor.setValue(defaultCode);
        }
    }

    // REPL functionality
    initREPL() {
        const replInput = document.getElementById('replInput');
        const replConsole = document.getElementById('replConsole');
        
        if (!replInput || !replConsole) return;

        // Setup input handling
        replInput.addEventListener('keydown', (e) => this.handleReplInput(e));
        document.getElementById('execReplBtn').addEventListener('click', () => this.executeRepl());
        document.getElementById('clearReplBtn').addEventListener('click', () => this.clearRepl());
        document.getElementById('resetVmBtn').addEventListener('click', () => this.resetVM());

        // Setup example buttons
        document.querySelectorAll('.repl-example').forEach(btn => {
            btn.addEventListener('click', (e) => {
                const code = e.target.dataset.code;
                replInput.value = code;
                replInput.focus();
            });
        });

        // Load saved code
        const savedCode = localStorage.getItem('replCode');
        if (savedCode) {
            replInput.value = savedCode;
            localStorage.removeItem('replCode');
        }

        // Auto-resize input
        this.autoResizeTextarea(replInput);
    }

    // Code execution
    async runCode() {
        if (!this.editor) return;
        
        const code = this.editor.getValue().trim();
        if (!code) return;

        this.setStatus('Running...', 'warning', true);
        const startTime = Date.now();

        try {
            // For "run" we just execute without storing
            const response = await fetch('/v1/execute', {
                method: 'POST',
                headers: { 'Content-Type': 'text/plain' },
                body: code
            });

            const result = await response.json();
            const duration = Date.now() - startTime;

            if (result.success) {
                this.showResult(result.result, result.consoleLog, null, duration);
                this.setStatus('Execution completed', 'success');
            } else {
                this.showResult(null, [], result.error, duration);
                this.setStatus('Execution failed', 'danger');
            }
        } catch (error) {
            const duration = Date.now() - startTime;
            this.showResult(null, [], `Network error: ${error.message}`, duration);
            this.setStatus('Network error', 'danger');
        }
    }

    async executeAndStore() {
        if (!this.editor) return;
        
        const code = this.editor.getValue().trim();
        if (!code) return;

        this.setStatus('Executing and storing...', 'info', true);
        const startTime = Date.now();

        try {
            const response = await fetch('/v1/execute', {
                method: 'POST',
                headers: { 'Content-Type': 'text/plain' },
                body: code
            });

            const result = await response.json();
            const duration = Date.now() - startTime;

            if (result.success) {
                this.showResult(result.result, result.consoleLog, null, duration, result.sessionID);
                this.setStatus('Stored in database', 'success');
                this.showToast('Code executed and stored successfully', 'success');
            } else {
                this.showResult(null, [], result.error, duration);
                this.setStatus('Execution failed', 'danger');
                this.showToast('Execution failed', 'danger');
            }
        } catch (error) {
            const duration = Date.now() - startTime;
            this.showResult(null, [], `Network error: ${error.message}`, duration);
            this.setStatus('Network error', 'danger');
            this.showToast('Network error', 'danger');
        }
    }

    // REPL execution
    async executeRepl() {
        const replInput = document.getElementById('replInput');
        const code = replInput.value.trim();
        
        if (!code) return;

        this.addReplEntry('input', code);
        this.replHistory.push(code);
        this.replHistoryIndex = this.replHistory.length;

        try {
            const response = await fetch('/v1/execute', {
                method: 'POST',
                headers: { 'Content-Type': 'text/plain' },
                body: code
            });

            const result = await response.json();

            if (result.success) {
                if (result.consoleLog && result.consoleLog.length > 0) {
                    result.consoleLog.forEach(log => {
                        this.addReplEntry('log', log);
                    });
                }
                if (result.result !== undefined) {
                    this.addReplEntry('result', this.formatValue(result.result));
                }
            } else {
                this.addReplEntry('error', result.error);
            }
        } catch (error) {
            this.addReplEntry('error', `Network error: ${error.message}`);
        }

        replInput.value = '';
        this.autoResizeTextarea(replInput);
    }

    // REPL input handling
    handleReplInput(e) {
        const input = e.target;
        
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            this.executeRepl();
        } else if (e.key === 'ArrowUp') {
            e.preventDefault();
            this.navigateReplHistory(-1);
        } else if (e.key === 'ArrowDown') {
            e.preventDefault();
            this.navigateReplHistory(1);
        }
        
        // Auto-resize
        setTimeout(() => this.autoResizeTextarea(input), 0);
    }

    navigateReplHistory(direction) {
        const input = document.getElementById('replInput');
        const newIndex = this.replHistoryIndex + direction;
        
        if (newIndex >= 0 && newIndex < this.replHistory.length) {
            this.replHistoryIndex = newIndex;
            input.value = this.replHistory[this.replHistoryIndex];
        } else if (newIndex >= this.replHistory.length) {
            this.replHistoryIndex = this.replHistory.length;
            input.value = '';
        }
        
        this.autoResizeTextarea(input);
    }

    // REPL console management
    addReplEntry(type, content) {
        const console = document.getElementById('replConsole');
        const entry = document.createElement('div');
        entry.className = `repl-${type}`;
        
        const prefix = type === 'input' ? '> ' : type === 'error' ? '✗ ' : type === 'result' ? '← ' : '  ';
        entry.textContent = prefix + content;
        
        console.appendChild(entry);
        console.scrollTop = console.scrollHeight;
    }

    clearRepl() {
        const console = document.getElementById('replConsole');
        console.innerHTML = `
            <div class="text-success">JavaScript REPL - Type JavaScript expressions and press Enter</div>
            <div class="text-muted">Use Shift+Enter for multi-line input</div>
            <div class="text-muted">Available: app, db, console, globalState</div>
            <div class="mb-2"></div>
        `;
        this.replHistory = [];
        this.replHistoryIndex = -1;
    }

    async resetVM() {
        try {
            await fetch('/api/reset-vm', { method: 'POST' });
            this.addReplEntry('log', 'VM reset successfully');
            this.showToast('VM reset', 'info');
        } catch (error) {
            this.addReplEntry('error', `Failed to reset VM: ${error.message}`);
        }
    }

    // UI utilities
    showResult(result, consoleLog, error, duration, sessionId) {
        // Update console output
        const consoleOutput = document.getElementById('consoleOutput');
        if (consoleLog && consoleLog.length > 0) {
            consoleOutput.innerHTML = consoleLog.map(log => 
                `<div class="repl-log">${this.escapeHtml(log)}</div>`
            ).join('');
        } else {
            consoleOutput.innerHTML = '<div class="text-muted">No console output</div>';
        }

        // Update result output
        const resultOutput = document.getElementById('resultOutput');
        if (error) {
            resultOutput.innerHTML = `<div class="repl-error">${this.escapeHtml(error)}</div>`;
        } else if (result !== undefined) {
            resultOutput.innerHTML = `<div class="repl-result">${this.escapeHtml(this.formatValue(result))}</div>`;
        } else {
            resultOutput.innerHTML = '<div class="text-muted">No result</div>';
        }

        // Update execution time
        document.getElementById('executionTime').textContent = `${duration}ms`;

        // Update session info
        if (sessionId) {
            document.getElementById('sessionId').textContent = sessionId;
            document.getElementById('sessionInfo').style.display = 'block';
        }
    }

    setStatus(text, type, loading = false) {
        const statusText = document.getElementById('statusText');
        const icon = loading ? 'bi-arrow-repeat' : 
                    type === 'success' ? 'bi-check-circle-fill' :
                    type === 'danger' ? 'bi-x-circle-fill' :
                    type === 'warning' ? 'bi-exclamation-triangle-fill' :
                    'bi-info-circle-fill';

        const color = type === 'success' ? 'text-success' :
                     type === 'danger' ? 'text-danger' :
                     type === 'warning' ? 'text-warning' :
                     'text-info';

        statusText.innerHTML = `<i class="bi ${icon} ${color} ${loading ? 'status-running' : ''}"></i> ${text}`;
    }

    clearEditor() {
        if (this.editor) {
            this.editor.setValue('');
        }
    }

    clearOutput() {
        document.getElementById('consoleOutput').innerHTML = '<div class="text-muted">Console output will appear here...</div>';
        document.getElementById('resultOutput').innerHTML = '<div class="text-muted">Execution result will appear here...</div>';
        document.getElementById('sessionInfo').style.display = 'none';
        this.setStatus('Ready', 'success');
    }

    toggleVimMode(enabled) {
        this.vimMode = enabled;
        if (this.editor) {
            this.editor.setOption('keyMap', enabled ? 'vim' : 'default');
        }
        localStorage.setItem('vimMode', enabled);
    }

    changeFontSize(size) {
        if (this.editor) {
            const wrapper = this.editor.getWrapperElement();
            wrapper.style.fontSize = size + 'px';
            this.editor.refresh();
        }
        localStorage.setItem('fontSize', size);
    }

    // Utility functions
    formatValue(value) {
        if (typeof value === 'object') {
            return JSON.stringify(value, null, 2);
        }
        return String(value);
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    // Load presets into dropdown menu
    async loadPresetsMenu() {
        const presetsMenu = document.getElementById('presetsMenu');
        if (!presetsMenu) return;

        try {
            // Get presets list - using hardcoded list for now (could be fetched from API)
            const presets = [
                { id: 'hello-world', name: 'Hello World', description: 'Basic console output example' },
                { id: 'express-basic', name: 'Express Route', description: 'Create a simple Express.js route' },
                { id: 'database-query', name: 'Database Query', description: 'SQLite database query example' },
                { id: 'api-crud', name: 'CRUD API', description: 'Complete CRUD operations example' },
                { id: 'websocket-echo', name: 'WebSocket Echo', description: 'WebSocket server example' },
                { id: 'file-operations', name: 'File Operations', description: 'File read/write operations' },
                { id: 'middleware', name: 'Middleware', description: 'Custom middleware example' },
                { id: 'json-validation', name: 'JSON Validation', description: 'Request validation example' }
            ];

            // Clear existing items except header and divider
            const existingItems = presetsMenu.querySelectorAll('li:not(.dropdown-header):not(:has(hr))');
            existingItems.forEach(item => item.remove());

            // Add preset items
            presets.forEach(preset => {
                const li = document.createElement('li');
                li.innerHTML = `
                    <a class="dropdown-item" href="#" onclick="window.loadPresetExample('${preset.id}'); return false;">
                        <div>
                            <strong>${this.escapeHtml(preset.name)}</strong>
                            <br>
                            <small class="text-muted">${this.escapeHtml(preset.description)}</small>
                        </div>
                    </a>
                `;
                presetsMenu.appendChild(li);
            });
        } catch (error) {
            console.error('Failed to load presets menu:', error);
        }
    }

    autoResizeTextarea(textarea) {
        textarea.style.height = 'auto';
        textarea.style.height = Math.min(textarea.scrollHeight, 150) + 'px';
    }

    // Toast notifications
    initToasts() {
        // Create toast container if it doesn't exist
        if (!document.querySelector('.toast-container')) {
            const container = document.createElement('div');
            container.className = 'toast-container position-fixed top-0 end-0 p-3';
            container.setAttribute('style', 'z-index: 1055');
            document.body.appendChild(container);
        }
    }

    showToast(message, type = 'info', duration = 3000) {
        const container = document.querySelector('.toast-container');
        const toastId = 'toast-' + Date.now();
        
        const bgClass = type === 'success' ? 'bg-success' :
                       type === 'danger' ? 'bg-danger' :
                       type === 'warning' ? 'bg-warning' :
                       'bg-info';

        const toast = document.createElement('div');
        toast.id = toastId;
        toast.className = `toast ${bgClass} text-white`;
        toast.setAttribute('role', 'alert');
        toast.innerHTML = `
            <div class="toast-body d-flex justify-content-between align-items-center">
                <span>${this.escapeHtml(message)}</span>
                <button type="button" class="btn-close btn-close-white" data-bs-dismiss="toast"></button>
            </div>
        `;

        container.appendChild(toast);
        const bsToast = new bootstrap.Toast(toast, { delay: duration });
        bsToast.show();

        // Remove from DOM after hide
        toast.addEventListener('hidden.bs.toast', () => {
            toast.remove();
        });
    }

    // Local storage management
    loadFromLocalStorage() {
        // Load vim mode preference
        const vimMode = localStorage.getItem('vimMode');
        if (vimMode !== null) {
            this.vimMode = vimMode === 'true';
            const toggle = document.getElementById('vimModeToggle');
            if (toggle) toggle.checked = this.vimMode;
        }

        // Load font size preference
        const fontSize = localStorage.getItem('fontSize');
        if (fontSize) {
            const range = document.getElementById('fontSizeRange');
            if (range) {
                range.value = fontSize;
                this.changeFontSize(fontSize);
            }
        }
    }
}

// Initialize app when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    window.jsPlayground = new JSPlaygroundApp();
});

// Global functions for template usage
window.loadToPlayground = function(code) {
    localStorage.setItem('playgroundCode', code);
    window.location.href = '/playground';
};

window.loadToRepl = function(code) {
    localStorage.setItem('replCode', code);
    window.location.href = '/repl';
};

window.copyToClipboard = function(text) {
    navigator.clipboard.writeText(text).then(() => {
        if (window.jsPlayground) {
            window.jsPlayground.showToast('Copied to clipboard', 'success', 1500);
        }
    });
};

window.copySessionId = function(sessionId) {
    navigator.clipboard.writeText(sessionId).then(() => {
        if (window.jsPlayground) {
            window.jsPlayground.showToast('Session ID copied', 'success', 1500);
        }
    });
};

// Load preset example into playground
window.loadPresetExample = async function(presetId) {
    try {
        const response = await fetch(`/api/preset?id=${encodeURIComponent(presetId)}`);
        if (!response.ok) {
            throw new Error('Failed to load preset');
        }
        
        const preset = await response.json();
        localStorage.setItem('playgroundCode', preset.code);
        
        if (window.jsPlayground) {
            window.jsPlayground.showToast(`Loaded preset: ${preset.name}`, 'success', 2000);
        }
        
        // Redirect to playground if not already there
        if (!window.location.pathname.includes('/playground')) {
            window.location.href = '/playground';
        } else {
            // If already on playground, reload the editor
            if (window.jsPlayground && window.jsPlayground.editor) {
                window.jsPlayground.editor.setValue(preset.code);
            }
        }
    } catch (error) {
        if (window.jsPlayground) {
            window.jsPlayground.showToast('Failed to load preset example', 'danger');
        }
        console.error('Error loading preset:', error);
    }
};
