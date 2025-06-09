// Practical Development Patterns

// Initialize database schema if not exists
if (!globalState.practicalPatternsInitialized) {
    // Create users table
    db.query(`
        CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL,
            email TEXT UNIQUE NOT NULL,
            password TEXT DEFAULT 'password123',
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )
    `);

    // Check if we have any users, if not create some sample data
    const existingUsers = db.query('SELECT COUNT(*) as count FROM users');
    if (existingUsers[0].count === 0) {
        // Insert sample users
        const sampleUsers = [
            { name: 'John Doe', email: 'john@example.com' },
            { name: 'Jane Smith', email: 'jane@example.com' },
            { name: 'Bob Johnson', email: 'bob@example.com' }
        ];

        sampleUsers.forEach(user => {
            db.query('INSERT INTO users (name, email) VALUES (?, ?)', [user.name, user.email]);
        });

        console.log(`Initialized users table with ${sampleUsers.length} sample users`);
    }

    globalState.practicalPatternsInitialized = true;
    console.log('Practical patterns database schema initialized');
}

// Pattern 1: Input validation middleware
function validateRequired(fields) {
    return (req, res, next) => {
        const missing = fields.filter(field => !req.body[field]);
        if (missing.length > 0) {
            return res.status(400).json({ 
                error: `Missing required fields: ${missing.join(', ')}` 
            });
        }
        next();
    };
}

// Pattern 2: Error handling wrapper
function asyncHandler(fn) {
    return (req, res, next) => {
        try {
            return fn(req, res, next);
        } catch (error) {
            console.error('Handler error:', error);
            res.status(500).json({ error: 'Internal server error' });
        }
    };
}

// Pattern 3: Simple caching
if (!globalState.cache) {
    globalState.cache = new Map();
}

function getCached(key, ttlMinutes = 5) {
    const cached = globalState.cache.get(key);
    if (cached && cached.expires > Date.now()) {
        return cached.data;
    }
    return null;
}

function setCache(key, data, ttlMinutes = 5) {
    globalState.cache.set(key, {
        data,
        expires: Date.now() + (ttlMinutes * 60 * 1000)
    });
}

// Pattern 4: Database model pattern
globalState.UserModel = {
    findAll: () => db.query('SELECT id, name, email, created_at FROM users'),
    findById: (id) => {
        const users = db.query('SELECT id, name, email, created_at FROM users WHERE id = ?', [id]);
        return users[0] || null;
    },
    create: (data) => {
        db.query('INSERT INTO users (name, email) VALUES (?, ?)', [data.name, data.email]);
        return globalState.UserModel.findByEmail(data.email);
    },
    findByEmail: (email) => {
        const users = db.query('SELECT id, name, email, created_at FROM users WHERE email = ?', [email]);
        return users[0] || null;
    },
    findByEmailWithPassword: (email) => {
        const users = db.query('SELECT id, name, email, password, created_at FROM users WHERE email = ?', [email]);
        return users[0] || null;
    }
};

// Pattern 5: RESTful endpoints with patterns
app.get('/api/users', asyncHandler((req, res) => {
    const cached = getCached('users');
    if (cached) return res.json(cached);
    
    const users = globalState.UserModel.findAll();
    setCache('users', users, 2);
    res.json(users);
}));

app.post('/api/users', validateRequired(['name', 'email']), asyncHandler((req, res) => {
    const user = globalState.UserModel.create(req.body);
    globalState.cache.delete('users'); // Invalidate cache
    res.status(201).json(user);
}));

// Pattern 6: Authentication with sessions
if (!globalState.sessions) {
    globalState.sessions = new Map();
}

app.post('/auth/login', (req, res) => {
    const { email, password } = req.body;
    // Simplified auth - use proper password hashing in production
    const user = globalState.UserModel.findByEmailWithPassword(email);
    
    if (!user || user.password !== password) {
        return res.status(401).json({ error: 'Invalid credentials' });
    }
    
    const token = Math.random().toString(36).substring(2, 15);
    // Remove password from user object before storing in session
    const userForSession = { id: user.id, name: user.name, email: user.email, created_at: user.created_at };
    globalState.sessions.set(token, { userId: user.id, user: userForSession });
    
    res.json({ token, user: userForSession });
});

function requireAuth(req, res, next) {
    const token = req.headers.authorization?.replace('Bearer ', '');
    const session = globalState.sessions.get(token);
    
    if (!session) {
        return res.status(401).json({ error: 'Authentication required' });
    }
    
    req.user = session.user;
    next();
}

app.get('/auth/profile', requireAuth, (req, res) => {
    res.json({ user: req.user });
});

// Pattern 7: Health check and monitoring
if (!globalState.stats) {
    globalState.stats = { requests: 0, errors: 0 };
}

app.use('/', (req, res, next) => {
    globalState.stats.requests++;
    next();
});

app.get('/health', (req, res) => {
    res.json({
        status: 'healthy',
        stats: globalState.stats,
        timestamp: new Date().toISOString()
    });
});

console.log('Practical patterns loaded!');
console.log('Available endpoints: /api/users, /auth/login, /auth/profile, /health');
