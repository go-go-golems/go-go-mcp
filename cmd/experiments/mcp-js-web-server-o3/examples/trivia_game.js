// Simple Trivia Game
// A web-based trivia game with multiple choice questions

// Initialize game state
if (!globalState.triviaGame) {
    globalState.triviaGame = {
        currentQuestionIndex: 0,
        score: 0,
        totalQuestions: 0,
        gameStarted: false,
        playerName: '',
        questions: [
            {
                question: "What is the capital of France?",
                options: ["London", "Berlin", "Paris", "Madrid"],
                correct: 2
            },
            {
                question: "Which planet is known as the Red Planet?",
                options: ["Venus", "Mars", "Jupiter", "Saturn"],
                correct: 1
            },
            {
                question: "What is 2 + 2?",
                options: ["3", "4", "5", "6"],
                correct: 1
            },
            {
                question: "Who painted the Mona Lisa?",
                options: ["Van Gogh", "Picasso", "Leonardo da Vinci", "Michelangelo"],
                correct: 2
            },
            {
                question: "What is the largest ocean on Earth?",
                options: ["Atlantic", "Indian", "Arctic", "Pacific"],
                correct: 3
            },
            {
                question: "In which year did World War II end?",
                options: ["1944", "1945", "1946", "1947"],
                correct: 1
            },
            {
                question: "What is the chemical symbol for gold?",
                options: ["Go", "Gd", "Au", "Ag"],
                correct: 2
            },
            {
                question: "Which is the smallest country in the world?",
                options: ["Monaco", "Vatican City", "San Marino", "Liechtenstein"],
                correct: 1
            }
        ]
    };
    globalState.triviaGame.totalQuestions = globalState.triviaGame.questions.length;
}

// Helper function to generate HTML
function generateHTML(title, body) {
    return `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>${title}</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            color: white;
        }
        .container {
            background: rgba(255, 255, 255, 0.1);
            border-radius: 15px;
            padding: 30px;
            backdrop-filter: blur(10px);
            box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
        }
        h1 {
            text-align: center;
            margin-bottom: 30px;
            font-size: 2.5em;
            text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.3);
        }
        .question {
            font-size: 1.3em;
            margin-bottom: 20px;
            font-weight: bold;
        }
        .options {
            display: grid;
            gap: 10px;
            margin-bottom: 20px;
        }
        .option {
            background: rgba(255, 255, 255, 0.2);
            border: 2px solid transparent;
            border-radius: 10px;
            padding: 15px;
            cursor: pointer;
            transition: all 0.3s ease;
            text-decoration: none;
            color: white;
            display: block;
        }
        .option:hover {
            background: rgba(255, 255, 255, 0.3);
            border-color: rgba(255, 255, 255, 0.5);
            transform: translateY(-2px);
        }
        .score {
            text-align: center;
            font-size: 1.2em;
            margin-bottom: 20px;
        }
        .progress {
            background: rgba(255, 255, 255, 0.2);
            border-radius: 10px;
            height: 10px;
            margin-bottom: 20px;
            overflow: hidden;
        }
        .progress-bar {
            background: linear-gradient(90deg, #4CAF50, #45a049);
            height: 100%;
            transition: width 0.3s ease;
        }
        .btn {
            background: linear-gradient(45deg, #4CAF50, #45a049);
            color: white;
            border: none;
            padding: 15px 30px;
            border-radius: 10px;
            cursor: pointer;
            font-size: 1.1em;
            text-decoration: none;
            display: inline-block;
            transition: all 0.3s ease;
            margin: 10px;
        }
        .btn:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 15px rgba(0, 0, 0, 0.2);
        }
        .input-group {
            margin-bottom: 20px;
        }
        .input-group input {
            width: 100%;
            padding: 15px;
            border: none;
            border-radius: 10px;
            font-size: 1.1em;
            background: rgba(255, 255, 255, 0.9);
            color: #333;
        }
        .result {
            text-align: center;
            font-size: 1.5em;
            margin: 20px 0;
        }
        .correct {
            color: #4CAF50;
            font-weight: bold;
        }
        .incorrect {
            color: #f44336;
            font-weight: bold;
        }
    </style>
</head>
<body>
    <div class="container">
        ${body}
    </div>
</body>
</html>`;
}

