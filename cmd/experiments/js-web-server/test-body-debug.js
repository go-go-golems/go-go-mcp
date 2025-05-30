// Debug request body parsing

app.post('/debug/body', (req, res) => {
    console.log('=== REQUEST DEBUG ===');
    console.log('Method:', req.method);
    console.log('Path:', req.path);
    console.log('Headers:', JSON.stringify(req.headers, null, 2));
    console.log('Body:', req.body);
    console.log('Body type:', typeof req.body);
    console.log('Body JSON:', JSON.stringify(req.body, null, 2));
    console.log('===================');
    
    res.json({
        method: req.method,
        path: req.path,
        body: req.body,
        bodyType: typeof req.body,
        contentType: req.headers['content-type']
    });
});

console.log('Debug endpoint loaded: POST /debug/body');
