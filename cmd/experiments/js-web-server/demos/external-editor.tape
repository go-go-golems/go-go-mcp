# Demo: External Editor Integration
# Shows Ctrl+E to open external editor for complex code

Output demos/external-editor.gif

Set FontSize 14
Set Width 120
Set Height 35
Set TypingSpeed 50ms
Set Theme "Dracula"

Type "go run . repl"
Enter
Sleep 2s

Type "// Let's create a complex function"
Enter
Sleep 1s

Type "function fibonacci(n) {"
Ctrl+J
Sleep 500ms

Type "  if (n <= 1) return n"
Ctrl+J
Sleep 500ms

Type "  return fibonacci(n-1) + fibonacci(n-2)"
Ctrl+J
Sleep 500ms

Type "}"
Sleep 2s

Type "// Now press Ctrl+E to edit this in external editor"
Enter
Sleep 2s

Type "// (In real usage, Ctrl+E would open your $EDITOR)"
Enter
Sleep 1s

Type "// Let's test the /edit command instead"
Enter
Sleep 1s

Type "/edit"
Enter
Sleep 2s

Type "// Editor would allow complex multi-line editing"
Enter
Sleep 1s

Type "fibonacci(7)"
Enter
Enter
Sleep 2s

Type "/quit"
Enter
Sleep 1s