// Main game page
registerHandler("GET", "/", () => {
    const game = globalState.triviaGame;
    
    if (!game.gameStarted) {
        return generateHTML("🧠 Trivia Challenge", `
            <h1>🧠 Welcome to Trivia Challenge!</h1>
            <div style="text-align: center;">
                <p style="font-size: 1.2em; margin-bottom: 30px;">
                    Test your knowledge with ${game.totalQuestions} exciting questions!
                </p>
                <form action="/start" method="POST">
                    <div class="input-group">
                        <input type="text" name="playerName" placeholder="Enter your name" required>
                    </div>
                    <button type="submit" class="btn">🚀 Start Game</button>
                </form>
            </div>
        `);
    }
    
    if (game.currentQuestionIndex >= game.totalQuestions) {
        const percentage = Math.round((game.score / game.totalQuestions) * 100);
        let message = "";
        if (percentage >= 80) message = "🏆 Excellent work!";
        else if (percentage >= 60) message = "👍 Good job!";
        else if (percentage >= 40) message = "📚 Keep studying!";
        else message = "💪 Better luck next time!";
        
        return generateHTML("Game Complete", `
            <h1>🎉 Game Complete!</h1>
            <div class="result">
                <p>Great job, <strong>${game.playerName}</strong>!</p>
                <p>Final Score: <span class="correct">${game.score}/${game.totalQuestions}</span></p>
                <p>Percentage: <strong>${percentage}%</strong></p>
                <p>${message}</p>
            </div>
            <div style="text-align: center;">
                <a href="/restart" class="btn">🔄 Play Again</a>
                <a href="/leaderboard" class="btn">🏅 View Stats</a>
            </div>
        `);
    }
    
    const currentQuestion = game.questions[game.currentQuestionIndex];
    const progress = ((game.currentQuestionIndex) / game.totalQuestions) * 100;
    
    return generateHTML("Trivia Question", `
        <h1>🧠 Trivia Challenge</h1>
        <div class="score">
            Player: <strong>${game.playerName}</strong> | 
            Score: <strong>${game.score}/${game.currentQuestionIndex}</strong> | 
            Question ${game.currentQuestionIndex + 1}/${game.totalQuestions}
        </div>
        <div class="progress">
            <div class="progress-bar" style="width: ${progress}%"></div>
        </div>
        <div class="question">
            ${currentQuestion.question}
        </div>
        <div class="options">
            ${currentQuestion.options.map((option, index) => 
                `<a href="/answer/${index}" class="option">
                    ${String.fromCharCode(65 + index)}. ${option}
                </a>`
            ).join('')}
        </div>
    `);
});

// Start game
registerHandler("POST", "/start", (request) => {
    const game = globalState.triviaGame;
    
    // Parse form data (simple parsing for name field)
    const body = request.body || "";
    const nameMatch = body.match(/playerName=([^&]+)/);
    const playerName = nameMatch ? decodeURIComponent(nameMatch[1].replace(/\+/g, ' ')) : "Anonymous";
    
    game.playerName = playerName;
    game.gameStarted = true;
    game.currentQuestionIndex = 0;
    game.score = 0;
    
    // Return HTML with immediate redirect
    return `
<!DOCTYPE html>
<html>
<head>
    <meta http-equiv="refresh" content="0; url=/">
    <title>Starting Game...</title>
</head>
<body>
    <p>Starting game... If you are not redirected, <a href="/">click here</a>.</p>
    <script>window.location.href = '/';</script>
</body>
</html>`;
});

