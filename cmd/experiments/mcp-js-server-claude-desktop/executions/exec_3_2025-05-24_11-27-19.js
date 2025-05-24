// Execution ID: 3
// Timestamp: 2025-05-24T11:27:19-04:00
// Success: false
// Error: SyntaxError: SyntaxError: (anonymous): Line 570:22 Unexpected token class (and 12 more errors)

// Todo Application - A simple task management system

console.log('Loading Todo Application...');

// Initialize todo database table
db.exec(`
    CREATE TABLE IF NOT EXISTS todos (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT NOT NULL,
        description TEXT,
        completed BOOLEAN DEFAULT FALSE,
        priority TEXT DEFAULT 'medium',
        due_date TEXT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    )
`);

// Seed some initial todos if empty
const existingTodos = db.query('SELECT COUNT(*) as count FROM todos');
if (existingTodos[0].count === 0) {
    console.log('Seeding initial todo data...');
    
    db.exec(`
        INSERT INTO todos (title, description, priority, due_date) VALUES 
        ('Learn JavaScript Runtime', 'Explore the capabilities of running JavaScript in Go', 'high', '2025-01-30'),
        ('Build REST API', 'Create a powerful REST API using the JS playground', 'medium', '2025-02-05'),
        ('Test Database Integration', 'Verify SQLite operations work correctly', 'low', '2025-01-28'),
        ('Create Documentation', 'Write comprehensive docs for the project', 'medium', '2025-02-10')
    `);
}

// Todo API Endpoints

// Get all todos with optional filtering
registerHandler('GET', '/todos', function(req, res) {
    try {
        let query = `
            SELECT id, title, description, completed, priority, due_date,
                   datetime(created_at, 'localtime') as created_at,
                   datetime(updated_at, 'localtime') as updated_at
            FROM todos
        `;
        
        const params = [];
        const whereConditions = [];
        
        // Filter by completion status
        if (req.query.completed !== undefined) {
            whereConditions.push('completed = ?');
            params.push(req.query.completed === 'true' ? 1 : 0);
        }
        
        // Filter by priority
        if (req.query.priority) {
            whereConditions.push('priority = ?');
            params.push(req.query.priority);
        }
        
        if (whereConditions.length > 0) {
            query += ' WHERE ' + whereConditions.join(' AND ');
        }
        
        query += ' ORDER BY completed ASC, priority DESC, due_date ASC';
        
        const todos = db.query(query, ...params);
        
        res.json({
            success: true,
            todos: todos,
            total: todos.length
        });
    } catch (error) {
        res.json({
            success: false,
            error: error.message
        });
    }
});

// Get single todo
registerHandler('GET', '/todos/:id', function(req, res) {
    try {
        const todoId = req.path.split('/').pop();
        
        const todo = db.get(`
            SELECT id, title, description, completed, priority, due_date,
                   datetime(created_at, 'localtime') as created_at,
                   datetime(updated_at, 'localtime') as updated_at
            FROM todos 
            WHERE id = ?
        `, todoId);
        
        if (!todo) {
            res.status(404);
            res.json({ success: false, error: 'Todo not found' });
            return;
        }
        
        res.json({
            success: true,
            todo: todo
        });
    } catch (error) {
        res.json({
            success: false,
            error: error.message
        });
    }
});

// Create new todo
registerHandler('POST', '/todos', function(req, res) {
    try {
        const data = JSON.parse(req.body);
        
        if (!data.title) {
            res.status(400);
            res.json({
                success: false,
                error: 'Title is required'
            });
            return;
        }
        
        const priority = data.priority || 'medium';
        const validPriorities = ['low', 'medium', 'high'];
        
        if (!validPriorities.includes(priority)) {
            res.status(400);
            res.json({
                success: false,
                error: 'Priority must be low, medium, or high'
            });
            return;
        }
        
        db.exec(`
            INSERT INTO todos (title, description, priority, due_date) 
            VALUES (?, ?, ?, ?)
        `, data.title, data.description || '', priority, data.due_date || null);
        
        res.json({
            success: true,
            message: 'Todo created successfully'
        });
    } catch (error) {
        res.json({
            success: false,
            error: error.message
        });
    }
});

