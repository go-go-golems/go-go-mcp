// Task Manager Web Application
// Demonstrates full CRUD operations with HTML forms and JavaScript
// Features: Task management with priority, status, categories, and notes

console.log("🚀 Task Manager Web Application Starting...");

// Initialize database schema
function initializeDatabase() {
    try {
        // Tasks table with comprehensive fields
        db.query(`
            CREATE TABLE IF NOT EXISTS tasks (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                title TEXT NOT NULL CHECK(length(title) >= 3),
                description TEXT,
                priority TEXT DEFAULT 'medium' CHECK(priority IN ('low', 'medium', 'high', 'urgent')),
                status TEXT DEFAULT 'todo' CHECK(status IN ('todo', 'in_progress', 'completed', 'cancelled')),
                category TEXT DEFAULT 'general',
                due_date DATE,
                created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                completed_at DATETIME
            )
        `);

        // Categories table
        db.query(`
            CREATE TABLE IF NOT EXISTS categories (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                name TEXT UNIQUE NOT NULL,
                color TEXT DEFAULT '#007bff',
                description TEXT,
                created_at DATETIME DEFAULT CURRENT_TIMESTAMP
            )
        `);

        // Task comments/notes
        db.query(`
            CREATE TABLE IF NOT EXISTS task_notes (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                task_id INTEGER NOT NULL,
                note TEXT NOT NULL,
                created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
                FOREIGN KEY (task_id) REFERENCES tasks (id) ON DELETE CASCADE
            )
        `);

        // Insert default categories
        const defaultCategories = [
            ['Work', '#dc3545', 'Work-related tasks'],
            ['Personal', '#28a745', 'Personal activities and goals'],
            ['Shopping', '#ffc107', 'Shopping lists and purchases'],
            ['Health', '#17a2b8', 'Health and fitness activities'],
            ['Learning', '#6f42c1', 'Educational and skill development']
        ];

        defaultCategories.forEach(([name, color, description]) => {
            try {
                db.query(
                    "INSERT OR IGNORE INTO categories (name, color, description) VALUES (?, ?, ?)",
                    [name, color, description]
                );
            } catch (err) {
                // Ignore duplicate errors
            }
        });

        console.log("✅ Database schema initialized successfully");
    } catch (error) {
        console.error("❌ Database initialization error:", error);
    }
}

// Initialize application state
if (!globalState.taskManager) {
    globalState.taskManager = {
        initialized: false,
        stats: {
            totalTasks: 0,
            completedTasks: 0,
            activeTasks: 0
        }
    };
}

if (!globalState.taskManager.initialized) {
    initializeDatabase();
    globalState.taskManager.initialized = true;
}