// Answer handler
registerHandler("GET", "/answer/:answerIndex", (request) => {
    const game = globalState.triviaGame;
    
    // Safely extract answer index from URL
    let answerIndex = 0;
    try {
        const urlPath = request.url || request.path || '';
        const pathParts = urlPath.split('/');
        answerIndex = parseInt(pathParts[2]) || 0;
    } catch (e) {
        console.log("Error parsing answer index:", e);
        answerIndex = 0;
    }
    
    if (game.currentQuestionIndex >= game.totalQuestions) {
        // Return HTML with immediate redirect
        return `
<!DOCTYPE html>
<html>
<head>
    <meta http-equiv="refresh" content="0; url=/">
    <title>Game Complete...</title>
</head>
<body>
    <p>Game complete... If you are not redirected, <a href="/">click here</a>.</p>
    <script>window.location.href = '/';</script>
</body>
</html>`;
    }
    
    const currentQuestion = game.questions[game.currentQuestionIndex];
    const isCorrect = answerIndex === currentQuestion.correct;
    
    if (isCorrect) {
        game.score++;
    }
    
    const resultHTML = generateHTML("Answer Result", `
        <h1>🧠 Trivia Challenge</h1>
        <div class="score">
            Player: <strong>${game.playerName}</strong> | 
            Score: <strong>${game.score}/${game.currentQuestionIndex + 1}</strong>
        </div>
        <div class="question">
            ${currentQuestion.question}
        </div>
        <div class="result ${isCorrect ? 'correct' : 'incorrect'}">
            ${isCorrect ? '✅ Correct!' : '❌ Incorrect!'}
        </div>
        <div style="text-align: center; margin: 20px 0;">
            <p><strong>The correct answer was:</strong></p>
            <p style="font-size: 1.2em; color: #4CAF50;">
                ${String.fromCharCode(65 + currentQuestion.correct)}. ${currentQuestion.options[currentQuestion.correct]}
            </p>
        </div>
        <div style="text-align: center;">
            <a href="/next" class="btn">
                ${game.currentQuestionIndex + 1 >= game.totalQuestions ? '🏁 View Results' : '➡️ Next Question'}
            </a>
        </div>
    `);
    
    return resultHTML;
});

// Next question
registerHandler("GET", "/next", () => {
    const game = globalState.triviaGame;
    game.currentQuestionIndex++;
    
    // Return HTML with immediate redirect
    return `
<!DOCTYPE html>
<html>
<head>
    <meta http-equiv="refresh" content="0; url=/">
    <title>Next Question...</title>
</head>
<body>
    <p>Loading next question... If you are not redirected, <a href="/">click here</a>.</p>
    <script>window.location.href = '/';</script>
</body>
</html>`;
});

// Restart game
registerHandler("GET", "/restart", () => {
    const game = globalState.triviaGame;
    game.gameStarted = false;
    game.currentQuestionIndex = 0;
    game.score = 0;
    game.playerName = '';
    
    // Return HTML with immediate redirect
    return `
<!DOCTYPE html>
<html>
<head>
    <meta http-equiv="refresh" content="0; url=/">
    <title>Restarting Game...</title>
</head>
<body>
    <p>Restarting game... If you are not redirected, <a href="/">click here</a>.</p>
    <script>window.location.href = '/';</script>
</body>
</html>`;
});

// Simple leaderboard/stats
registerHandler("GET", "/leaderboard", () => {
    const game = globalState.triviaGame;
    
    return generateHTML("Game Stats", `
        <h1>📊 Game Statistics</h1>
        <div style="text-align: center;">
            <p style="font-size: 1.3em;">Last Game Results</p>
            <div class="result">
                <p>Player: <strong>${game.playerName || 'No games played yet'}</strong></p>
                <p>Score: <strong>${game.score}/${game.totalQuestions}</strong></p>
                <p>Percentage: <strong>${Math.round((game.score / game.totalQuestions) * 100)}%</strong></p>
            </div>
            <div style="margin-top: 30px;">
                <a href="/" class="btn">🏠 Home</a>
                <a href="/restart" class="btn">🎮 New Game</a>
            </div>
        </div>
    `);
});

// API endpoint to get game status
registerHandler("GET", "/api/status", () => {
    return {
        gameStarted: globalState.triviaGame.gameStarted,
        currentQuestion: globalState.triviaGame.currentQuestionIndex + 1,
        totalQuestions: globalState.triviaGame.totalQuestions,
        score: globalState.triviaGame.score,
        playerName: globalState.triviaGame.playerName
    };
});

console.log("🎮 Trivia Game initialized!");
console.log("📍 Visit http://localhost:8080 to start playing!");
console.log("🎯 Features:");
console.log("   - Multiple choice questions");
console.log("   - Score tracking");
console.log("   - Progress indicator");
console.log("   - Beautiful responsive UI");
console.log("   - Game statistics"); 