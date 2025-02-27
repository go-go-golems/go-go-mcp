---
Title: Shell Commands in Go Go MCP
Slug: shell-commands
Short: Learn how to create and use shell commands in go-go-mcp.
Topics:
  - shell
  - commands
  - tools
Commands:
  - run-command
  - start
Flags:
  - command
  - args
  - verbose
  - debug
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

This tutorial will guide you through creating and using shell commands in go-go-mcp. Shell commands allow you to define executable commands and scripts in YAML files, with support for templated arguments, environment variables, and flexible execution options.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Command Structure](#command-structure)
3. [Templating System](#templating-system)
4. [Basic Shell Command](#basic-shell-command)
5. [Advanced Features](#advanced-features)
6. [Real-World Examples](#real-world-examples)
7. [Best Practices](#best-practices)
8. [Troubleshooting](#troubleshooting)

## Getting Started

Shell commands in go-go-mcp are defined in YAML files. These files describe how to execute commands, what parameters they accept, and how to handle their output.

### Prerequisites

- go-go-mcp installed
- Basic understanding of YAML
- Basic shell scripting knowledge

### Directory Setup

Create a directory for your shell commands:

```bash
mkdir -p tools/shell-commands
cd tools/shell-commands
```

## Command Structure

A shell command YAML file has the following structure:

```yaml
# Metadata
name: command-name           # Required: Command name (use lowercase and underscores)
short: Short description     # Required: One-line description
long: |                      # Optional: Detailed multi-line description
  Detailed description that can
  span multiple lines

# Parameter Definition
flags:                       # Optional: Command parameters
  - name: flag_name         # Required: Parameter name (use underscores, not hyphens)
    type: string            # Required: Parameter type
    help: Description       # Required: Parameter description
    required: true          # Optional: Whether the parameter is required
    default: value          # Optional: Default value
    choices: [a, b, c]      # Optional: For choice/choiceList types

# Execution Configuration
command:                    # Either command: or shell-script: is required
  - executable             # List of command parts
  - arg1
  - "{{ .Args.flag_name }}"

# OR

shell-script: |             # For complex shell scripts
  #!/bin/bash
  echo "Using {{ .Args.flag_name }}"

# Optional Configuration
environment:                # Optional: Environment variables
  ENV_VAR: "{{ .Args.flag_name }}"
cwd: /path/to/working/dir  # Optional: Working directory
capture-stderr: true       # Optional: Capture stderr in output
```

### Parameter Types

The following parameter types are supported:

- **Basic Types**
  - `string`: Text values
  - `int`: Integer numbers
  - `float`: Floating point numbers
  - `bool`: True/false values
  - `date`: Date values

- **List Types**
  - `stringList`: List of strings
  - `intList`: List of integers
  - `floatList`: List of floating point numbers

- **Choice Types**
  - `choice`: Single selection from predefined options
  - `choiceList`: Multiple selections from predefined options

- **File Types**
  - `file`: Single file input
  - `fileList`: Multiple file inputs
  - `stringFromFile`: String content read from a file
  - `objectFromFile`: Structured data read from a file
  - `stringListFromFile`: List of strings read from a file
  - `objectListFromFile`: List of structured data read from a file

- **Special Types**
  - `keyValue`: Key-value pairs

## Templating System

Shell commands use Go's template language for variable interpolation and control flow. Additionally, all [Sprig template functions](http://masterminds.github.io/sprig/) are available for use.

### Variable Access

Access flag values using the `.Args` object:

```yaml
command:
  - echo
  - "{{ .Args.name }}"     # Access a flag value
```

### Control Flow

Use Go template syntax for control flow:

```yaml
command:
  - rsync
  - "{{ if .Args.verbose }}-v{{ end }}"  # Conditional
  - "{{ range .Args.files }}{{ . }} {{ end }}"  # Iteration
```

### Sprig Functions

Examples of using Sprig functions:

```yaml
shell-script: |
  #!/bin/bash
  
  # String manipulation
  NAME="{{ .Args.name | lower | trim }}"
  
  # Date formatting
  DATE="{{ now | date "2006-01-02" }}"
  
  # List operations
  ITEMS="{{ .Args.items | join "," }}"
  
  # Math operations
  COUNT="{{ .Args.number | add 1 }}"
```

Common Sprig functions:
- String: `trim`, `upper`, `lower`, `title`, `indent`
- Lists: `first`, `last`, `join`, `split`, `compact`
- Math: `add`, `mul`, `div`, `mod`
- Date: `now`, `date`, `dateInZone`
- Encoding: `b64enc`, `b64dec`, `toJson`, `fromJson`

## Basic Shell Command

Let's start with a simple example that copies a file:

```yaml
# tools/shell-commands/copy-file.yaml
name: copy-file
short: Copy a file with progress
flags:
  - name: source
    type: string
    help: Source file
    required: true
  - name: dest
    type: string
    help: Destination path
    required: true
  - name: verbose
    type: bool
    help: Show progress
    default: false
command:
  - rsync
  - -av
  - "{{ if .Args.verbose }}-P{{ end }}"
  - "{{ .Args.source }}"
  - "{{ .Args.dest }}"
```

### Running Your First Command

1. Save the above YAML as `copy-file.yaml`
2. Run it using go-go-mcp:

```bash
go-go-mcp run-command tools/shell-commands/copy-file.yaml \
  --source data.txt \
  --dest backup/ \
  --verbose
```

### Understanding the Structure

Let's break down the key components:

1. **Metadata**
   ```yaml
   name: copy-file
   short: Copy a file with progress
   ```
   These fields identify and describe your command.

2. **Flags**
   ```yaml
   flags:
     - name: source
       type: string
       help: Source file
       required: true
   ```
   Define the parameters your command accepts.

3. **Command Definition**
   ```yaml
   command:
     - rsync
     - -av
   ```
   Specify the actual command to execute.

## Advanced Features

### Using Shell Scripts

For more complex operations, use shell scripts:

```yaml
# tools/shell-commands/process-logs.yaml
name: process-logs
short: Process log files
flags:
  - name: pattern
    type: string
    help: Search pattern
    required: true
  - name: output
    type: string
    help: Output file
    default: output.log
shell-script: |
  #!/bin/bash
  set -euo pipefail
  
  echo "Searching for {{ .Args.pattern }}"
  find . -type f -name "*.log" -exec grep -H "{{ .Args.pattern }}" {} \; > {{ .Args.output }}
```

### Environment Variables

Add environment variables to your commands:

```yaml
# tools/shell-commands/db-backup.yaml
name: db-backup
short: Backup database
flags:
  - name: database
    type: string
    help: Database name
    required: true
  - name: password
    type: string
    help: Database password
    required: true
environment:
  PGPASSWORD: "{{ .Args.password }}"
  PGDATABASE: "{{ .Args.database }}"
command:
  - pg_dump
  - -Fc
  - "{{ .Args.database }}"
```

### Working Directory

Specify a working directory for your command:

```yaml
# tools/shell-commands/build.yaml
name: build
short: Build project
cwd: /path/to/project
command:
  - make
  - build
```

## Real-World Examples

### Docker Management

```yaml
# tools/shell-commands/docker-manage.yaml
name: docker-manage
short: Manage Docker containers
flags:
  - name: action
    type: choice
    help: Action to perform
    choices: [start, stop, restart, logs]
    required: true
  - name: container
    type: string
    help: Container name
    required: true
shell-script: |
  #!/bin/bash
  set -euo pipefail
  
  case {{ .Args.action }} in
    start)
      docker start {{ .Args.container }}
      ;;
    stop)
      docker stop {{ .Args.container }}
      ;;
    restart)
      docker restart {{ .Args.container }}
      ;;
    logs)
      docker logs -f {{ .Args.container }}
      ;;
  esac
```

### Git Operations

```yaml
# tools/shell-commands/git-sync.yaml
name: git-sync
short: Sync git repositories
flags:
  - name: repos_dir
    type: string
    help: Directory containing git repos
    required: true
  - name: branch
    type: string
    help: Branch to sync
    default: main
shell-script: |
  #!/bin/bash
  set -euo pipefail
  
  find {{ .Args.repos_dir }} -type d -name ".git" | while read -r gitdir; do
    repo=$(dirname "$gitdir")
    echo "Syncing $repo..."
    cd "$repo"
    git fetch origin
    git checkout {{ .Args.branch }}
    git pull origin {{ .Args.branch }}
  done
```

## Best Practices

### 1. Error Handling

Always include proper error handling in shell scripts:

```yaml
shell-script: |
  #!/bin/bash
  set -euo pipefail
  
  if ! command -v required-tool &> /dev/null; then
    echo "Error: required-tool is not installed" >&2
    exit 1
  fi
```

### 2. Input Validation

Validate inputs before using them:

```yaml
shell-script: |
  #!/bin/bash
  set -euo pipefail
  
  if [[ ! -d "{{ .Args.directory }}" ]]; then
    echo "Error: Directory does not exist: {{ .Args.directory }}" >&2
    exit 1
  fi
```

### 3. Progress Information

Keep users informed:

```yaml
shell-script: |
  #!/bin/bash
  set -euo pipefail
  
  echo "Starting operation..."
  for item in *; do
    echo "Processing $item..."
    # process item
  done
  echo "Operation complete"
```

### 4. Flag Naming

Use underscores in flag names, not hyphens:

```yaml
flags:
  - name: source_dir    # Good
    type: string
  - name: source-dir    # Bad
    type: string
```

## Troubleshooting

### Common Issues

1. **Template Errors**
   ```
   Error: template: command:1: bad character U+002D '-'
   ```
   Solution: Use underscores (_) instead of hyphens (-) in flag names.

2. **Permission Issues**
   ```
   Error: permission denied: ./script.sh
   ```
   Solution: Ensure your shell script is executable or run with proper permissions.

3. **Missing Dependencies**
   ```
   Error: command not found: required-tool
   ```
   Solution: Install required dependencies or add error checking.

4. **Expansion errors in heredoc strings**
   ```
   Error: $1: unbound variable
   ```
   Solution: Use single quotes for heredoc strings.
   cat <<'EOF' > FILE
   {{ .data }}
   EOF

### Debugging Tips

1. Enable verbose output:
   ```bash
   go-go-mcp run-command command.yaml --verbose
   ```

2. Use debug mode:
   ```bash
   go-go-mcp run-command command.yaml --debug
   ```

3. Check command help:
   ```bash
   go-go-mcp run-command command.yaml --help
   ```

## Integration with Configuration

Shell commands can be integrated into your go-go-mcp configuration:

```yaml
# config.yaml
version: "1"
profiles:
  default:
    tools:
      directories:
        - path: ./tools/shell-commands
          defaults:
            default:
              verbose: true
```

This allows you to:
1. Load all shell commands from a directory
2. Set default parameters
3. Control access through blacklists/whitelists

## Next Steps

1. Create your own shell commands
2. Integrate them with your go-go-mcp configuration
3. Build complex automation workflows
4. Share commands with your team

Remember to check the [Configuration File Tutorial](01-config-file.md) for more information about integrating shell commands into your go-go-mcp setup. 