// Main application page - serves the complete HTML interface
app.get("/", (req, res) => {
    const html = `
<!DOCTYPE html>
<html lang="en" data-bs-theme="dark">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Task Manager - Full-Stack Demo</title>
    <!-- Bootstrap CSS -->
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <!-- Bootstrap Icons -->
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bootstrap-icons@1.11.0/font/bootstrap-icons.css">
    <style>
        .priority-high { border-left: 4px solid #dc3545; }
        .priority-urgent { border-left: 4px solid #fd7e14; background-color: #fff3cd; color: #856404; }
        .priority-medium { border-left: 4px solid #ffc107; }
        .priority-low { border-left: 4px solid #6c757d; }
        .status-completed { opacity: 0.7; text-decoration: line-through; }
        .task-card { transition: all 0.2s ease; }
        .task-card:hover { transform: translateY(-2px); box-shadow: 0 4px 12px rgba(0,0,0,0.3); }
        .stats-card { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); }
        .category-badge { font-size: 0.75rem; }
        .notes-section { background-color: rgba(108, 117, 125, 0.1); border-radius: 0.375rem; }
    </style>
</head>
<body>
    <!-- Navigation -->
    <nav class="navbar navbar-expand-lg navbar-dark bg-primary">
        <div class="container">
            <a class="navbar-brand" href="#"><i class="bi bi-check2-square"></i> Task Manager</a>
            <button class="btn btn-outline-light btn-sm" onclick="showStats()">
                <i class="bi bi-graph-up"></i> Stats
            </button>
        </div>
    </nav>

    <div class="container my-4">
        <!-- Stats Dashboard -->
        <div class="row mb-4">
            <div class="col-md-12">
                <div class="card stats-card text-white">
                    <div class="card-body">
                        <div class="row" id="statsRow">
                            <div class="col-md-3 text-center">
                                <h4 id="totalTasks">0</h4>
                                <small>Total Tasks</small>
                            </div>
                            <div class="col-md-3 text-center">
                                <h4 id="activeTasks">0</h4>
                                <small>Active Tasks</small>
                            </div>
                            <div class="col-md-3 text-center">
                                <h4 id="completedTasks">0</h4>
                                <small>Completed</small>
                            </div>
                            <div class="col-md-3 text-center">
                                <h4 id="completionRate">0%</h4>
                                <small>Completion Rate</small>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Task Form -->
        <div class="row mb-4">
            <div class="col-md-12">
                <div class="card">
                    <div class="card-header">
                        <h5 class="mb-0"><i class="bi bi-plus-circle"></i> Add New Task</h5>
                    </div>
                    <div class="card-body">
                        <form id="taskForm">
                            <div class="row">
                                <div class="col-md-6">
                                    <div class="mb-3">
                                        <label for="title" class="form-label">Title *</label>
                                        <input type="text" class="form-control" id="title" required 
                                               placeholder="Enter task title" minlength="3">
                                    </div>
                                </div>
                                <div class="col-md-3">
                                    <div class="mb-3">
                                        <label for="priority" class="form-label">Priority</label>
                                        <select class="form-select" id="priority">
                                            <option value="low">🟢 Low</option>
                                            <option value="medium" selected>🟡 Medium</option>
                                            <option value="high">🔴 High</option>
                                            <option value="urgent">🚨 Urgent</option>
                                        </select>
                                    </div>
                                </div>
                                <div class="col-md-3">
                                    <div class="mb-3">
                                        <label for="category" class="form-label">Category</label>
                                        <select class="form-select" id="category">
                                            <option value="general">General</option>
                                        </select>
                                    </div>
                                </div>
                            </div>
                            <div class="row">
                                <div class="col-md-9">
                                    <div class="mb-3">
                                        <label for="description" class="form-label">Description</label>
                                        <textarea class="form-control" id="description" rows="2" 
                                                  placeholder="Optional task description"></textarea>
                                    </div>
                                </div>
                                <div class="col-md-3">
                                    <div class="mb-3">
                                        <label for="dueDate" class="form-label">Due Date</label>
                                        <input type="date" class="form-control" id="dueDate">
                                    </div>
                                </div>
                            </div>
                            <div class="d-flex gap-2">
                                <button type="submit" class="btn btn-primary">
                                    <i class="bi bi-plus"></i> Add Task
                                </button>
                                <button type="button" class="btn btn-secondary" onclick="clearForm()">
                                    <i class="bi bi-x"></i> Clear
                                </button>
                                <button type="button" class="btn btn-info" onclick="loadSampleTasks()">
                                    <i class="bi bi-collection"></i> Load Sample Tasks
                                </button>
                            </div>
                        </form>
                    </div>
                </div>
            </div>
        </div>

        <!-- Filter Controls -->
        <div class="row mb-3">
            <div class="col-md-12">
                <div class="card">
                    <div class="card-body py-2">
                        <div class="row align-items-center">
                            <div class="col-md-3">
                                <select class="form-select form-select-sm" id="statusFilter" onchange="loadTasks()">
                                    <option value="">All Status</option>
                                    <option value="todo">📋 To Do</option>
                                    <option value="in_progress">⚡ In Progress</option>
                                    <option value="completed">✅ Completed</option>
                                    <option value="cancelled">❌ Cancelled</option>
                                </select>
                            </div>
                            <div class="col-md-3">
                                <select class="form-select form-select-sm" id="priorityFilter" onchange="loadTasks()">
                                    <option value="">All Priorities</option>
                                    <option value="urgent">🚨 Urgent</option>
                                    <option value="high">🔴 High</option>
                                    <option value="medium">🟡 Medium</option>
                                    <option value="low">🟢 Low</option>
                                </select>
                            </div>
                            <div class="col-md-3">
                                <select class="form-select form-select-sm" id="categoryFilter" onchange="loadTasks()">
                                    <option value="">All Categories</option>
                                </select>
                            </div>
                            <div class="col-md-3">
                                <input type="text" class="form-control form-control-sm" id="searchFilter" 
                                       placeholder="Search tasks..." onkeyup="debounceSearch()">
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Tasks List -->
        <div class="row">
            <div class="col-md-12">
                <div id="tasksContainer">
                    <!-- Tasks will be loaded here -->
                </div>
            </div>
        </div>
    </div>

    <!-- Task Edit Modal -->
    <div class="modal fade" id="editTaskModal" tabindex="-1">
        <div class="modal-dialog modal-lg">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title"><i class="bi bi-pencil"></i> Edit Task</h5>
                    <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
                </div>
                <div class="modal-body">
                    <form id="editTaskForm">
                        <input type="hidden" id="editTaskId">
                        <div class="row">
                            <div class="col-md-6">
                                <div class="mb-3">
                                    <label for="editTitle" class="form-label">Title *</label>
                                    <input type="text" class="form-control" id="editTitle" required>
                                </div>
                            </div>
                            <div class="col-md-3">
                                <div class="mb-3">
                                    <label for="editPriority" class="form-label">Priority</label>
                                    <select class="form-select" id="editPriority">
                                        <option value="low">🟢 Low</option>
                                        <option value="medium">🟡 Medium</option>
                                        <option value="high">🔴 High</option>
                                        <option value="urgent">🚨 Urgent</option>
                                    </select>
                                </div>
                            </div>
                            <div class="col-md-3">
                                <div class="mb-3">
                                    <label for="editStatus" class="form-label">Status</label>
                                    <select class="form-select" id="editStatus">
                                        <option value="todo">📋 To Do</option>
                                        <option value="in_progress">⚡ In Progress</option>
                                        <option value="completed">✅ Completed</option>
                                        <option value="cancelled">❌ Cancelled</option>
                                    </select>
                                </div>
                            </div>
                        </div>
                        <div class="row">
                            <div class="col-md-9">
                                <div class="mb-3">
                                    <label for="editDescription" class="form-label">Description</label>
                                    <textarea class="form-control" id="editDescription" rows="3"></textarea>
                                </div>
                            </div>
                            <div class="col-md-3">
                                <div class="mb-3">
                                    <label for="editDueDate" class="form-label">Due Date</label>
                                    <input type="date" class="form-control" id="editDueDate">
                                </div>
                            </div>
                        </div>
                        
                        <!-- Notes Section -->
                        <div class="mb-3">
                            <label class="form-label">Task Notes</label>
                            <div class="notes-section p-3">
                                <div id="taskNotes" class="mb-3">
                                    <!-- Notes will be loaded here -->
                                </div>
                                <div class="input-group">
                                    <input type="text" class="form-control" id="newNote" placeholder="Add a note...">
                                    <button type="button" class="btn btn-outline-secondary" onclick="addNote()">
                                        <i class="bi bi-plus"></i> Add Note
                                    </button>
                                </div>
                            </div>
                        </div>
                    </form>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Cancel</button>
                    <button type="button" class="btn btn-primary" onclick="updateTask()">Save Changes</button>
                </div>
            </div>
        </div>
    </div>

    <!-- Bootstrap JS -->
    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"></script>
    
    <script>
        // Global variables
        let editingTaskId = null;
        let searchTimeout = null;
        
        // Initialize app when page loads
        document.addEventListener('DOMContentLoaded', function() {
            console.log('🎯 Initializing Task Manager UI...');
            loadCategories();
            loadTasks();
            updateStats();
            
            // Setup form submission
            document.getElementById('taskForm').addEventListener('submit', function(e) {
                e.preventDefault();
                createTask();
            });
        });
        
        // Load categories into dropdowns
        async function loadCategories() {
            try {
                const response = await fetch('/api/categories');
                const categories = await response.json();
                
                const categorySelects = ['category', 'categoryFilter'];
                categorySelects.forEach(selectId => {
                    const select = document.getElementById(selectId);
                    // Clear existing options (except "All Categories" for filter)
                    if (selectId === 'categoryFilter') {
                        select.innerHTML = '<option value="">All Categories</option>';
                    } else {
                        select.innerHTML = '<option value="general">General</option>';
                    }
                    
                    categories.forEach(cat => {
                        const option = document.createElement('option');
                        option.value = cat.name.toLowerCase();
                        option.textContent = cat.name;
                        if (selectId === 'categoryFilter') {
                            option.style.color = cat.color;
                        }
                        select.appendChild(option);
                    });
                });
            } catch (error) {
                console.error('Error loading categories:', error);
            }
        }
        
        // Create new task
        async function createTask() {
            const formData = {
                title: document.getElementById('title').value.trim(),
                description: document.getElementById('description').value.trim(),
                priority: document.getElementById('priority').value,
                category: document.getElementById('category').value,
                due_date: document.getElementById('dueDate').value || null
            };
            
            if (!formData.title || formData.title.length < 3) {
                alert('Title must be at least 3 characters long');
                return;
            }
            
            try {
                const response = await fetch('/api/tasks', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(formData)
                });
                
                const result = await response.json();
                
                if (response.ok) {
                    showAlert('Task created successfully!', 'success');
                    clearForm();
                    loadTasks();
                    updateStats();
                } else {
                    showAlert('Error: ' + result.error, 'danger');
                }
            } catch (error) {
                showAlert('Network error: ' + error.message, 'danger');
            }
        }
        
        // Load and display tasks
        async function loadTasks() {
            try {
                const params = new URLSearchParams();
                const status = document.getElementById('statusFilter').value;
                const priority = document.getElementById('priorityFilter').value;
                const category = document.getElementById('categoryFilter').value;
                const search = document.getElementById('searchFilter').value.trim();
                
                if (status) params.append('status', status);
                if (priority) params.append('priority', priority);
                if (category) params.append('category', category);
                if (search) params.append('search', search);
                
                const response = await fetch('/api/tasks?' + params.toString());
                const data = await response.json();
                
                displayTasks(data.tasks || []);
            } catch (error) {
                console.error('Error loading tasks:', error);
                showAlert('Error loading tasks', 'danger');
            }
        }
        
        // Display tasks in the UI
        function displayTasks(tasks) {
            const container = document.getElementById('tasksContainer');
            
            if (tasks.length === 0) {
                container.innerHTML = \`
                    <div class="card">
                        <div class="card-body text-center py-5">
                            <i class="bi bi-inbox display-1 text-muted"></i>
                            <h4 class="text-muted mt-3">No tasks found</h4>
                            <p class="text-muted">Create your first task using the form above!</p>
                        </div>
                    </div>
                \`;
                return;
            }
            
            const tasksHtml = tasks.map(task => {
                const priorityClass = \`priority-\${task.priority}\`;
                const statusClass = task.status === 'completed' ? 'status-completed' : '';
                const dueDateDisplay = task.due_date ? 
                    \`<small class="text-muted"><i class="bi bi-calendar"></i> \${new Date(task.due_date).toLocaleDateString()}</small>\` : '';
                
                const priorityIcon = {
                    'urgent': '🚨',
                    'high': '🔴',
                    'medium': '🟡',
                    'low': '🟢'
                }[task.priority] || '⚪';
                
                const statusIcon = {
                    'todo': '📋',
                    'in_progress': '⚡',
                    'completed': '✅',
                    'cancelled': '❌'
                }[task.status] || '📋';
                
                return \`
                    <div class="card task-card mb-3 \${priorityClass} \${statusClass}">
                        <div class="card-body">
                            <div class="d-flex justify-content-between align-items-start">
                                <div class="flex-grow-1">
                                    <h6 class="card-title mb-1">
                                        \${priorityIcon} \${task.title}
                                        <span class="badge category-badge ms-2" style="background-color: var(--bs-secondary)">
                                            \${task.category}
                                        </span>
                                    </h6>
                                    \${task.description ? \`<p class="card-text small text-muted mb-2">\${task.description}</p>\` : ''}
                                    <div class="d-flex justify-content-between align-items-center">
                                        <div>
                                            <span class="badge bg-secondary me-2">\${statusIcon} \${task.status.replace('_', ' ')}</span>
                                            \${dueDateDisplay}
                                        </div>
                                        <small class="text-muted">
                                            Created: \${new Date(task.created_at).toLocaleDateString()}
                                        </small>
                                    </div>
                                </div>
                                <div class="btn-group ms-3" role="group">
                                    <button class="btn btn-sm btn-outline-primary" onclick="editTask(\${task.id})" title="Edit">
                                        <i class="bi bi-pencil"></i>
                                    </button>
                                    \${task.status !== 'completed' ? 
                                        \`<button class="btn btn-sm btn-outline-success" onclick="markCompleted(\${task.id})" title="Mark Complete">
                                            <i class="bi bi-check"></i>
                                        </button>\` : ''
                                    }
                                    <button class="btn btn-sm btn-outline-danger" onclick="deleteTask(\${task.id})" title="Delete">
                                        <i class="bi bi-trash"></i>
                                    </button>
                                </div>
                            </div>
                        </div>
                    </div>
                \`;
            }).join('');
            
            container.innerHTML = tasksHtml;
        }
        
        // Edit task - open modal with task data
        async function editTask(taskId) {
            try {
                const response = await fetch(\`/api/tasks/\${taskId}\`);
                const task = await response.json();
                
                if (response.ok) {
                    editingTaskId = taskId;
                    
                    // Populate form fields
                    document.getElementById('editTaskId').value = task.id;
                    document.getElementById('editTitle').value = task.title;
                    document.getElementById('editDescription').value = task.description || '';
                    document.getElementById('editPriority').value = task.priority;
                    document.getElementById('editStatus').value = task.status;
                    document.getElementById('editDueDate').value = task.due_date || '';
                    
                    // Load task notes
                    loadTaskNotes(taskId);
                    
                    // Show modal
                    new bootstrap.Modal(document.getElementById('editTaskModal')).show();
                } else {
                    showAlert('Error loading task: ' + task.error, 'danger');
                }
            } catch (error) {
                showAlert('Network error: ' + error.message, 'danger');
            }
        }
        
        // Update task
        async function updateTask() {
            if (!editingTaskId) return;
            
            const formData = {
                title: document.getElementById('editTitle').value.trim(),
                description: document.getElementById('editDescription').value.trim(),
                priority: document.getElementById('editPriority').value,
                status: document.getElementById('editStatus').value,
                due_date: document.getElementById('editDueDate').value || null
            };
            
            try {
                const response = await fetch(\`/api/tasks/\${editingTaskId}\`, {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(formData)
                });
                
                const result = await response.json();
                
                if (response.ok) {
                    showAlert('Task updated successfully!', 'success');
                    bootstrap.Modal.getInstance(document.getElementById('editTaskModal')).hide();
                    loadTasks();
                    updateStats();
                } else {
                    showAlert('Error: ' + result.error, 'danger');
                }
            } catch (error) {
                showAlert('Network error: ' + error.message, 'danger');
            }
        }
        
        // Mark task as completed
        async function markCompleted(taskId) {
            try {
                const response = await fetch(\`/api/tasks/\${taskId}/complete\`, {
                    method: 'PATCH'
                });
                
                const result = await response.json();
                
                if (response.ok) {
                    showAlert('Task marked as completed!', 'success');
                    loadTasks();
                    updateStats();
                } else {
                    showAlert('Error: ' + result.error, 'danger');
                }
            } catch (error) {
                showAlert('Network error: ' + error.message, 'danger');
            }
        }
        
        // Delete task
        async function deleteTask(taskId) {
            if (!confirm('Are you sure you want to delete this task?')) return;
            
            try {
                const response = await fetch(\`/api/tasks/\${taskId}\`, {
                    method: 'DELETE'
                });
                
                if (response.ok) {
                    showAlert('Task deleted successfully!', 'success');
                    loadTasks();
                    updateStats();
                } else {
                    const result = await response.json();
                    showAlert('Error: ' + result.error, 'danger');
                }
            } catch (error) {
                showAlert('Network error: ' + error.message, 'danger');
            }
        }
        
        // Load task notes
        async function loadTaskNotes(taskId) {
            try {
                const response = await fetch(\`/api/tasks/\${taskId}/notes\`);
                const notes = await response.json();
                
                const notesContainer = document.getElementById('taskNotes');
                if (notes.length === 0) {
                    notesContainer.innerHTML = '<p class="text-muted small mb-0">No notes yet</p>';
                } else {
                    notesContainer.innerHTML = notes.map(note => \`
                        <div class="d-flex justify-content-between align-items-start mb-2 p-2 bg-dark rounded">
                            <span class="small">\${note.note}</span>
                            <div>
                                <small class="text-muted">\${new Date(note.created_at).toLocaleDateString()}</small>
                                <button class="btn btn-sm btn-outline-danger ms-2" onclick="deleteNote(\${note.id})" title="Delete note">
                                    <i class="bi bi-x"></i>
                                </button>
                            </div>
                        </div>
                    \`).join('');
                }
            } catch (error) {
                console.error('Error loading notes:', error);
            }
        }
        
        // Add note to task
        async function addNote() {
            const noteText = document.getElementById('newNote').value.trim();
            if (!noteText || !editingTaskId) return;
            
            try {
                const response = await fetch(\`/api/tasks/\${editingTaskId}/notes\`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ note: noteText })
                });
                
                if (response.ok) {
                    document.getElementById('newNote').value = '';
                    loadTaskNotes(editingTaskId);
                } else {
                    const result = await response.json();
                    showAlert('Error: ' + result.error, 'danger');
                }
            } catch (error) {
                showAlert('Network error: ' + error.message, 'danger');
            }
        }
        
        // Delete note
        async function deleteNote(noteId) {
            try {
                const response = await fetch(\`/api/notes/\${noteId}\`, {
                    method: 'DELETE'
                });
                
                if (response.ok) {
                    loadTaskNotes(editingTaskId);
                } else {
                    const result = await response.json();
                    showAlert('Error: ' + result.error, 'danger');
                }
            } catch (error) {
                showAlert('Network error: ' + error.message, 'danger');
            }
        }
        
        // Update statistics
        async function updateStats() {
            try {
                const response = await fetch('/api/tasks/stats');
                const stats = await response.json();
                
                document.getElementById('totalTasks').textContent = stats.total;
                document.getElementById('activeTasks').textContent = stats.active;
                document.getElementById('completedTasks').textContent = stats.completed;
                document.getElementById('completionRate').textContent = 
                    stats.total > 0 ? Math.round((stats.completed / stats.total) * 100) + '%' : '0%';
            } catch (error) {
                console.error('Error updating stats:', error);
            }
        }
        
        // Show statistics modal
        async function showStats() {
            try {
                const response = await fetch('/api/tasks/detailed-stats');
                const stats = await response.json();
                
                alert(\`📊 Detailed Statistics:
                
Total Tasks: \${stats.total}
Active Tasks: \${stats.active}
Completed Tasks: \${stats.completed}
Cancelled Tasks: \${stats.cancelled}

By Priority:
🚨 Urgent: \${stats.byPriority.urgent || 0}
🔴 High: \${stats.byPriority.high || 0}
🟡 Medium: \${stats.byPriority.medium || 0}
🟢 Low: \${stats.byPriority.low || 0}

By Category:
\${Object.entries(stats.byCategory).map(([cat, count]) => \`• \${cat}: \${count}\`).join('\\n')}
                \`);
            } catch (error) {
                console.error('Error loading detailed stats:', error);
            }
        }
        
        // Load sample tasks for demonstration
        async function loadSampleTasks() {
            if (!confirm('This will create several sample tasks. Continue?')) return;
            
            try {
                const response = await fetch('/api/tasks/sample', {
                    method: 'POST'
                });
                
                const result = await response.json();
                
                if (response.ok) {
                    showAlert(\`\${result.created} sample tasks created!\`, 'success');
                    loadTasks();
                    updateStats();
                } else {
                    showAlert('Error: ' + result.error, 'danger');
                }
            } catch (error) {
                showAlert('Network error: ' + error.message, 'danger');
            }
        }
        
        // Utility functions
        function clearForm() {
            document.getElementById('taskForm').reset();
            document.getElementById('priority').value = 'medium';
        }
        
        function debounceSearch() {
            clearTimeout(searchTimeout);
            searchTimeout = setTimeout(loadTasks, 300);
        }
        
        function showAlert(message, type) {
            // Create and show a temporary alert
            const alertDiv = document.createElement('div');
            alertDiv.className = \`alert alert-\${type} alert-dismissible fade show position-fixed\`;
            alertDiv.style.cssText = 'top: 20px; right: 20px; z-index: 9999; min-width: 300px;';
            alertDiv.innerHTML = \`
                \${message}
                <button type="button" class="btn-close" data-bs-dismiss="alert"></button>
            \`;
            
            document.body.appendChild(alertDiv);
            
            // Auto-remove after 5 seconds
            setTimeout(() => {
                if (alertDiv.parentNode) {
                    alertDiv.remove();
                }
            }, 5000);
        }
    </script>
</body>
</html>
    `;

    res.send(html);
});

