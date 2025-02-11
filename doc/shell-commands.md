# Shell Commands Documentation

Shell commands allow you to define executable commands and scripts in YAML files, with support for templated arguments, environment variables, and flexible execution options.

## Table of Contents
1. [Command Structure](#command-structure)
2. [Parameter Types](#parameter-types)
3. [Templating System](#templating-system)
4. [Examples](#examples)
5. [Best Practices](#best-practices)
6. [Troubleshooting](#troubleshooting)

## Command Structure

A shell command YAML file has the following structure:

```yaml
# Metadata (Required)
name: command-name           # Command name (use lowercase and underscores)
short: Short description     # One-line description
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

## Parameter Types

The following parameter types are supported:

### Basic Types
- `string`: Text values (e.g., `--name=value`)
- `int`: Integer numbers (e.g., `--count=10`)
- `float`: Floating point numbers (e.g., `--threshold=0.75`)
- `bool`: True/false values (e.g., `--verbose` or `--verbose=false`)
- `date`: Date values (e.g., `--from=2024-01-01`)

### List Types
- `stringList`: List of strings (e.g., `--tags=a,b,c` or multiple `--tags=a --tags=b`)
- `intList`: List of integers (e.g., `--numbers=1,2,3`)
- `floatList`: List of floating point numbers

### Choice Types
- `choice`: Single selection from predefined options
- `choiceList`: Multiple selections from predefined options

### File Types
- `file`: Single file input
- `fileList`: Multiple file inputs
- `stringFromFile`: String content read from a file
- `objectFromFile`: Structured data read from a file
- `stringListFromFile`: List of strings read from a file
- `objectListFromFile`: List of structured data read from a file

### Special Types
- `keyValue`: Key-value pairs (e.g., `--header='Content-Type:application/json'`)

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

For a complete list of available functions, visit the [Sprig documentation](http://masterminds.github.io/sprig/).

## Examples

### Simple Command

A simple file copy command:

```yaml
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

### Complex Script

A log processing script with multiple options:

```yaml
name: analyze-logs
short: Analyze application logs
flags:
  - name: log_dir
    type: string
    help: Log directory
    required: true
  - name: pattern
    type: string
    help: Search pattern
    required: true
  - name: since
    type: string
    help: Process logs since timestamp
    default: "24h"
  - name: format
    type: choice
    help: Output format
    choices: [text, json, csv]
    default: text
shell-script: |
  #!/bin/bash
  set -euo pipefail
  
  # Convert time period to timestamp using Sprig's date functions
  since_time=$(date -d "{{ .Args.since | duration "s" }} ago" +%s)
  
  # Process each log file
  find {{ .Args.log_dir }} -type f -name "*.log" | while read -r log; do
    if [[ $(stat -c %Y "$log") -gt $since_time ]]; then
      case {{ .Args.format }} in
        json)
          grep "{{ .Args.pattern }}" "$log" | jq -R -s 'split("\n")[:-1] | map({line:.})'
          ;;
        csv)
          grep "{{ .Args.pattern }}" "$log" | sed 's/^/$(basename "$log"),/' >> output.csv
          ;;
        text|*)
          grep "{{ .Args.pattern }}" "$log"
          ;;
      esac
    fi
  done
```

### Docker Operations

Managing Docker containers:

```yaml
name: docker-ops
short: Manage Docker containers
flags:
  - name: action
    type: string
    help: Action to perform
    choices: [start, stop, restart, logs]
    required: true
  - name: container
    type: string
    help: Container name
    required: true
  - name: tail
    type: int
    help: Number of log lines
    default: 100
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
      docker logs --tail {{ .Args.tail }} -f {{ .Args.container }}
      ;;
  esac
```

### Git Operations

Managing multiple git repositories:

```yaml
name: git-sync
short: Synchronize git repositories
flags:
  - name: repos-dir
    type: string
    help: Directory containing git repos
    required: true
  - name: branch
    type: string
    help: Branch to sync
    default: main
  - name: push
    type: bool
    help: Push changes
    default: false
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
    
    if [[ "{{ .Args.push }}" == "true" ]]; then
      git push origin {{ .Args.branch }}
    fi
  done
environment:
  GIT_TERMINAL_PROMPT: "0"
```

### Database Operations

Database backup script:

```yaml
name: db-backup
short: Backup database to S3
flags:
  - name: database
    type: string
    help: Database name
    required: true
  - name: bucket
    type: string
    help: S3 bucket name
    required: true
  - name: keep-local
    type: bool
    help: Keep local backup
    default: false
environment:
  PGPASSWORD: "{{ .Args.password }}"
  AWS_PROFILE: "{{ .Args.profile }}"
shell-script: |
  #!/bin/bash
  set -euo pipefail
  
  # Create backup
  backup_file="/tmp/{{ .Args.database }}_$(date +%Y%m%d_%H%M%S).sql"
  pg_dump {{ .Args.database }} > "$backup_file"
  
  # Upload to S3
  aws s3 cp "$backup_file" "s3://{{ .Args.bucket }}/backups/"
  
  # Cleanup
  if [[ "{{ .Args.keep_local }}" != "true" ]]; then
    rm "$backup_file"
  fi
```

## Best Practices

1. **Error Handling**
   - Use `set -e` in shell scripts to fail fast
   - Add proper error messages
   - Consider cleanup on failure

2. **Security**
   - Don't hardcode sensitive values
   - Use environment variables for secrets
   - Validate user input

3. **Maintainability**
   - Add clear descriptions
   - Document all parameters
   - Use meaningful variable names

4. **Flexibility**
   - Make paths configurable
   - Provide sensible defaults
   - Allow overriding key behaviors

5. **Output**
   - Use `capture-stderr` appropriately
   - Provide progress information
   - Format output for readability

## Troubleshooting

1. **Template Errors**
   - Check flag names match template variables
   - Verify template syntax
   - Ensure required flags are provided
   - Make sure that flag and argument names use _ and not -

2. **Permission Issues**
   - Check file permissions
   - Verify working directory access
   - Ensure sufficient privileges

3. **Environment Issues**
   - Verify environment variables
   - Check path configurations
   - Validate external dependencies 