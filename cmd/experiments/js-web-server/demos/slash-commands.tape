# Demo: Slash Commands in JavaScript REPL
# Shows built-in REPL commands and help system

Output demos/slash-commands.gif

Set FontSize 14
Set Width 120
Set Height 30
Set TypingSpeed 50ms
Set Theme "Dracula"

Type "go run . repl"
Enter
Sleep 2s

Type "/help"
Enter
Sleep 3s

Type "let counter = 0"
Enter
Sleep 1s

Type "counter++"
Enter
Sleep 1s

Type "counter++"
Enter
Sleep 1s

Type "console.log('Counter is now:', counter)"
Enter
Sleep 2s

Type "/clear"
Enter
Sleep 2s

Type "// Screen cleared! Let's test multiline toggle"
Enter
Sleep 1s

Type "/multiline"
Enter
Sleep 1s

Type "const obj = {"
Enter
Sleep 500ms

Type "  name: 'test',"
Enter
Sleep 500ms

Type "  value: 42"
Enter
Sleep 500ms

Type "}"
Enter
Enter
Sleep 1s

Type "/multiline"
Enter
Sleep 1s

Type "/quit"
Enter
Sleep 1s