// API Routes for task management

// Get all categories
app.get("/api/categories", (req, res) => {
    try {
        const categories = db.query("SELECT * FROM categories ORDER BY name");
        res.json(categories);
    } catch (error) {
        console.error("Categories fetch error:", error);
        res.status(500).json({ error: "Failed to fetch categories" });
    }
});

// Get all tasks with optional filtering
app.get("/api/tasks", (req, res) => {
    try {
        const { status, priority, category, search, limit = 50 } = req.query;
        
        let sql = "SELECT * FROM tasks WHERE 1=1";
        let params = [];
        
        if (status) {
            sql += " AND status = ?";
            params.push(status);
        }
        
        if (priority) {
            sql += " AND priority = ?";
            params.push(priority);
        }
        
        if (category) {
            sql += " AND LOWER(category) = LOWER(?)";
            params.push(category);
        }
        
        if (search) {
            sql += " AND (title LIKE ? OR description LIKE ?)";
            params.push(`%${search}%`, `%${search}%`);
        }
        
        sql += " ORDER BY ";
        sql += "CASE priority WHEN 'urgent' THEN 1 WHEN 'high' THEN 2 WHEN 'medium' THEN 3 WHEN 'low' THEN 4 END, ";
        sql += "created_at DESC LIMIT ?";
        params.push(parseInt(limit));
        
        const tasks = db.query(sql, params);
        res.json({ tasks, count: tasks.length });
    } catch (error) {
        console.error("Tasks fetch error:", error);
        res.status(500).json({ error: "Failed to fetch tasks" });
    }
});

