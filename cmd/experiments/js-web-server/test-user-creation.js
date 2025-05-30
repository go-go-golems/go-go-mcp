// Test user creation directly

app.post('/test/create-user', (req, res) => {
    console.log('=== USER CREATION TEST ===');
    console.log('Request body:', JSON.stringify(req.body, null, 2));
    
    const { name, email } = req.body;
    
    if (!name || !email) {
        return res.status(400).json({ error: 'Name and email are required' });
    }
    
    try {
        // Insert user directly
        db.query('INSERT INTO users (name, email) VALUES (?, ?)', [name, email]);
        console.log('User inserted successfully');
        
        // Get the created user
        const users = db.query('SELECT id, name, email, created_at FROM users WHERE email = ?', [email]);
        console.log('Retrieved user:', JSON.stringify(users[0], null, 2));
        
        res.status(201).json(users[0]);
    } catch (error) {
        console.error('Error creating user:', error);
        res.status(500).json({ error: 'Failed to create user', details: error.message });
    }
});

console.log('Test user creation endpoint loaded: POST /test/create-user');
