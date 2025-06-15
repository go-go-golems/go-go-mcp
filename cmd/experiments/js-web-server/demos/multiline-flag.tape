# Demo: Starting REPL with --multiline flag
# Shows REPL starting in multiline mode

Output demos/multiline-flag.gif

Set FontSize 14
Set Width 120
Set Height 30
Set TypingSpeed 50ms
Set Theme "Dracula"

Type "go run . repl --multiline"
Enter
Sleep 2s

Type "const numbers = [1, 2, 3, 4, 5]"
Enter
Sleep 500ms

Type "const doubled = numbers.map(n => n * 2)"
Enter
Sleep 500ms

Type "console.log('Original:', numbers)"
Enter
Sleep 500ms

Type "console.log('Doubled:', doubled)"
Enter
Enter
Sleep 2s

Type "// Now let's create a more complex function"
Enter
Sleep 500ms

Type "function processArray(arr, operation) {"
Enter
Sleep 500ms

Type "  return arr.map(operation)"
Enter
Sleep 500ms

Type "}"
Enter
Enter
Sleep 1s

Type "processArray([1, 2, 3], x => x * x)"
Enter
Enter
Sleep 2s

Type "/quit"
Enter
Sleep 1s