// Get single task by ID
app.get("/api/tasks/:id", (req, res) => {
    try {
        const taskId = parseInt(req.params.id);
        if (isNaN(taskId)) {
            return res.status(400).json({ error: "Invalid task ID" });
        }
        
        const tasks = db.query("SELECT * FROM tasks WHERE id = ?", [taskId]);
        if (tasks.length === 0) {
            return res.status(404).json({ error: "Task not found" });
        }
        
        res.json(tasks[0]);
    } catch (error) {
        console.error("Task fetch error:", error);
        res.status(500).json({ error: "Failed to fetch task" });
    }
});

// Create new task
app.post("/api/tasks", (req, res) => {
    try {
        const { title, description, priority = 'medium', category = 'general', due_date } = req.body;
        
        if (!title || title.trim().length < 3) {
            return res.status(400).json({ error: "Title must be at least 3 characters long" });
        }
        
        if (!['low', 'medium', 'high', 'urgent'].includes(priority)) {
            return res.status(400).json({ error: "Invalid priority level" });
        }
        
        const result = db.query(
            `INSERT INTO tasks (title, description, priority, category, due_date) 
             VALUES (?, ?, ?, ?, ?)`,
            [title.trim(), description?.trim() || null, priority, category, due_date || null]
        );
        
        const newTask = db.query("SELECT * FROM tasks WHERE id = ?", [result.lastInsertRowid])[0];
        
        console.log(`✅ Task created: "${title}" (ID: ${newTask.id})`);
        res.status(201).json(newTask);
    } catch (error) {
        console.error("Task creation error:", error);
        res.status(500).json({ error: "Failed to create task" });
    }
});

