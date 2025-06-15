# Demo: Multiline Mode in JavaScript REPL
# Shows multiline input with Ctrl+J and function definitions

Output demos/multiline-mode.gif

Set FontSize 14
Set Width 120
Set Height 35
Set TypingSpeed 50ms
Set Theme "Dracula"

Type "go run . repl"
Enter
Sleep 2s

Type "function greet(name) {"
Ctrl+J
Sleep 500ms

Type "  return 'Hello, ' + name + '!'"
Ctrl+J  
Sleep 500ms

Type "}"
Enter
Sleep 1s

Type "greet('World')"
Enter
Sleep 2s

Type "/multiline"
Enter
Sleep 1s

Type "const fibonacci = (n) => {"
Enter
Sleep 500ms

Type "  if (n <= 1) return n"
Enter
Sleep 500ms

Type "  return fibonacci(n-1) + fibonacci(n-2)"
Enter
Sleep 500ms

Type "}"
Enter
Enter
Sleep 1s

Type "fibonacci(7)"
Enter
Sleep 2s

Type "/quit"
Enter
Sleep 1s
