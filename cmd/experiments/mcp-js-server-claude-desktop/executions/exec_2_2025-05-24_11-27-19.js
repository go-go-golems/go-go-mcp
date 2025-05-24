// Execution ID: 2
// Timestamp: 2025-05-24T11:27:19-04:00
// Success: false
// Error: SyntaxError: SyntaxError: (anonymous): Line 383:22 Unexpected token class (and 6 more errors)

// Dashboard Application - Main landing page with navigation

console.log('Loading Dashboard Application...');

// Dashboard HTML Interface
registerFileHandler('/dashboard.html', function(req, res) {
    const html = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>JavaScript Playground Dashboard</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            color: #333;
        }
        
        .container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
        }
        
        .header {
            text-align: center;
            color: white;
            margin-bottom: 40px;
        }
        
        .header h1 {
            font-size: 3.5em;
            margin-bottom: 10px;
            text-shadow: 2px 2px 4px rgba(0,0,0,0.3);
        }
        
        .header p {
            font-size: 1.2em;
            opacity: 0.9;
            margin-bottom: 30px;
        }
        
        .features {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 30px;
            margin-bottom: 40px;
        }
        
        .feature-card {
            background: white;
            border-radius: 15px;
            padding: 30px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.2);
            transition: transform 0.3s ease, box-shadow 0.3s ease;
            border: 3px solid transparent;
        }
        
        .feature-card:hover {
            transform: translateY(-10px);
            box-shadow: 0 20px 40px rgba(0,0,0,0.3);
            border-color: #667eea;
        }
        
        .feature-icon {
            font-size: 3em;
            margin-bottom: 20px;
            display: block;
        }
        
        .feature-card h3 {
            font-size: 1.5em;
            margin-bottom: 15px;
            color: #333;
        }
        
        .feature-card p {
            color: #666;
            line-height: 1.6;
            margin-bottom: 20px;
        }
        
        .feature-link {
            display: inline-block;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            text-decoration: none;
            padding: 12px 25px;
            border-radius: 25px;
            font-weight: bold;
            transition: all 0.3s ease;
        }
        
        .feature-link:hover {
            transform: scale(1.05);
            box-shadow: 0 5px 15px rgba(102, 126, 234, 0.4);
        }
        
        .api-section {
            background: white;
            border-radius: 15px;
            padding: 30px;
            margin-bottom: 30px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.2);
        }
        
        .api-section h2 {
            color: #333;
            margin-bottom: 20px;
            border-bottom: 3px solid #667eea;
            padding-bottom: 10px;
        }
        
        .api-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
        }
        
        .api-endpoint {
            background: #f8f9fa;
            border-left: 4px solid #667eea;
            padding: 15px;
            border-radius: 8px;
            font-family: 'Courier New', monospace;
        }
        
        .method {
            font-weight: bold;
            padding: 2px 8px;
            border-radius: 4px;
            font-size: 0.8em;
            margin-right: 10px;
        }
        
        .method.GET {
            background: #d4edda;
            color: #155724;
        }
        
        .method.POST {
            background: #cce5ff;
            color: #004085;
        }
        
        .method.PUT {
            background: #fff3cd;
            color: #856404;
        }
        
        .method.DELETE {
            background: #f8d7da;
            color: #721c24;
        }
        
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
            gap: 20px;
            margin-top: 20px;
        }
        
        .stat-card {
            background: linear-gradient(135deg, #4caf50 0%, #45a049 100%);
            color: white;
            padding: 20px;
            border-radius: 10px;
            text-align: center;
        }
        
        .stat-number {
            font-size: 2.5em;
            font-weight: bold;
            display: block;
        }
        
        .stat-label {
            font-size: 0.9em;
            opacity: 0.9;
        }
        
        .footer {
            text-align: center;
            color: white;
            margin-top: 40px;
            opacity: 0.8;
        }
        
        @media (max-width: 768px) {
            .header h1 {
                font-size: 2.5em;
            }
            
            .features {
                grid-template-columns: 1fr;
            }
            
            .api-grid {
                grid-template-columns: 1fr;
            }
        }
        
        .loading {
            text-align: center;
            color: #666;
            font-style: italic;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🚀 JavaScript Playground</h1>
            <p>Build full-stack web applications with JavaScript, Go, and SQLite</p>
        </div>
        
        <div class="features">
            <div class="feature-card">
                <span class="feature-icon">📝</span>
                <h3>Blog System</h3>
                <p>A complete blogging platform with posts, comments, and a beautiful interface. Create, read, and interact with blog content.</p>
                <a href="/files/blog.html" class="feature-link">Open Blog →</a>
            </div>
            
            <div class="feature-card">
                <span class="feature-icon">✅</span>
                <h3>Todo Manager</h3>
                <p>Full-featured task management with priorities, due dates, completion tracking, and advanced filtering options.</p>
                <a href="/files/todos.html" class="feature-link">Open Todos →</a>
            </div>
            
            <div class="feature-card">
                <span class="feature-icon">🔧</span>
                <h3>API Playground</h3>
                <p>Interactive API testing interface to explore and test all available endpoints with real-time responses.</p>
                <a href="#" onclick="toggleApiSection()" class="feature-link">View APIs →</a>
            </div>
        </div>
        
        <div id="live-stats" class="api-section">
            <h2>📊 Live Statistics</h2>
            <div id="stats-content" class="loading">Loading statistics...</div>
        </div>
        
        <div id="api-documentation" class="api-section" style="display: none;">
            <h2>🔗 Available APIs</h2>
            
            <h3 style="margin-top: 30px; color: #667eea;">Blog API</h3>
            <div class="api-grid">
                <div class="api-endpoint">
                    <span class="method GET">GET</span>
                    <code>/js/blog/posts</code>
                    <p>Get all blog posts</p>
                </div>
                <div class="api-endpoint">
                    <span class="method GET">GET</span>
                    <code>/js/blog/posts/:id</code>
                    <p>Get single post with comments</p>
                </div>
                <div class="api-endpoint">
                    <span class="method POST">POST</span>
                    <code>/js/blog/posts</code>
                    <p>Create new blog post</p>
                </div>
                <div class="api-endpoint">
                    <span class="method POST">POST</span>
                    <code>/js/blog/posts/:id/comments</code>
                    <p>Add comment to post</p>
                </div>
            </div>
            
            <h3 style="margin-top: 30px; color: #667eea;">Todo API</h3>
            <div class="api-grid">
                <div class="api-endpoint">
                    <span class="method GET">GET</span>
                    <code>/js/todos</code>
                    <p>Get all todos (with filtering)</p>
                </div>
                <div class="api-endpoint">
                    <span class="method GET">GET</span>
                    <code>/js/todos/:id</code>
                    <p>Get single todo</p>
                </div>
                <div class="api-endpoint">
                    <span class="method POST">POST</span>
                    <code>/js/todos</code>
                    <p>Create new todo</p>
                </div>
                <div class="api-endpoint">
                    <span class="method PUT">PUT</span>
                    <code>/js/todos/:id</code>
                    <p>Update todo</p>
                </div>
                <div class="api-endpoint">
                    <span class="method DELETE">DELETE</span>
                    <code>/js/todos/:id</code>
                    <p>Delete todo</p>
                </div>
                <div class="api-endpoint">
                    <span class="method GET">GET</span>
                    <code>/js/todos/stats</code>
                    <p>Get todo statistics</p>
                </div>
            </div>
            
            <h3 style="margin-top: 30px; color: #667eea;">Playground API</h3>
            <div class="api-grid">
                <div class="api-endpoint">
                    <span class="method POST">POST</span>
                    <code>/api/execute</code>
                    <p>Execute JavaScript code</p>
                </div>
                <div class="api-endpoint">
                    <span class="method GET">GET</span>
                    <code>/api/executions</code>
                    <p>Get execution history</p>
                </div>
                <div class="api-endpoint">
                    <span class="method GET">GET</span>
                    <code>/api/handlers</code>
                    <p>Get registered handlers</p>
                </div>
                <div class="api-endpoint">
                    <span class="method GET">GET</span>
                    <code>/api/state</code>
                    <p>Get global state</p>
                </div>
            </div>
        </div>
        
        <div class="footer">
            <p>Built with ❤️ using JavaScript runtime in Go</p>
            <p>Powered by Goja, Gorilla Mux, and SQLite</p>
        </div>
    </div>

    <script>
        async function loadLiveStats() {
            try {
                // Load blog stats
                const blogResponse = await fetch('/js/blog/posts');
                const blogData = await blogResponse.json();
                
                // Load todo stats
                const todoResponse = await fetch('/js/todos/stats');
                const todoData = await todoResponse.json();
                
                // Load handlers
                const handlersResponse = await fetch('/api/handlers');
                const handlersData = await handlersResponse.json();
                
                // Load executions
                const executionsResponse = await fetch('/api/executions');
                const executionsData = await executionsResponse.json();
                
                displayStats({
                    blogPosts: blogData.success ? blogData.posts.length : 0,
                    todoStats: todoData.success ? todoData.stats : { total: 0, completed: 0, pending: 0 },
                    totalHandlers: handlersData.length || 0,
                    totalExecutions: executionsData.length || 0
                });
                
            } catch (error) {
                document.getElementById('stats-content').innerHTML = 
                    '<p style="color: #d32f2f;">Error loading statistics: ' + error.message + '</p>';
            }
        }
        
        function displayStats(stats) {
            const completionRate = stats.todoStats.total > 0 ? 
                Math.round((stats.todoStats.completed / stats.todoStats.total) * 100) : 0;
            
            document.getElementById('stats-content').innerHTML = `
                <div class="stats-grid">
                    <div class="stat-card">
                        <span class="stat-number">${stats.blogPosts}</span>
                        <span class="stat-label">Blog Posts</span>
                    </div>
                    <div class="stat-card">
                        <span class="stat-number">${stats.todoStats.total}</span>
                        <span class="stat-label">Total Todos</span>
                    </div>
                    <div class="stat-card">
                        <span class="stat-number">${stats.todoStats.pending}</span>
                        <span class="stat-label">Pending Tasks</span>
                    </div>
                    <div class="stat-card">
                        <span class="stat-number">${completionRate}%</span>
                        <span class="stat-label">Completion Rate</span>
                    </div>
                    <div class="stat-card">
                        <span class="stat-number">${stats.totalHandlers}</span>
                        <span class="stat-label">API Endpoints</span>
                    </div>
                    <div class="stat-card">
                        <span class="stat-number">${stats.totalExecutions}</span>
                        <span class="stat-label">JS Executions</span>
                    </div>
                </div>
            `;
        }
        
        function toggleApiSection() {
            const apiSection = document.getElementById('api-documentation');
            if (apiSection.style.display === 'none') {
                apiSection.style.display = 'block';
                apiSection.scrollIntoView({ behavior: 'smooth' });
            } else {
                apiSection.style.display = 'none';
            }
        }
        
        // Load stats on page load and refresh every 30 seconds
        loadLiveStats();
        setInterval(loadLiveStats, 30000);
    </script>
</body>
</html>
    `;
    
    res.html(html);
}, 'text/html');

// Redirect root to dashboard
registerHandler('GET', '/dashboard', function(req, res) {
    res.header('Location', '/files/dashboard.html');
    res.status(302);
    res.text('');
});

console.log('Dashboard Application loaded successfully!');
console.log('- Main dashboard: /files/dashboard.html');
console.log('- Root redirect: /js/dashboard');