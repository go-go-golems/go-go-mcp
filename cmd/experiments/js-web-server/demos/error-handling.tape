# Demo: Error Handling in JavaScript REPL
# Shows syntax errors, runtime errors, and recovery

Output demos/error-handling.gif

Set FontSize 14
Set Width 120
Set Height 30
Set TypingSpeed 50ms
Set Theme "Dracula"

Type "go run . repl"
Enter
Sleep 2s

Type "// Let's test error handling"
Enter
Sleep 1s

Type "2 + 2"
Enter
Sleep 1s

Type "invalid syntax {"
Enter
Sleep 2s

Type "// Syntax error shown, REPL still works"
Enter
Sleep 1s

Type "unknownVariable"
Enter
Sleep 2s

Type "// Runtime error, but we can continue"
Enter
Sleep 1s

Type "function throwError() { throw new Error('Test error!') }"
Enter
Sleep 1s

Type "throwError()"
Enter
Sleep 2s

Type "// Errors don't crash the REPL"
Enter
Sleep 1s

Type "console.log('Still working!')"
Enter
Sleep 2s

Type "/unknown-command"
Enter
Sleep 2s

Type "// Even unknown commands are handled gracefully"
Enter
Sleep 1s

Type "/quit"
Enter
Sleep 1s
