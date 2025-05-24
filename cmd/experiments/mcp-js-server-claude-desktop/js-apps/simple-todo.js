// Simple Todo Application

console.log('Loading Simple Todo Application...');

// Initialize todo database table
db.exec('CREATE TABLE IF NOT EXISTS todos (id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT NOT NULL, completed BOOLEAN DEFAULT FALSE, created_at DATETIME DEFAULT CURRENT_TIMESTAMP)');

// Seed some initial todos if empty
var existingTodos = db.query('SELECT COUNT(*) as count FROM todos');
if (existingTodos[0].count === 0) {
    console.log('Seeding initial todo data...');
    db.exec("INSERT INTO todos (title) VALUES ('Learn JavaScript Runtime')");
    db.exec("INSERT INTO todos (title) VALUES ('Build REST API')");
    db.exec("INSERT INTO todos (title) VALUES ('Test Database Integration')");
}

// Get all todos API
registerHandler('GET', '/todos', function(req, res) {
    try {
        var todos = db.query('SELECT id, title, completed, datetime(created_at, "localtime") as created_at FROM todos ORDER BY completed ASC, created_at DESC');
        
        res.json({
            success: true,
            todos: todos
        });
    } catch (error) {
        res.json({
            success: false,
            error: error.message
        });
    }
});

// Create new todo API
registerHandler('POST', '/todos', function(req, res) {
    try {
        var data = JSON.parse(req.body);
        
        if (!data.title) {
            res.status(400);
            res.json({
                success: false,
                error: 'Title is required'
            });
            return;
        }
        
        db.exec('INSERT INTO todos (title) VALUES (?)', data.title);
        
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

// Update todo API
registerHandler('PUT', '/todos/:id', function(req, res) {
    try {
        var todoId = req.path.split('/').pop();
        var data = JSON.parse(req.body);
        
        if (data.completed !== undefined) {
            db.exec('UPDATE todos SET completed = ? WHERE id = ?', data.completed ? 1 : 0, todoId);
        }
        
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

// Delete todo API
registerHandler('DELETE', '/todos/:id', function(req, res) {
    try {
        var todoId = req.path.split('/').pop();
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

// Simple HTML interface
registerFileHandler('/todos.html', function(req, res) {
    var html = '<!DOCTYPE html><html><head><title>Simple Todo</title><style>body{font-family:Arial,sans-serif;max-width:600px;margin:0 auto;padding:20px}h1{color:#333}.todo{display:flex;align-items:center;padding:10px;margin:5px 0;border:1px solid #ddd;border-radius:5px}.todo.completed{opacity:0.6;text-decoration:line-through}.todo input[type="checkbox"]{margin-right:10px}.todo-title{flex:1}.todo-actions{display:flex;gap:5px}button{padding:5px 10px;border:none;border-radius:3px;cursor:pointer}.delete{background:#f44336;color:white}.form{background:#f5f5f5;padding:20px;margin:20px 0;border-radius:5px}input{width:100%;margin:5px 0;padding:8px}.add-btn{background:#4caf50;color:white;padding:10px 20px}</style></head><body>';
    html += '<h1>Simple Todo Manager</h1>';
    html += '<div class="form">';
    html += '<input type="text" id="todoTitle" placeholder="Enter new todo...">';
    html += '<button class="add-btn" onclick="createTodo()">Add Todo</button>';
    html += '</div>';
    html += '<div id="todos"></div>';
    html += '<script>';
    html += 'function loadTodos(){fetch("/js/todos").then(r=>r.json()).then(data=>{if(data.success){var html="";data.todos.forEach(todo=>{html+="<div class=\\"todo"+(todo.completed?" completed":"")+"\\"><input type=\\"checkbox\\""+(todo.completed?" checked":"")+" onchange=\\"toggleTodo("+todo.id+",this.checked)\\"><span class=\\"todo-title\\">"+todo.title+"</span><div class=\\"todo-actions\\"><button class=\\"delete\\" onclick=\\"deleteTodo("+todo.id+")\\">&times;</button></div></div>"});document.getElementById("todos").innerHTML=html}});}';
    html += 'function createTodo(){var title=document.getElementById("todoTitle").value;if(!title){alert("Please enter a todo");return;}fetch("/js/todos",{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({title:title})}).then(r=>r.json()).then(data=>{if(data.success){document.getElementById("todoTitle").value="";loadTodos()}else{alert("Error: "+data.error)}});}';
    html += 'function toggleTodo(id,completed){fetch("/js/todos/"+id,{method:"PUT",headers:{"Content-Type":"application/json"},body:JSON.stringify({completed:completed})}).then(r=>r.json()).then(data=>{if(!data.success){loadTodos();alert("Error: "+data.error)}});}';
    html += 'function deleteTodo(id){if(confirm("Delete this todo?")){fetch("/js/todos/"+id,{method:"DELETE"}).then(r=>r.json()).then(data=>{if(data.success){loadTodos()}else{alert("Error: "+data.error)}});}}';
    html += 'loadTodos();';
    html += '</script></body></html>';
    
    res.html(html);
}, 'text/html');

console.log('Simple Todo Application loaded successfully!');
console.log('- API: /js/todos (GET, POST, PUT, DELETE)');
console.log('- Interface: /files/todos.html');