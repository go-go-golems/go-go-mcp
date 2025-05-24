// Test path parameter routing

// Register a handler with path parameters
registerHandler("GET", "/users/:id", (request) => {
    console.log("Request object:", JSON.stringify(request, null, 2));
    
    // Use Go field names (capitalized) for direct access
    console.log("request.Params:", request.Params);
    console.log("request.Path:", request.Path);
    console.log("request.URL:", request.URL);
    console.log("request.Method:", request.Method);
    
    return Response.json({
        message: "User endpoint",
        userId: request.Params.id,  // Use Params (capitalized)
        path: request.Path,         // Use Path (capitalized)
        originalUrl: request.URL    // Use URL (capitalized)
    });
});

// Register another handler with multiple parameters
registerHandler("GET", "/api/:version/users/:userId/posts/:postId", (request) => {
    return Response.json({
        message: "Complex path parameters",
        params: request.Params,        // Use Params (capitalized)
        version: request.Params.version,
        userId: request.Params.userId,
        postId: request.Params.postId
    });
});

// Test the answer endpoint like in the trivia game
registerHandler("GET", "/answer/:answerIndex", (request) => {
    const answerIndex = parseInt(request.Params.answerIndex);  // Use Params (capitalized)
    console.log("Answer index from params:", answerIndex);
    
    return Response.json({
        message: "Answer selected",
        answerIndex: answerIndex,
        isValid: !isNaN(answerIndex) && answerIndex >= 0
    });
});

console.log("Path parameter handlers registered!");