// Update task
app.put("/api/tasks/:id", (req, res) => {
    try {
        const taskId = parseInt(req.params.id);
        if (isNaN(taskId)) {
            return res.status(400).json({ error: "Invalid task ID" });
        }
        
        const { title, description, priority, status, category, due_date } = req.body;
        
        if (!title || title.trim().length < 3) {
            return res.status(400).json({ error: "Title must be at least 3 characters long" });
        }
        
        // Check if task exists
        const existingTasks = db.query("SELECT id FROM tasks WHERE id = ?", [taskId]);
        if (existingTasks.length === 0) {
            return res.status(404).json({ error: "Task not found" });
        }
        
        const completedAt = status === 'completed' ? new Date().toISOString() : null;
        
        db.query(
            `UPDATE tasks SET 
             title = ?, description = ?, priority = ?, status = ?, 
             category = ?, due_date = ?, updated_at = CURRENT_TIMESTAMP,
             completed_at = ?
             WHERE id = ?`,
            [title.trim(), description?.trim() || null, priority, status, 
             category, due_date || null, completedAt, taskId]
        );
        
        const updatedTask = db.query("SELECT * FROM tasks WHERE id = ?", [taskId])[0];
        
        console.log(`📝 Task updated: "${title}" (ID: ${taskId})`);
        res.json(updatedTask);
    } catch (error) {
        console.error("Task update error:", error);
        res.status(500).json({ error: "Failed to update task" });
    }
});

