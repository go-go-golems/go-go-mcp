// API endpoints for the playground

let apiCounter = 0;

registerHandler("GET", "/api/status", () => ({
    status: "running",
    version: "1.0.0",
    requests: ++apiCounter,
    timestamp: new Date().toISOString()
}));

registerHandler("POST", "/api/echo", (req) => ({
    echo: req.body || "No body provided",
    method: req.method,
    path: req.path,
    headers: req.headers
}));

registerHandler("GET", "/api/users", () => {
    const users = db.query("SELECT * FROM users LIMIT 10");
    return {
        users: users || [],
        count: users ? users.length : 0
    };
});

// Create users table if it doesn't exist
try {
    db.query(`CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        email TEXT UNIQUE NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    )`);
    
    // Insert some sample data
    const existingUsers = db.query("SELECT COUNT(*) as count FROM users");
    if (existingUsers && existingUsers[0] && existingUsers[0].count === 0) {
        db.query("INSERT INTO users (name, email) VALUES (?, ?)", ["Alice", "alice@example.com"]);
        db.query("INSERT INTO users (name, email) VALUES (?, ?)", ["Bob", "bob@example.com"]);
        db.query("INSERT INTO users (name, email) VALUES (?, ?)", ["Charlie", "charlie@example.com"]);
        console.log("Sample users created");
    }
} catch (e) {
    console.error("Database setup error:", e);
}

console.log("API endpoints registered");