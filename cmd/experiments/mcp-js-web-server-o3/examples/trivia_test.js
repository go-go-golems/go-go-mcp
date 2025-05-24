// Test the trivia game answer endpoint
registerHandler("GET", "/answer/:answerIndex", (request) => {
    console.log("Answer endpoint called with params:", request.params);
    
    const answerIndex = parseInt(request.params.answerIndex);
    console.log("Parsed answer index:", answerIndex);
    
    if (isNaN(answerIndex) || answerIndex < 0) {
        return Response.error("Invalid answer index", 400);
    }
    
    // Simulate game logic
    const correctAnswer = 2; // Example correct answer
    const isCorrect = answerIndex === correctAnswer;
    
    return Response.json({
        selectedAnswer: answerIndex,
        correct: isCorrect,
        message: isCorrect ? "Correct!" : "Wrong answer, try again!"
    });
});

console.log("Trivia game answer endpoint registered!");