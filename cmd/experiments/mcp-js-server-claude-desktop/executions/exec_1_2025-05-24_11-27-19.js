// Execution ID: 1
// Timestamp: 2025-05-24T11:27:19-04:00
// Success: false
// Error: SyntaxError: SyntaxError: (anonymous): Line 330:22 Unexpected token class (and 12 more errors)

// Blog Application - A simple blog with posts and comments

console.log('Loading Blog Application...');

// Initialize blog database tables
db.exec(`
    CREATE TABLE IF NOT EXISTS posts (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT NOT NULL,
        content TEXT NOT NULL,
        author TEXT NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    )
`);

db.exec(`
    CREATE TABLE IF NOT EXISTS comments (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        post_id INTEGER NOT NULL,
        author TEXT NOT NULL,
        content TEXT NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY(post_id) REFERENCES posts(id)
    )
`);

// Seed some initial data if empty
const existingPosts = db.query('SELECT COUNT(*) as count FROM posts');
if (existingPosts[0].count === 0) {
    console.log('Seeding initial blog data...');
    
    db.exec(`
        INSERT INTO posts (title, content, author) VALUES 
        ('Welcome to the JavaScript Blog!', 'This is our first post showcasing the power of JavaScript in Go. You can create dynamic web applications with database integration!', 'Admin'),
        ('Building REST APIs with JavaScript', 'Learn how to build powerful REST APIs using our JavaScript runtime. No need for Node.js - everything runs in Go!', 'Developer'),
        ('Database Magic', 'SQLite integration makes it easy to store and retrieve data. Perfect for rapid prototyping and small applications.', 'Data Engineer')
    `);
    
    db.exec(`
        INSERT INTO comments (post_id, author, content) VALUES 
        (1, 'Reader1', 'Amazing! This is so cool!'),
        (1, 'Coder123', 'I love the simplicity of this approach.'),
        (2, 'APIFan', 'This makes building APIs so much easier!'),
        (3, 'DBExpert', 'SQLite is perfect for this use case.')
    `);
}

// Blog API Endpoints

// Get all posts
registerHandler('GET', '/blog/posts', function(req, res) {
    try {
        const posts = db.query(`
            SELECT id, title, content, author, 
                   datetime(created_at, 'localtime') as created_at
            FROM posts 
            ORDER BY created_at DESC
        `);
        
        res.json({
            success: true,
            posts: posts
        });
    } catch (error) {
        res.json({
            success: false,
            error: error.message
        });
    }
});

// Get single post with comments
registerHandler('GET', '/blog/posts/:id', function(req, res) {
    try {
        const postId = req.path.split('/').pop();
        
        const post = db.get(`
            SELECT id, title, content, author, 
                   datetime(created_at, 'localtime') as created_at
            FROM posts 
            WHERE id = ?
        `, postId);
        
        if (!post) {
            res.status(404);
            res.json({ success: false, error: 'Post not found' });
            return;
        }
        
        const comments = db.query(`
            SELECT id, author, content, 
                   datetime(created_at, 'localtime') as created_at
            FROM comments 
            WHERE post_id = ?
            ORDER BY created_at ASC
        `, postId);
        
        res.json({
            success: true,
            post: post,
            comments: comments
        });
    } catch (error) {
        res.json({
            success: false,
            error: error.message
        });
    }
});

// Create new post
registerHandler('POST', '/blog/posts', function(req, res) {
    try {
        const data = JSON.parse(req.body);
        
        if (!data.title || !data.content || !data.author) {
            res.status(400);
            res.json({
                success: false,
                error: 'Title, content, and author are required'
            });
            return;
        }
        
        db.exec(`
            INSERT INTO posts (title, content, author) 
            VALUES (?, ?, ?)
        `, data.title, data.content, data.author);
        
        res.json({
            success: true,
            message: 'Post created successfully'
        });
    } catch (error) {
        res.json({
            success: false,
            error: error.message
        });
    }
});

// Add comment to post
registerHandler('POST', '/blog/posts/:id/comments', function(req, res) {
    try {
        const postId = req.path.split('/')[3]; // /blog/posts/:id/comments
        const data = JSON.parse(req.body);
        
        if (!data.author || !data.content) {
            res.status(400);
            res.json({
                success: false,
                error: 'Author and content are required'
            });
            return;
        }
        
        // Check if post exists
        const post = db.get('SELECT id FROM posts WHERE id = ?', postId);
        if (!post) {
            res.status(404);
            res.json({ success: false, error: 'Post not found' });
            return;
        }
        
        db.exec(`
            INSERT INTO comments (post_id, author, content) 
            VALUES (?, ?, ?)
        `, postId, data.author, data.content);
        
        res.json({
            success: true,
            message: 'Comment added successfully'
        });
    } catch (error) {
        res.json({
            success: false,
            error: error.message
        });
    }
});