// Update todo
registerHandler('PUT', '/todos/:id', function(req, res) {
    try {
        const todoId = req.path.split('/').pop();
        const data = JSON.parse(req.body);
        
        // Check if todo exists
        const existingTodo = db.get('SELECT id FROM todos WHERE id = ?', todoId);
        if (!existingTodo) {
            res.status(404);
            res.json({ success: false, error: 'Todo not found' });
            return;
        }
        
        const updateFields = [];
        const params = [];
        
        if (data.title !== undefined) {
            updateFields.push('title = ?');
            params.push(data.title);
        }
        
        if (data.description !== undefined) {
            updateFields.push('description = ?');
            params.push(data.description);
        }
        
        if (data.completed !== undefined) {
            updateFields.push('completed = ?');
            params.push(data.completed ? 1 : 0);
        }
        
        if (data.priority !== undefined) {
            const validPriorities = ['low', 'medium', 'high'];
            if (!validPriorities.includes(data.priority)) {
                res.status(400);
                res.json({
                    success: false,
                    error: 'Priority must be low, medium, or high'
                });
                return;
            }
            updateFields.push('priority = ?');
            params.push(data.priority);
        }
        
        if (data.due_date !== undefined) {
            updateFields.push('due_date = ?');
            params.push(data.due_date);
        }
        
        if (updateFields.length === 0) {
            res.status(400);
            res.json({
                success: false,
                error: 'No fields to update'
            });
            return;
        }
        
        updateFields.push('updated_at = CURRENT_TIMESTAMP');
        params.push(todoId);
        
        const query = `UPDATE todos SET ${updateFields.join(', ')} WHERE id = ?`;
        db.exec(query, ...params);
        
        res.json({
            success: true,
            message: 'Todo updated successfully'
        });
    } catch (error) {
        res.json({
            success: false,
            error: error.message
        });
    }
});

// Delete todo
registerHandler('DELETE', '/todos/:id', function(req, res) {
    try {
        const todoId = req.path.split('/').pop();
        
        // Check if todo exists
        const existingTodo = db.get('SELECT id FROM todos WHERE id = ?', todoId);
        if (!existingTodo) {
            res.status(404);
            res.json({ success: false, error: 'Todo not found' });
            return;
        }
        
        db.exec('DELETE FROM todos WHERE id = ?', todoId);
        
        res.json({
            success: true,
            message: 'Todo deleted successfully'
        });
    } catch (error) {
        res.json({
            success: false,
            error: error.message
        });
    }
});

// Get todo statistics
registerHandler('GET', '/todos/stats', function(req, res) {
    try {
        const stats = db.get(`
            SELECT 
                COUNT(*) as total,
                SUM(CASE WHEN completed = 1 THEN 1 ELSE 0 END) as completed,
                SUM(CASE WHEN completed = 0 THEN 1 ELSE 0 END) as pending,
                SUM(CASE WHEN priority = 'high' AND completed = 0 THEN 1 ELSE 0 END) as high_priority_pending
            FROM todos
        `);
        
        const priorityBreakdown = db.query(`
            SELECT priority, COUNT(*) as count
            FROM todos
            WHERE completed = 0
            GROUP BY priority
        `);
        
        res.json({
            success: true,
            stats: stats,
            priority_breakdown: priorityBreakdown
        });
    } catch (error) {
        res.json({
            success: false,
            error: error.message
        });
    }
});