// Mark task as completed (PATCH request)
app.patch("/api/tasks/:id/complete", (req, res) => {
    try {
        const taskId = parseInt(req.params.id);
        if (isNaN(taskId)) {
            return res.status(400).json({ error: "Invalid task ID" });
        }
        
        // Check if task exists
        const existingTasks = db.query("SELECT id, status FROM tasks WHERE id = ?", [taskId]);
        if (existingTasks.length === 0) {
            return res.status(404).json({ error: "Task not found" });
        }
        
        if (existingTasks[0].status === 'completed') {
            return res.status(400).json({ error: "Task is already completed" });
        }
        
        db.query(
            "UPDATE tasks SET status = 'completed', completed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
            [taskId]
        );
        
        const updatedTask = db.query("SELECT * FROM tasks WHERE id = ?", [taskId])[0];
        
        console.log(`✅ Task completed: ${updatedTask.title} (ID: ${taskId})`);
        res.json(updatedTask);
    } catch (error) {
        console.error("Task completion error:", error);
        res.status(500).json({ error: "Failed to complete task" });
    }
});

// Delete task
app.delete("/api/tasks/:id", (req, res) => {
    try {
        const taskId = parseInt(req.params.id);
        if (isNaN(taskId)) {
            return res.status(400).json({ error: "Invalid task ID" });
        }
        
        // Check if task exists
        const existingTasks = db.query("SELECT title FROM tasks WHERE id = ?", [taskId]);
        if (existingTasks.length === 0) {
            return res.status(404).json({ error: "Task not found" });
        }
        
        const taskTitle = existingTasks[0].title;
        
        // Delete task (notes will be deleted automatically due to foreign key cascade)
        db.query("DELETE FROM tasks WHERE id = ?", [taskId]);
        
        console.log(`🗑️  Task deleted: "${taskTitle}" (ID: ${taskId})`);
        res.status(204).end();
    } catch (error) {
        console.error("Task deletion error:", error);
        res.status(500).json({ error: "Failed to delete task" });
    }
});

