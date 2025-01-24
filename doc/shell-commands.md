# Shell Commands Documentation

Shell commands allow you to define executable commands and scripts in YAML files, with support for templated arguments, environment variables, and flexible execution options.

## Table of Contents
1. [Basic Structure](#basic-structure)
2. [Command Types](#command-types)
   - [Command Lists](#command-lists)
   - [Shell Scripts](#shell-scripts)
3. [Parameters](#parameters)
   - [Flags](#flags)
   - [Arguments](#arguments)
4. [Environment Variables](#environment-variables)
5. [Working Directory](#working-directory)
6. [Output Handling](#output-handling)
7. [Examples](#examples)
   - [Simple Command](#simple-command)
   - [Complex Script](#complex-script)
   - [Docker Operations](#docker-operations)
   - [Git Operations](#git-operations)
   - [Database Operations](#database-operations)

## Basic Structure

A shell command YAML file has the following basic structure:

```yaml
name: command-name
short: Short description
long: |
  Detailed description that can
  span multiple lines
flags:
  - name: flag_name
    type: string
    help: Flag description
    required: true
command:
  - executable
  - arg1
  - "{{ .Args.flag_name }}"
# OR
shell-script: |
  #!/bin/bash
  echo "Using {{ .Args.flag_name }}"
environment:
  ENV_VAR: "{{ .Args.some_flag }}"
cwd: /path/to/working/dir
capture-stderr: true
```

## Command Types

### Command Lists

Command lists are useful for simple commands with fixed arguments:

```yaml
name: aws-s3-upload
short: Upload a file to S3
flags:
  - name: file
    type: string
    help: File to upload
    required: true
  - name: bucket
    type: string
    help: Target bucket
    required: true
command:
  - aws
  - s3
  - cp
  - "{{ .Args.file }}"
  - "s3://{{ .Args.bucket }}/"
```

### Shell Scripts

Shell scripts are better for complex operations requiring logic:

```yaml
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

## Parameters

### Flags

Flags are named parameters with types and options:

```yaml
flags:
  - name: verbose
    type: bool
    help: Enable verbose output
    default: false
  
  - name: count
    type: int
    help: Number of iterations
    required: true
    
  - name: mode
    type: choice
    help: Operation mode
    choices: [fast, safe, debug]
    default: safe
```

Supported flag types:
- `string`: Text values (e.g., `--name=value`)
- `int`: Integer numbers (e.g., `--count=10`)
- `float`: Floating point numbers (e.g., `--threshold=0.75`)
- `bool`: True/false values (e.g., `--verbose` or `--verbose=false`)
- `date`: Date values (e.g., `--from=2024-01-01`)
- `stringList`: List of strings (e.g., `--tags=a,b,c` or multiple `--tags=a --tags=b`)
- `intList`: List of integers (e.g., `--numbers=1,2,3`)
- `floatList`: List of floating point numbers
- `choice`: Single selection from predefined options
- `choiceList`: Multiple selections from predefined options
- `keyValue`: Key-value pairs (e.g., `--header='Content-Type:application/json'`)
- `file`: Single file input
- `fileList`: Multiple file inputs
- `stringFromFile`: String content read from a file
- `objectFromFile`: Structured data read from a file
- `stringListFromFile`: List of strings read from a file
- `objectListFromFile`: List of structured data read from a file

Each flag definition supports these fields:
- `name`: (required) The parameter name used in CLI
- `type`: (required) One of the types listed above
- `help`: Short description of the parameter
- `default`: Default value if not provided
- `required`: Set to true if the parameter must be provided
- `choices`: List of valid options (for `choice` and `choiceList` types)

Examples of different flag types:

```yaml
flags:
  # Date range with defaults
  - name: from
    type: date
    help: Start date (inclusive)
    default: 2024-01-01
  
  # Choice from predefined options
  - name: group_by
    type: choice
    help: Result grouping
    choices: [year, month, all-time]
    default: month
  
  # List of allowed statuses
  - name: status
    type: stringList
    help: Order statuses to include
    default: ['pending', 'processing']
  
  # File input
  - name: config
    type: file
    help: Configuration file
    required: true
  
  # Key-value pairs
  - name: labels
    type: keyValue
    help: Resource labels
    default:
      env: dev
      team: backend
```

### Arguments

Arguments are positional parameters:

```yaml
arguments:
  - name: source
    type: string
    help: Source file
    required: true
  
  - name: destination
    type: string
    help: Destination path
```

## Environment Variables

Environment variables can be templated using flag values:

```yaml
flags:
  - name: environment
    type: string
    help: Deployment environment
    choices: [dev, staging, prod]
    default: dev

environment:
  NODE_ENV: "{{ .Args.environment }}"
  DB_HOST: "db.{{ .Args.environment }}.internal"
  LOG_LEVEL: "{{ if eq .Args.environment \"prod\" }}error{{ else }}debug{{ end }}"
```

## Working Directory

Set the working directory for command execution:

```yaml
name: build
short: Build project
cwd: /path/to/project
command:
  - make
  - build
```

## Output Handling

Control stderr capture:

```yaml
name: risky-operation
short: Run risky operation
capture-stderr: true  # Capture stderr in command output
shell-script: |
  #!/bin/bash
  set -e
  
  if ! some_command; then
    echo "Failed!" >&2
    exit 1
  fi
```

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

Usage:
```bash
mcp-client run-command copy-file.yaml --source data.txt --dest backup/ --verbose
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
    type: string
    help: Output format
    choices: [text, json, csv]
    default: text
shell-script: |
  #!/bin/bash
  set -euo pipefail
  
  # Convert time period to timestamp
  since_time=$(date -d "{{ .Args.since }} ago" +%s)
  
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

## Running Commands

Commands can be run using the `run-command` subcommand:

```bash
mcp-server run-command path/to/command.yaml [flags]
```

Example:
```bash
mcp-server run-command examples/db-backup.yaml \
  --database myapp \
  --bucket backups.example.com \
  --keep-local
```

## Debugging Tips

1. Use `--help` to see available flags:
   ```bash
   mcp-server run-command example.yaml --help
   ```

2. Enable verbose output when available:
   ```bash
   mcp-server run-command example.yaml --verbose
   ```

3. Check script output:
   ```bash
   mcp-server run-command example.yaml --debug
   ```

## Common Issues

1. **Template Errors**
   - Check flag names match template variables
   - Verify template syntax
   - Ensure required flags are provided

2. **Permission Issues**
   - Check file permissions
   - Verify working directory access
   - Ensure sufficient privileges

3. **Environment Issues**
   - Verify environment variables
   - Check path configurations
   - Validate external dependencies 