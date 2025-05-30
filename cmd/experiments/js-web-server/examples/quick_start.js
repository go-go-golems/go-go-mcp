// Quick Start Examples - Get productive in 5 minutes

// 1. Hello World
app.get('/hello', (req, res) => {
    res.json({ message: 'Hello from JavaScript sandbox!' });
});

// 2. Echo endpoint with request info
app.post('/echo', (req, res) => {
    res.json({
        received: req.body,
        headers: req.headers,
        query: req.query,
        timestamp: new Date().toISOString()
    });
});

// 3. Simple database setup and usage
if (!globalState.initialized) {
    db.query(`
        CREATE TABLE IF NOT EXISTS todos (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            text TEXT NOT NULL,
            done BOOLEAN DEFAULT 0,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )
    `);
    globalState.initialized = true;
    console.log('Database initialized');
}

// 4. CRUD operations
app.get('/todos', (req, res) => {
    const todos = db.query('SELECT * FROM todos ORDER BY created_at DESC');
    res.json(todos);
});

app.post('/todos', (req, res) => {
    const { text } = req.body;
    if (!text) return res.status(400).json({ error: 'Text required' });
    
    db.query('INSERT INTO todos (text) VALUES (?)', [text]);
    const todos = db.query('SELECT * FROM todos WHERE text = ? ORDER BY id DESC LIMIT 1', [text]);
    res.status(201).json(todos[0]);
});

app.put('/todos/:id', (req, res) => {
    const { id } = req.params;
    const { done } = req.body;
    
    db.query('UPDATE todos SET done = ? WHERE id = ?', [done, id]);
    const todos = db.query('SELECT * FROM todos WHERE id = ?', [id]);
    res.json(todos[0]);
});

app.delete('/todos/:id', (req, res) => {
    const { id } = req.params;
    db.query('DELETE FROM todos WHERE id = ?', [id]);
    res.status(204).end();
});

console.log('Quick start examples loaded!');
console.log('Try: GET /hello, POST /echo, GET /todos');