// Todo HTML Interface
registerFileHandler('/todos.html', function(req, res) {
    const html = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Todo Manager</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            max-width: 1000px;
            margin: 0 auto;
            padding: 20px;
            line-height: 1.6;
            background-color: #f5f5f5;
        }
        .container {
            background: white;
            padding: 30px;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        h1 {
            color: #333;
            border-bottom: 3px solid #4caf50;
            padding-bottom: 10px;
        }
        .stats {
            display: flex;
            gap: 20px;
            margin: 20px 0;
        }
        .stat-card {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 20px;
            border-radius: 8px;
            text-align: center;
            flex: 1;
        }
        .stat-card h3 {
            margin: 0;
            font-size: 2em;
        }
        .stat-card p {
            margin: 5px 0 0 0;
            opacity: 0.9;
        }
        .todo-form {
            background: #f9f9f9;
            padding: 20px;
            border-radius: 8px;
            margin: 20px 0;
        }
        .form-row {
            display: flex;
            gap: 10px;
            margin: 10px 0;
        }
        .form-row input, .form-row select, .form-row textarea {
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 4px;
            font-size: 14px;
        }
        .form-row input[type="text"] {
            flex: 2;
        }
        .form-row select {
            flex: 1;
        }
        .form-row input[type="date"] {
            flex: 1;
        }
        textarea {
            width: 100%;
            min-height: 60px;
            resize: vertical;
            box-sizing: border-box;
        }
        button {
            background: #4caf50;
            color: white;
            border: none;
            padding: 10px 20px;
            border-radius: 5px;
            cursor: pointer;
            margin: 5px;
            font-size: 14px;
        }
        button:hover {
            background: #45a049;
        }
        button.delete {
            background: #f44336;
        }
        button.delete:hover {
            background: #da190b;
        }
        button.edit {
            background: #2196f3;
        }
        button.edit:hover {
            background: #1976d2;
        }
        .filters {
            display: flex;
            gap: 10px;
            margin: 20px 0;
            align-items: center;
        }
        .todo-item {
            border: 1px solid #ddd;
            border-radius: 8px;
            padding: 15px;
            margin: 10px 0;
            background: white;
            transition: box-shadow 0.2s;
        }
        .todo-item:hover {
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }
        .todo-item.completed {
            opacity: 0.7;
            background: #f0f8f0;
        }
        .todo-header {
            display: flex;
            justify-content: space-between;
            align-items: flex-start;
        }
        .todo-title {
            font-size: 1.2em;
            font-weight: bold;
            margin: 0 0 10px 0;
            color: #333;
        }
        .todo-title.completed {
            text-decoration: line-through;
            color: #666;
        }
        .todo-meta {
            display: flex;
            gap: 15px;
            margin: 10px 0;
            font-size: 0.9em;
            color: #666;
        }
        .priority {
            padding: 2px 8px;
            border-radius: 12px;
            font-size: 0.8em;
            font-weight: bold;
        }
        .priority.high {
            background: #ffebee;
            color: #c62828;
        }
        .priority.medium {
            background: #fff3e0;
            color: #ef6c00;
        }
        .priority.low {
            background: #e8f5e8;
            color: #2e7d32;
        }
        .due-date {
            color: #666;
        }
        .due-date.overdue {
            color: #d32f2f;
            font-weight: bold;
        }
        .todo-actions {
            display: flex;
            gap: 5px;
        }
        .checkbox {
            width: 20px;
            height: 20px;
            margin-right: 10px;
        }
        .error {
            color: #d32f2f;
            background: #ffebee;
            padding: 10px;
            border-radius: 4px;
            margin: 10px 0;
        }
        .success {
            color: #2e7d32;
            background: #e8f5e8;
            padding: 10px;
            border-radius: 4px;
            margin: 10px 0;
        }
        @media (max-width: 768px) {
            .stats {
                flex-direction: column;
            }
            .form-row {
                flex-direction: column;
            }
            .filters {
                flex-direction: column;
                align-items: stretch;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>✅ Todo Manager</h1>
        <p><em>Full-featured task management powered by JavaScript + Go + SQLite</em></p>
        
        <div id="stats" class="stats"></div>
        
        <div class="todo-form">
            <h3>📝 Add New Todo</h3>
            <form id="todoForm">
                <div class="form-row">
                    <input type="text" id="todoTitle" placeholder="Todo title..." required>
                    <select id="todoPriority">
                        <option value="low">Low Priority</option>
                        <option value="medium" selected>Medium Priority</option>
                        <option value="high">High Priority</option>
                    </select>
                    <input type="date" id="todoDueDate">
                </div>
                <textarea id="todoDescription" placeholder="Description (optional)..."></textarea>
                <button type="submit">Add Todo</button>
            </form>
        </div>
        
        <div class="filters">
            <label>Filter:</label>
            <select id="filterCompleted">
                <option value="">All Todos</option>
                <option value="false">Pending Only</option>
                <option value="true">Completed Only</option>
            </select>
            <select id="filterPriority">
                <option value="">All Priorities</option>
                <option value="high">High Priority</option>
                <option value="medium">Medium Priority</option>
                <option value="low">Low Priority</option>
            </select>
            <button onclick="loadTodos()">Refresh</button>
        </div>
        
        <div id="message"></div>
        <div id="todos"></div>
    </div>

    <script>
        let todos = [];
        
        // Load stats
        async function loadStats() {
            try {
                const response = await fetch('/js/todos/stats');
                const data = await response.json();
                
                if (data.success) {
                    displayStats(data.stats, data.priority_breakdown);
                }
            } catch (error) {
                console.error('Error loading stats:', error);
            }
        }
        
        function displayStats(stats, priorityBreakdown) {
            const statsDiv = document.getElementById('stats');
            const completionRate = stats.total > 0 ? Math.round((stats.completed / stats.total) * 100) : 0;
            
            statsDiv.innerHTML = `
                <div class="stat-card">
                    <h3>${stats.total}</h3>
                    <p>Total Tasks</p>
                </div>
                <div class="stat-card">
                    <h3>${stats.pending}</h3>
                    <p>Pending</p>
                </div>
                <div class="stat-card">
                    <h3>${stats.completed}</h3>
                    <p>Completed</p>
                </div>
                <div class="stat-card">
                    <h3>${completionRate}%</h3>
                    <p>Completion Rate</p>
                </div>
            `;
        }
        
        // Load and display todos
        async function loadTodos() {
            try {
                const completedFilter = document.getElementById('filterCompleted').value;
                const priorityFilter = document.getElementById('filterPriority').value;
                
                let url = '/js/todos';
                const params = new URLSearchParams();
                
                if (completedFilter) params.append('completed', completedFilter);
                if (priorityFilter) params.append('priority', priorityFilter);
                
                if (params.toString()) {
                    url += '?' + params.toString();
                }
                
                const response = await fetch(url);
                const data = await response.json();
                
                if (data.success) {
                    todos = data.todos;
                    displayTodos(todos);
                    loadStats();
                } else {
                    showMessage('Error loading todos: ' + data.error, 'error');
                }
            } catch (error) {
                showMessage('Network error: ' + error.message, 'error');
            }
        }
        
        function displayTodos(todos) {
            const todosDiv = document.getElementById('todos');
            
            if (todos.length === 0) {
                todosDiv.innerHTML = '<p>No todos found. Add your first task above!</p>';
                return;
            }
            
            todosDiv.innerHTML = todos.map(todo => {
                const isOverdue = todo.due_date && new Date(todo.due_date) < new Date() && !todo.completed;
                
                return `
                    <div class="todo-item ${todo.completed ? 'completed' : ''}">
                        <div class="todo-header">
                            <div style="flex: 1;">
                                <div class="todo-title ${todo.completed ? 'completed' : ''}">
                                    <input type="checkbox" class="checkbox" ${todo.completed ? 'checked' : ''} 
                                           onchange="toggleTodo(${todo.id}, this.checked)">
                                    ${escapeHtml(todo.title)}
                                </div>
                                ${todo.description ? `<div style="margin: 10px 0;">${escapeHtml(todo.description)}</div>` : ''}
                                <div class="todo-meta">
                                    <span class="priority ${todo.priority}">${todo.priority.toUpperCase()}</span>
                                    ${todo.due_date ? `<span class="due-date ${isOverdue ? 'overdue' : ''}">Due: ${todo.due_date}</span>` : ''}
                                    <span>Created: ${todo.created_at}</span>
                                </div>
                            </div>
                            <div class="todo-actions">
                                <button class="edit" onclick="editTodo(${todo.id})">Edit</button>
                                <button class="delete" onclick="deleteTodo(${todo.id})">Delete</button>
                            </div>
                        </div>
                    </div>
                `;
            }).join('');
        }
        
        // Toggle todo completion
        async function toggleTodo(id, completed) {
            try {
                const response = await fetch('/js/todos/' + id, {
                    method: 'PUT',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({completed})
                });
                
                const data = await response.json();
                
                if (data.success) {
                    loadTodos();
                } else {
                    showMessage('Error: ' + data.error, 'error');
                    loadTodos(); // Reload to reset checkbox
                }
            } catch (error) {
                showMessage('Network error: ' + error.message, 'error');
                loadTodos();
            }
        }
        
        // Delete todo
        async function deleteTodo(id) {
            if (!confirm('Are you sure you want to delete this todo?')) {
                return;
            }
            
            try {
                const response = await fetch('/js/todos/' + id, {
                    method: 'DELETE'
                });
                
                const data = await response.json();
                
                if (data.success) {
                    showMessage('Todo deleted!', 'success');
                    loadTodos();
                } else {
                    showMessage('Error: ' + data.error, 'error');
                }
            } catch (error) {
                showMessage('Network error: ' + error.message, 'error');
            }
        }
        
        // Edit todo (simplified - just prompt for new title)
        async function editTodo(id) {
            const todo = todos.find(t => t.id === id);
            if (!todo) return;
            
            const newTitle = prompt('Edit todo title:', todo.title);
            if (newTitle && newTitle !== todo.title) {
                try {
                    const response = await fetch('/js/todos/' + id, {
                        method: 'PUT',
                        headers: {'Content-Type': 'application/json'},
                        body: JSON.stringify({title: newTitle})
                    });
                    
                    const data = await response.json();
                    
                    if (data.success) {
                        showMessage('Todo updated!', 'success');
                        loadTodos();
                    } else {
                        showMessage('Error: ' + data.error, 'error');
                    }
                } catch (error) {
                    showMessage('Network error: ' + error.message, 'error');
                }
            }
        }
        
        // Create new todo
        document.getElementById('todoForm').addEventListener('submit', async function(e) {
            e.preventDefault();
            
            const title = document.getElementById('todoTitle').value;
            const priority = document.getElementById('todoPriority').value;
            const description = document.getElementById('todoDescription').value;
            const due_date = document.getElementById('todoDueDate').value;
            
            try {
                const response = await fetch('/js/todos', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({title, priority, description, due_date})
                });
                
                const data = await response.json();
                
                if (data.success) {
                    showMessage('Todo created successfully!', 'success');
                    this.reset();
                    loadTodos();
                } else {
                    showMessage('Error: ' + data.error, 'error');
                }
            } catch (error) {
                showMessage('Network error: ' + error.message, 'error');
            }
        });
        
        // Filter event listeners
        document.getElementById('filterCompleted').addEventListener('change', loadTodos);
        document.getElementById('filterPriority').addEventListener('change', loadTodos);
        
        function showMessage(message, type) {
            const messageDiv = document.getElementById('message');
            messageDiv.innerHTML = '<div class="' + type + '">' + escapeHtml(message) + '</div>';
            setTimeout(() => messageDiv.innerHTML = '', 5000);
        }
        
        function escapeHtml(text) {
            const div = document.createElement('div');
            div.textContent = text;
            return div.innerHTML;
        }
        
        // Load todos on page load
        loadTodos();
    </script>
</body>
</html>
    `;
    
    res.html(html);
}, 'text/html');

console.log('Todo Application loaded successfully!');
console.log('- API endpoints: /js/todos (GET, POST)');
console.log('- Single todo: /js/todos/:id (GET, PUT, DELETE)');
console.log('- Statistics: /js/todos/stats (GET)');
console.log('- Web interface: /files/todos.html');