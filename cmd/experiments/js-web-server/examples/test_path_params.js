// Express.js-style path parameter routing examples

// Basic user endpoint with path parameter
app.get('/users/:id', (req, res) => {
    console.log(`GET /users/${req.params.id} from ${req.ip}`);
    
    const userId = parseInt(req.params.id);
    if (isNaN(userId)) {
        return res.status(400).json({ error: 'Invalid user ID' });
    }
    
    res.json({
        message: "User endpoint",
        userId: userId,
        path: req.path,
        method: req.method,
        userAgent: req.headers['user-agent']
    });
});

// Multiple path parameters
app.get('/api/:version/users/:userId/posts/:postId', (req, res) => {
    const { version, userId, postId } = req.params;
    
    res.json({
        message: "Complex path parameters",
        api: {
            version: version,
            endpoint: `users/${userId}/posts/${postId}`
        },
        params: req.params,
        query: req.query
    });
});

// Practical trivia game answer endpoint
app.get('/trivia/answer/:answerIndex', (req, res) => {
    const answerIndex = parseInt(req.params.answerIndex);
    
    if (isNaN(answerIndex) || answerIndex < 0) {
        return res.status(400).json({ 
            error: 'Answer index must be a positive number' 
        });
    }
    
    // Simulate checking answer (in real app, get from database)
    const correctAnswer = 2;
    const isCorrect = answerIndex === correctAnswer;
    
    res.json({
        message: "Answer submitted",
        answerIndex: answerIndex,
        isCorrect: isCorrect,
        result: isCorrect ? 'Correct!' : 'Try again'
    });
});

// Product catalog with optional category
app.get('/products/:category?', (req, res) => {
    const category = req.params.category || 'all';
    
    res.json({
        message: `Products in ${category} category`,
        category: category,
        filters: req.query
    });
});

// RESTful resource pattern
app.get('/blog/:slug', (req, res) => {
    const { slug } = req.params;
    
    // In real app, query database
    res.json({
        post: {
            slug: slug,
            title: `Blog post: ${slug}`,
            content: 'This would come from the database...'
        }
    });
});

console.log("Express.js path parameter handlers registered!");
console.log("Try: GET /users/123, /api/v1/users/456/posts/789, /trivia/answer/2");