// Blog HTML Interface
registerFileHandler('/blog.html', function(req, res) {
    const html = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>JavaScript Blog</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            max-width: 800px;
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
            border-bottom: 3px solid #007acc;
            padding-bottom: 10px;
        }
        .post {
            border: 1px solid #ddd;
            border-radius: 8px;
            padding: 20px;
            margin: 20px 0;
            background: #fafafa;
        }
        .post h3 {
            margin-top: 0;
            color: #007acc;
        }
        .meta {
            color: #666;
            font-size: 0.9em;
            border-top: 1px solid #eee;
            padding-top: 10px;
            margin-top: 15px;
        }
        .comment {
            background: #f0f8ff;
            border-left: 4px solid #007acc;
            padding: 10px 15px;
            margin: 10px 0;
        }
        .comment-author {
            font-weight: bold;
            color: #007acc;
        }
        button {
            background: #007acc;
            color: white;
            border: none;
            padding: 10px 20px;
            border-radius: 5px;
            cursor: pointer;
            margin: 5px;
        }
        button:hover {
            background: #005c99;
        }
        input, textarea {
            width: 100%;
            padding: 10px;
            margin: 5px 0;
            border: 1px solid #ddd;
            border-radius: 4px;
            box-sizing: border-box;
        }
        textarea {
            min-height: 100px;
            resize: vertical;
        }
        .form-section {
            background: #f9f9f9;
            padding: 20px;
            border-radius: 8px;
            margin: 20px 0;
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
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 JavaScript Blog</h1>
        <p><em>Powered by JavaScript running in Go with SQLite</em></p>
        
        <div class="form-section">
            <h3>📝 Create New Post</h3>
            <form id="postForm">
                <input type="text" id="postTitle" placeholder="Post Title" required>
                <input type="text" id="postAuthor" placeholder="Your Name" required>
                <textarea id="postContent" placeholder="Write your post content..." required></textarea>
                <button type="submit">Publish Post</button>
            </form>
        </div>
        
        <div id="message"></div>
        <div id="posts"></div>
    </div>

    <script>
        // Load and display posts
        async function loadPosts() {
            try {
                const response = await fetch('/js/blog/posts');
                const data = await response.json();
                
                if (data.success) {
                    displayPosts(data.posts);
                } else {
                    showMessage('Error loading posts: ' + data.error, 'error');
                }
            } catch (error) {
                showMessage('Network error: ' + error.message, 'error');
            }
        }
        
        function displayPosts(posts) {
            const postsDiv = document.getElementById('posts');
            
            if (posts.length === 0) {
                postsDiv.innerHTML = '<p>No posts yet. Create the first one!</p>';
                return;
            }
            
            postsDiv.innerHTML = posts.map(post => `
                <div class="post">
                    <h3>${escapeHtml(post.title)}</h3>
                    <p>${escapeHtml(post.content)}</p>
                    <div class="meta">
                        By <strong>${escapeHtml(post.author)}</strong> on ${post.created_at}
                        <button onclick="loadComments(${post.id})">View Comments</button>
                    </div>
                    <div id="comments-${post.id}" style="display:none;"></div>
                </div>
            `).join('');
        }
        
        async function loadComments(postId) {
            try {
                const commentsDiv = document.getElementById('comments-' + postId);
                
                if (commentsDiv.style.display === 'block') {
                    commentsDiv.style.display = 'none';
                    return;
                }
                
                const response = await fetch('/js/blog/posts/' + postId);
                const data = await response.json();
                
                if (data.success) {
                    commentsDiv.innerHTML = `
                        <h4>💬 Comments (${data.comments.length})</h4>
                        ${data.comments.map(comment => `
                            <div class="comment">
                                <div class="comment-author">${escapeHtml(comment.author)}</div>
                                <div>${escapeHtml(comment.content)}</div>
                                <small>${comment.created_at}</small>
                            </div>
                        `).join('')}
                        <div style="margin-top: 15px;">
                            <input type="text" id="commentAuthor-${postId}" placeholder="Your name" style="width: 30%; margin-right: 10px;">
                            <input type="text" id="commentContent-${postId}" placeholder="Add a comment..." style="width: 50%; margin-right: 10px;">
                            <button onclick="addComment(${postId})">Add Comment</button>
                        </div>
                    `;
                    commentsDiv.style.display = 'block';
                } else {
                    showMessage('Error loading comments: ' + data.error, 'error');
                }
            } catch (error) {
                showMessage('Network error: ' + error.message, 'error');
            }
        }
        
        async function addComment(postId) {
            const author = document.getElementById('commentAuthor-' + postId).value;
            const content = document.getElementById('commentContent-' + postId).value;
            
            if (!author || !content) {
                showMessage('Please fill in both name and comment', 'error');
                return;
            }
            
            try {
                const response = await fetch('/js/blog/posts/' + postId + '/comments', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({author, content})
                });
                
                const data = await response.json();
                
                if (data.success) {
                    showMessage('Comment added!', 'success');
                    loadComments(postId); // Reload comments
                } else {
                    showMessage('Error: ' + data.error, 'error');
                }
            } catch (error) {
                showMessage('Network error: ' + error.message, 'error');
            }
        }
        
        // Create new post
        document.getElementById('postForm').addEventListener('submit', async function(e) {
            e.preventDefault();
            
            const title = document.getElementById('postTitle').value;
            const author = document.getElementById('postAuthor').value;
            const content = document.getElementById('postContent').value;
            
            try {
                const response = await fetch('/js/blog/posts', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({title, author, content})
                });
                
                const data = await response.json();
                
                if (data.success) {
                    showMessage('Post created successfully!', 'success');
                    this.reset();
                    loadPosts(); // Reload posts
                } else {
                    showMessage('Error: ' + data.error, 'error');
                }
            } catch (error) {
                showMessage('Network error: ' + error.message, 'error');
            }
        });
        
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
        
        // Load posts on page load
        loadPosts();
    </script>
</body>
</html>
    `;
    
    res.html(html);
}, 'text/html');

console.log('Blog Application loaded successfully!');
console.log('- API endpoints: /js/blog/posts (GET, POST)');
console.log('- Single post: /js/blog/posts/:id (GET)');
console.log('- Add comment: /js/blog/posts/:id/comments (POST)');
console.log('- Web interface: /files/blog.html');