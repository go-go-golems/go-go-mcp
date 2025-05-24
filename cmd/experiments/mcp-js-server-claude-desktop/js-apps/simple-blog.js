// Simple Blog Application

console.log('Loading Simple Blog Application...');

// Initialize blog database tables
db.exec('CREATE TABLE IF NOT EXISTS posts (id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT NOT NULL, content TEXT NOT NULL, author TEXT NOT NULL, created_at DATETIME DEFAULT CURRENT_TIMESTAMP)');

// Seed some initial data if empty
var existingPosts = db.query('SELECT COUNT(*) as count FROM posts');
if (existingPosts[0].count === 0) {
    console.log('Seeding initial blog data...');
    db.exec("INSERT INTO posts (title, content, author) VALUES ('Welcome Post', 'This is our first post!', 'Admin')");
    db.exec("INSERT INTO posts (title, content, author) VALUES ('Second Post', 'Another great post here.', 'User')");
}

// Get all posts API
registerHandler('GET', '/blog/posts', function(req, res) {
    try {
        var posts = db.query('SELECT id, title, content, author, datetime(created_at, "localtime") as created_at FROM posts ORDER BY created_at DESC');
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

// Create new post API
registerHandler('POST', '/blog/posts', function(req, res) {
    try {
        var data = JSON.parse(req.body);
        
        if (!data.title || !data.content || !data.author) {
            res.status(400);
            res.json({
                success: false,
                error: 'Title, content, and author are required'
            });
            return;
        }
        
        db.exec('INSERT INTO posts (title, content, author) VALUES (?, ?, ?)', data.title, data.content, data.author);
        
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

// Simple HTML interface
registerFileHandler('/blog.html', function(req, res) {
    var html = '<!DOCTYPE html><html><head><title>Simple Blog</title><style>body{font-family:Arial,sans-serif;max-width:800px;margin:0 auto;padding:20px}h1{color:#333}.post{border:1px solid #ddd;padding:15px;margin:10px 0;border-radius:5px}.form{background:#f5f5f5;padding:20px;margin:20px 0;border-radius:5px}input,textarea{width:100%;margin:5px 0;padding:8px}button{background:#007acc;color:white;padding:10px 20px;border:none;border-radius:3px;cursor:pointer}button:hover{background:#005c99}</style></head><body>';
    html += '<h1>Simple Blog</h1>';
    html += '<div class="form"><h3>Create New Post</h3>';
    html += '<input type="text" id="title" placeholder="Post Title">';
    html += '<input type="text" id="author" placeholder="Your Name">';
    html += '<textarea id="content" placeholder="Post content..." rows="4"></textarea>';
    html += '<button onclick="createPost()">Create Post</button></div>';
    html += '<div id="posts"></div>';
    html += '<script>';
    html += 'function loadPosts(){fetch("/js/blog/posts").then(r=>r.json()).then(data=>{if(data.success){var html="";data.posts.forEach(post=>{html+="<div class=\\"post\\"><h3>"+post.title+"</h3><p>"+post.content+"</p><small>By "+post.author+" on "+post.created_at+"</small></div>"});document.getElementById("posts").innerHTML=html}});}';
    html += 'function createPost(){var title=document.getElementById("title").value;var author=document.getElementById("author").value;var content=document.getElementById("content").value;if(!title||!author||!content){alert("Please fill all fields");return;}fetch("/js/blog/posts",{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({title:title,author:author,content:content})}).then(r=>r.json()).then(data=>{if(data.success){document.getElementById("title").value="";document.getElementById("author").value="";document.getElementById("content").value="";loadPosts();alert("Post created!")}else{alert("Error: "+data.error)}});}';
    html += 'loadPosts();';
    html += '</script></body></html>';
    
    res.html(html);
}, 'text/html');

console.log('Simple Blog Application loaded successfully!');
console.log('- API: /js/blog/posts (GET, POST)');
console.log('- Interface: /files/blog.html');