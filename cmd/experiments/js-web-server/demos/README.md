# JavaScript REPL Demos

This directory contains VHS demo tapes showcasing the JavaScript REPL functionality.

## Available Demos

### 1. Basic REPL Usage (`basic-repl.tape`)
- Simple expressions and calculations
- Console.log usage
- Variable assignment
- JSON manipulation

**Generate:** `vhs < basic-repl.tape`

### 2. Multiline Mode (`multiline-mode.tape`)
- Function definitions using Ctrl+J
- Toggling multiline mode with `/multiline`
- Complex code blocks

**Generate:** `vhs < multiline-mode.tape`

### 3. Slash Commands (`slash-commands.tape`)
- Help system (`/help`)
- Screen clearing (`/clear`)
- Multiline toggle (`/multiline`)
- Exiting (`/quit`)

**Generate:** `vhs < slash-commands.tape`

### 4. Error Handling (`error-handling.tape`)
- Syntax error recovery
- Runtime error handling
- Unknown command handling
- REPL resilience

**Generate:** `vhs < error-handling.tape`

### 5. Multiline Flag (`multiline-flag.tape`)
- Starting REPL with `--multiline` flag
- Working in multiline mode by default
- Complex function definitions

**Generate:** `vhs < multiline-flag.tape`

## Requirements

- [VHS](https://github.com/charmbracelet/vhs) installed
- Go runtime for building the js-web-server
- Terminal with good color support

## Generating All Demos

```bash
# From the demos directory
for tape in *.tape; do
    echo "Generating $tape..."
    vhs < "$tape"
done
```

## REPL Features Demonstrated

- **Interactive JavaScript execution** with immediate feedback
- **Multiline input support** using Ctrl+J or starting with `--multiline`
- **Command history** showing previous inputs and outputs
- **Built-in commands** for REPL control
- **Error handling** that doesn't crash the session
- **Console.log support** for debugging
- **Standard JavaScript features** including functions, objects, arrays

## Usage Examples

```bash
# Start basic REPL
go run . repl

# Start in multiline mode
go run . repl --multiline

# Show help
go run . repl --help
```

## REPL Commands

- `/help` - Show available commands
- `/clear` - Clear the screen/history
- `/multiline` - Toggle multiline mode
- `/quit` or `/exit` - Exit the REPL
- `Ctrl+C` - Exit the REPL
- `Ctrl+J` - Add new line in multiline input