// Get task statistics
app.get("/api/tasks/stats", (req, res) => {
    try {
        const stats = db.query(`
            SELECT 
                COUNT(*) as total,
                COUNT(CASE WHEN status != 'completed' AND status != 'cancelled' THEN 1 END) as active,
                COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed,
                COUNT(CASE WHEN status = 'cancelled' THEN 1 END) as cancelled
            FROM tasks
        `)[0];
        
        res.json(stats);
    } catch (error) {
        console.error("Stats fetch error:", error);
        res.status(500).json({ error: "Failed to fetch statistics" });
    }
});

// Get detailed statistics
app.get("/api/tasks/detailed-stats", (req, res) => {
    try {
        const basicStats = db.query(`
            SELECT 
                COUNT(*) as total,
                COUNT(CASE WHEN status != 'completed' AND status != 'cancelled' THEN 1 END) as active,
                COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed,
                COUNT(CASE WHEN status = 'cancelled' THEN 1 END) as cancelled
            FROM tasks
        `)[0];
        
        const priorityStats = db.query(`
            SELECT priority, COUNT(*) as count 
            FROM tasks 
            GROUP BY priority
        `);
        
        const categoryStats = db.query(`
            SELECT category, COUNT(*) as count 
            FROM tasks 
            GROUP BY category
        `);
        
        const byPriority = {};
        priorityStats.forEach(stat => {
            byPriority[stat.priority] = stat.count;
        });
        
        const byCategory = {};
        categoryStats.forEach(stat => {
            byCategory[stat.category] = stat.count;
        });
        
        res.json({
            ...basicStats,
            byPriority,
            byCategory
        });
    } catch (error) {
        console.error("Detailed stats fetch error:", error);
        res.status(500).json({ error: "Failed to fetch detailed statistics" });
    }
});

