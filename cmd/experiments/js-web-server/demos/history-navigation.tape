# Demo: History Navigation with Arrow Keys
# Shows arrow key navigation through command history

Output demos/history-navigation.gif

Set FontSize 14
Set Width 120
Set Height 30
Set TypingSpeed 50ms
Set Theme "Dracula"

Type "go run . repl"
Enter
Sleep 2s

Type "2 + 2"
Enter
Sleep 1s

Type "Math.PI * 2"
Enter
Sleep 1s

Type "console.log('Hello World')"
Enter
Sleep 1s

Type "let x = 42"
Enter
Sleep 1s

Type "// Now let's navigate history with arrow keys"
Enter
Sleep 2s

# Navigate up through history
Up
Sleep 1s
Up
Sleep 1s
Up
Sleep 1s
Up
Sleep 1s

# Execute the historical command
Enter
Sleep 2s

# Navigate down through history
Down
Sleep 1s
Down
Sleep 1s

# Execute another historical command
Enter
Sleep 2s

Type "/quit"
Enter
Sleep 1s
