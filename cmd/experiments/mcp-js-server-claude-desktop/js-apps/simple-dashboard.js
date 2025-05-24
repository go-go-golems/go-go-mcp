// Simple Dashboard Application

console.log('Loading Simple Dashboard Application...');

// Dashboard HTML Interface
registerFileHandler('/dashboard.html', function(req, res) {
    var html = '<!DOCTYPE html><html><head><title>JavaScript Playground</title><style>body{font-family:Arial,sans-serif;max-width:800px;margin:0 auto;padding:20px;background:#f5f5f5}h1{color:#333;text-align:center;margin-bottom:30px}.apps{display:grid;grid-template-columns:repeat(auto-fit,minmax(250px,1fr));gap:20px;margin-bottom:30px}.app-card{background:white;padding:20px;border-radius:10px;box-shadow:0 2px 10px rgba(0,0,0,0.1);text-align:center;transition:transform 0.2s}.app-card:hover{transform:translateY(-5px)}.app-icon{font-size:3em;margin-bottom:10px}.app-link{display:inline-block;background:#007acc;color:white;text-decoration:none;padding:10px 20px;border-radius:5px;margin-top:10px}.app-link:hover{background:#005c99}.stats{background:white;padding:20px;border-radius:10px;box-shadow:0 2px 10px rgba(0,0,0,0.1)}.stat-grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(150px,1fr));gap:15px;margin-top:15px}.stat-item{background:#f8f9fa;padding:15px;border-radius:5px;text-align:center}.stat-number{font-size:2em;font-weight:bold;color:#007acc}.stat-label{font-size:0.9em;color:#666}</style></head><body>';
    html += '<h1>🚀 JavaScript Playground Dashboard</h1>';
    html += '<div class="apps">';
    html += '<div class="app-card">';
    html += '<div class="app-icon">📝</div>';
    html += '<h3>Blog System</h3>';
    html += '<p>Create and manage blog posts with a simple interface.</p>';
    html += '<a href="/files/blog.html" class="app-link">Open Blog</a>';
    html += '</div>';
    html += '<div class="app-card">';
    html += '<div class="app-icon">✅</div>';
    html += '<h3>Todo Manager</h3>';
    html += '<p>Manage your tasks with this simple todo application.</p>';
    html += '<a href="/files/todos.html" class="app-link">Open Todos</a>';
    html += '</div>';
    html += '</div>';
    html += '<div class="stats">';
    html += '<h2>📊 Live Statistics</h2>';
    html += '<div id="stats-content">Loading...</div>';
    html += '</div>';
    html += '<script>';
    html += 'function loadStats(){Promise.all([fetch("/js/blog/posts"),fetch("/js/todos")]).then(responses=>Promise.all(responses.map(r=>r.json()))).then(([blogData,todoData])=>{var blogCount=blogData.success?blogData.posts.length:0;var todoCount=todoData.success?todoData.todos.length:0;var completedTodos=todoData.success?todoData.todos.filter(t=>t.completed).length:0;document.getElementById("stats-content").innerHTML="<div class=\\"stat-grid\\"><div class=\\"stat-item\\"><div class=\\"stat-number\\">"+blogCount+"</div><div class=\\"stat-label\\">Blog Posts</div></div><div class=\\"stat-item\\"><div class=\\"stat-number\\">"+todoCount+"</div><div class=\\"stat-label\\">Total Todos</div></div><div class=\\"stat-item\\"><div class=\\"stat-number\\">"+completedTodos+"</div><div class=\\"stat-label\\">Completed</div></div><div class=\\"stat-item\\"><div class=\\"stat-number\\">"+(todoCount>0?Math.round((completedTodos/todoCount)*100):0)+"%</div><div class=\\"stat-label\\">Progress</div></div></div>"}).catch(err=>{document.getElementById("stats-content").innerHTML="Error loading stats: "+err.message});}';
    html += 'loadStats();setInterval(loadStats,30000);';
    html += '</script></body></html>';
    
    res.html(html);
}, 'text/html');

// Redirect root to dashboard
registerHandler('GET', '/', function(req, res) {
    res.header('Location', '/files/dashboard.html');
    res.status(302);
    res.text('');
});

console.log('Simple Dashboard Application loaded successfully!');
console.log('- Dashboard: /files/dashboard.html');
console.log('- Root redirect: /js/');