// Task notes endpoints

// Get notes for a task
app.get("/api/tasks/:id/notes", (req, res) => {
    try {
        const taskId = parseInt(req.params.id);
        if (isNaN(taskId)) {
            return res.status(400).json({ error: "Invalid task ID" });
        }
        
        const notes = db.query(
            "SELECT * FROM task_notes WHERE task_id = ? ORDER BY created_at DESC",
            [taskId]
        );
        
        res.json(notes);
    } catch (error) {
        console.error("Notes fetch error:", error);
        res.status(500).json({ error: "Failed to fetch notes" });
    }
});

// Add note to task
app.post("/api/tasks/:id/notes", (req, res) => {
    try {
        const taskId = parseInt(req.params.id);
        if (isNaN(taskId)) {
            return res.status(400).json({ error: "Invalid task ID" });
        }
        
        const { note } = req.body;
        if (!note || note.trim().length === 0) {
            return res.status(400).json({ error: "Note text is required" });
        }
        
        // Check if task exists
        const existingTasks = db.query("SELECT id FROM tasks WHERE id = ?", [taskId]);
        if (existingTasks.length === 0) {
            return res.status(404).json({ error: "Task not found" });
        }
        
        const result = db.query(
            "INSERT INTO task_notes (task_id, note) VALUES (?, ?)",
            [taskId, note.trim()]
        );
        
        const newNote = db.query("SELECT * FROM task_notes WHERE id = ?", [result.lastInsertRowid])[0];
        
        res.status(201).json(newNote);
    } catch (error) {
        console.error("Note creation error:", error);
        res.status(500).json({ error: "Failed to create note" });
    }
});

// Delete note
app.delete("/api/notes/:id", (req, res) => {
    try {
        const noteId = parseInt(req.params.id);
        if (isNaN(noteId)) {
            return res.status(400).json({ error: "Invalid note ID" });
        }
        
        // Check if note exists
        const existingNotes = db.query("SELECT id FROM task_notes WHERE id = ?", [noteId]);
        if (existingNotes.length === 0) {
            return res.status(404).json({ error: "Note not found" });
        }
        
        db.query("DELETE FROM task_notes WHERE id = ?", [noteId]);
        
        res.status(204).end();
    } catch (error) {
        console.error("Note deletion error:", error);
        res.status(500).json({ error: "Failed to delete note" });
    }
});

// Create sample tasks for demonstration
app.post("/api/tasks/sample", (req, res) => {
    try {
        const sampleTasks = [
            {
                title: "Complete project documentation",
                description: "Write comprehensive documentation for the task manager project",
                priority: "high",
                category: "work",
                status: "in_progress"
            },
            {
                title: "Grocery shopping",
                description: "Buy ingredients for dinner party: salmon, vegetables, wine",
                priority: "medium",
                category: "personal",
                status: "todo"
            },
            {
                title: "Learn React Hooks",
                description: "Study useState, useEffect, and custom hooks with practical examples",
                priority: "medium",
                category: "learning",
                status: "todo"
            },
            {
                title: "Doctor appointment",
                description: "Annual checkup - remember to bring insurance card",
                priority: "high",
                category: "health",
                status: "todo"
            },
            {
                title: "Fix website bug",
                description: "Resolve the mobile responsive layout issue on the contact page",
                priority: "urgent",
                category: "work",
                status: "todo"
            },
            {
                title: "Read technical book",
                description: "Continue reading 'Clean Code' - currently on chapter 5",
                priority: "low",
                category: "learning",
                status: "in_progress"
            }
        ];
        
        let created = 0;
        sampleTasks.forEach(task => {
            try {
                db.query(
                    `INSERT INTO tasks (title, description, priority, category, status) 
                     VALUES (?, ?, ?, ?, ?)`,
                    [task.title, task.description, task.priority, task.category, task.status]
                );
                created++;
            } catch (err) {
                // Skip if task already exists or other error
                console.log(`Skipped sample task: ${task.title}`);
            }
        });
        
        console.log(`📋 Created ${created} sample tasks`);
        res.json({ created, total: sampleTasks.length });
    } catch (error) {
        console.error("Sample tasks creation error:", error);
        res.status(500).json({ error: "Failed to create sample tasks" });
    }
});

console.log("🎯 Task Manager Web Application Ready!");
console.log("📱 Features: Create, Read, Update, Delete tasks with priorities, categories, and notes");
console.log("🌐 Visit http://localhost:8080/ to access the full-featured task manager");
console.log("🔧 API endpoints available: GET/POST/PUT/PATCH/DELETE /api/tasks/*");
