---
Title: Running MCP in Practice
Slug: mcp-in-practice
Short: A practical guide to running MCP with shell commands and configuration.
Topics:
  - mcp
  - shell
  - config
  - tutorial
Commands:
  - start
  - client
  - run-command
Flags:
  - config-file
  - profile
  - transport
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

This tutorial will walk you through setting up and running MCP in practice, showcasing how MCP tools can be used for various real-world scenarios. By the end of this tutorial, you'll understand how to:
- Create and organize MCP tools for different use cases
- Configure MCP for various scenarios
- Run and test your MCP setup

Each section includes practical exercises to help reinforce the concepts.

## Table of Contents

1. [Setting Up the Project Structure](#setting-up-the-project-structure)
2. [Understanding and Creating Shell Commands](#creating-shell-commands)
3. [Working with Configuration Files](#configuring-mcp)
4. [Running the Server](#running-the-server)
5. [Testing with the Client](#testing-with-the-client)

## Using MCP with Claude Desktop 🤖

Claude desktop can be configured to use MCP servers through its configuration file. This allows you to use MCP tools directly through Claude's edit verbs.

### Configuration

First, configure Claude desktop to use your MCP server:

```bash
# Initialize Claude desktop configuration
go-go-mcp claude-config init

# Add an MCP server configuration
go-go-mcp claude-config add-mcp-server dev \
  --command go-go-mcp \
  --args start --profile development --log-level debug
```

### Using Edit Verbs

Once configured, you can use MCP tools through Claude's edit verbs:

1. **@tool**: Execute an MCP tool
   ```
   @tool list-github-issues --state open --assignee me
   ```

2. **@run**: Run a shell command through MCP
   ```
   @run git status
   ```

3. **@fetch**: Fetch and process web content
   ```
   @fetch https://example.com
   ```

The output from these commands will be automatically processed and included in Claude's context, allowing for natural interaction with the results.

### Best Practices

1. Use descriptive tool names that reflect their purpose
2. Keep sensitive data in environment variables
3. Use debug logging during development
4. Create specific profiles for different use cases

## Setting Up the Project Structure

MCP works best with a well-organized project structure that groups tools by their purpose and functionality. We'll create a structure that follows these principles:

- Group tools by their use case (system monitoring, data analysis, etc.)
- Keep configuration files at the root level
- Organize prompts in their own directory

Let's create this structure:

```bash
mkdir -p mcp-demo/{tools,prompts}/{system,data,calendar,web}
cd mcp-demo
```

This creates:
```
mcp-demo/
├── tools/
   ├── system/    # System monitoring and debugging tools
   ├── data/      # Data analysis and processing tools
   ├── calendar/  # Calendar and scheduling tools
   └── web/       # Web scraping and API tools
```

**Exercise 1: Project Setup**
1. Create the directory structure above
2. Create a README.md file explaining the purpose of each directory
3. Add a `.gitignore` file to exclude sensitive configuration files

## Understanding and Creating Shell Commands

### Shell Command Anatomy

MCP shell commands are defined in YAML files and consist of several key components:

1. **Metadata Section**
   ```yaml
   name: command-name      # Unique identifier for the command
   short: Description      # Brief description shown in listings
   long: |                 # Optional detailed description
     Multiline description
     with more details
   ```

2. **Parameter Definition**
   ```yaml
   flags:
     - name: parameter_name    # Use underscores, not hyphens
       type: string           # Parameter type (string, int, bool, etc.)
       help: Description      # Help text shown in --help
       required: true/false   # Whether the parameter is mandatory
       default: value         # Default value if not specified
       choices: [a, b, c]     # For choice/choiceList types
   ```

3. **Execution Configuration**
   ```yaml
   # Either use command: for simple commands
   command:
     - executable
     - arg1
     - "{{ .Args.parameter }}"
   
   # Or shell-script: for complex scripts
   shell-script: |
     #!/bin/bash
     # Your script here
   ```

4. **Environment and Working Directory**
   ```yaml
   environment:
     ENV_VAR: "{{ .Args.parameter }}"
   cwd: /working/directory
   ```

This allows us to expose arbitrary shell commands as MCP tools and provide a schema and prompt.

### Template System

MCP uses Go's template system with the following key features:

1. **Variable Access**
   - `.Args.parameter` - Access parameter values
   - `.Env.VARIABLE` - Access environment variables

2. **Control Flow**
   ```yaml
   shell-script: |
     # If condition
     {{ if .Args.verbose }}
       echo "Verbose mode"
     {{ end }}
     
     # Loops
     {{ range .Args.items }}
       process "{{ . }}"
     {{ end }}
   ```

3. **Built-in Functions**
   - `trim`, `upper`, `lower` - String manipulation
   - `add`, `sub`, `mul` - Mathematical operations
   - `now`, `date` - Date/time handling

Let's create some practical examples:

### System Monitoring Tool

This tool demonstrates system monitoring and resource tracking:

```yaml
name: system-monitor
short: Monitor system resources and performance
flags:
  - name: format
    type: choice
    help: Output format
    choices: [text, json]
    default: text
  - name: watch
    type: bool
    help: Continuously watch stats
    default: false
  - name: metrics
    type: stringList
    help: Metrics to monitor
    choices: [cpu, memory, disk, network]
    default: [cpu, memory]
shell-script: |
  #!/bin/bash
  set -euo pipefail
  
  METRICS="{{ join .Args.metrics "," }}"
  
  if [[ "{{ .Args.format }}" == "json" ]]; then
    if [[ "{{ .Args.watch }}" == "true" ]]; then
      top -b -n 1 -w 512 -1 | awk 'NR>7 {printf "{\"pid\":%s,\"cpu\":%s,\"mem\":%s,\"command\":\"%s\"}\\n", $1, $9, $10, $12}'
    else
      top -b -n 1 | awk 'NR>7 {printf "{\"pid\":%s,\"cpu\":%s,\"mem\":%s,\"command\":\"%s\"}\\n", $1, $9, $10, $12}'
    fi
  else
    if [[ "{{ .Args.watch }}" == "true" ]]; then
      top
    else
      top -b -n 1
    fi
  fi
```

Key points about this command:
- Monitors system resources in real-time
- Supports multiple output formats
- Allows selecting specific metrics
- Shows proper error handling

**Exercise 2: Create a System Tool**
1. Create a command called `disk-usage.yaml` that:
   - Shows disk usage for specified directories
   - Supports different size units (MB, GB, etc.)
   - Can filter by file types
2. Test it with different parameters
3. Add proper error handling

### Calendar Integration Tool

This example shows calendar management features:

```yaml
name: calendar-event
short: Manage calendar events and meetings
flags:
  - name: title
    type: string
    help: Event title
    required: true
  - name: start_time
    type: string
    help: Event start time (YYYY-MM-DD HH:MM)
    required: true
  - name: duration
    type: int
    help: Duration in minutes
    default: 60
  - name: attendees
    type: stringList
    help: List of attendee email addresses
    default: []
  - name: calendar_id
    type: string
    help: Google Calendar ID
    default: "primary"
shell-script: |
  #!/bin/bash
  set -euo pipefail
  
  # Format the event data
  EVENT_DATA=$(cat << EOF
  {
    "summary": "{{ .Args.title }}",
    "start": {
      "dateTime": "{{ .Args.start_time }}",
      "timeZone": "{{ .Env.TIMEZONE }}"
    },
    "end": {
      "dateTime": "$(date -d "{{ .Args.start_time }} + {{ .Args.duration }} minutes" "+%Y-%m-%dT%H:%M:%S")",
      "timeZone": "{{ .Env.TIMEZONE }}"
    },
    "attendees": [
      {{ range .Args.attendees }}
      {"email": "{{ . }}"},
      {{ end }}
    ]
  }
  EOF
  )
  
  # Create the event using Google Calendar API
  curl -X POST \
    -H "Authorization: Bearer $GOOGLE_TOKEN" \
    -H "Content-Type: application/json" \
    "https://www.googleapis.com/calendar/v3/calendars/{{ .Args.calendar_id }}/events" \
    -d "$EVENT_DATA"
```

Key points:
- Integrates with Google Calendar API
- Handles date/time calculations
- Supports multiple attendees
- Uses environment variables for configuration

**Exercise 3: Enhance the Calendar Tool**
1. Add recurring event support
2. Add meeting location/video link options
3. Add calendar availability checking
4. Add notification settings

### Data Analysis Tool

This example demonstrates data processing capabilities:

```yaml
name: data-analyzer
short: Analyze data files and generate reports
flags:
  - name: input_file
    type: string
    help: Path to input data file (CSV, JSON, etc.)
    required: true
  - name: format
    type: choice
    help: Input file format
    choices: [csv, json, excel]
    required: true
  - name: operations
    type: stringList
    help: Analysis operations to perform
    choices: [summary, correlation, trends, outliers]
    default: [summary]
  - name: output_format
    type: choice
    help: Output format for results
    choices: [text, json, html]
    default: text
shell-script: |
  #!/bin/bash
  set -euo pipefail
  
  # Process the file based on format
  case "{{ .Args.format }}" in
    csv)
      DATA=$(python3 -c "
import pandas as pd
import json

df = pd.read_csv('{{ .Args.input_file }}')
results = {}

{% for op in .Args.operations %}
if '{{ op }}' == 'summary':
    results['summary'] = df.describe().to_dict()
elif '{{ op }}' == 'correlation':
    results['correlation'] = df.corr().to_dict()
elif '{{ op }}' == 'trends':
    results['trends'] = df.rolling(window=3).mean().to_dict()
elif '{{ op }}' == 'outliers':
    results['outliers'] = df[abs(df - df.mean()) > (3 * df.std())].to_dict()
{% endfor %}

print(json.dumps(results))
      ")
      ;;
    json)
      # Similar processing for JSON
      ;;
    excel)
      # Similar processing for Excel
      ;;
  esac
  
  # Format output
  case "{{ .Args.output_format }}" in
    json)
      echo "$DATA"
      ;;
    html)
      # Convert to HTML report
      echo "$DATA" | python3 -c "
import json
import pandas as pd

data = json.loads(input())
html = '<html><body>'
for op, result in data.items():
    html += f'<h2>{op.title()} Analysis</h2>'
    html += pd.DataFrame(result).to_html()
html += '</body></html>'
print(html)
      "
      ;;
    text)
      # Format as readable text
      echo "$DATA" | python3 -c "
import json
data = json.loads(input())
for op, result in data.items():
    print(f'\n=== {op.title()} Analysis ===')
    for key, value in result.items():
        print(f'{key}: {value}')
      "
      ;;
  esac
```

Key points:
- Handles multiple file formats
- Performs various analysis operations
- Supports different output formats
- Uses Python for data processing

**Exercise 4: Extend the Data Analyzer**
1. Add support for more file formats
2. Add visualization options
3. Add data filtering capabilities
4. Add export to different formats

## Configuring MCP

Let's set up our configuration using the built-in configuration management tools:

1. **Initialize Configuration**
   ```bash
   # Create a new configuration file
   go-go-mcp config init
   
   # Edit it in your default editor to review
   go-go-mcp config edit
   ```

2. **Create Profiles for Different Environments**
   ```bash
   # Create development profile
   go-go-mcp config add-profile development "Development environment with debug tools"
   
   # Add tool directories
   go-go-mcp config add-tool development --dir ./tools/system
   go-go-mcp config add-tool development --dir ./tools/data
   
   # Create production profile
   go-go-mcp config add-profile production "Production environment with strict controls"
   
   # Add production tools
   go-go-mcp config add-tool production --dir /opt/go-go-mcp/tools
   
   # Create staging by duplicating development
   go-go-mcp config duplicate-profile development staging "Staging environment"
   ```

3. **Review Configuration**
   ```bash
   # List all profiles
   go-go-mcp config list-profiles
   
   # Show development profile configuration
   go-go-mcp config show-profile development
   
   # Show production profile configuration
   go-go-mcp config show-profile production
   ```

4. **Set Default Profile**
   ```bash
   # Set development as default for local work
   go-go-mcp config set-default-profile development
   ```

Now let's create a configuration file that organizes tools by their use cases:

## Working with Configuration Files

Configuration files in MCP allow you to:
- Define different tool categories and use cases
- Set default parameters for commands
- Control access to commands and parameters
- Configure tool and prompt directories

### Configuration File Structure

A typical MCP configuration file has these main sections:

1. **Version and Default Profile**
   ```yaml
   version: "1"
   defaultProfile: all
   ```

2. **Profile Definitions**
   ```yaml
   profiles:
     system:
       description: "System monitoring tools"
       # Profile-specific settings
     
     data:
       description: "Data analysis tools"
       # Data analysis settings
   ```

3. **Tool Configuration**
   ```yaml
   tools:
     directories:     # Load all tools from directories
       - path: ./tools/system
         defaults:    # Default parameters by layer, overrides the defaults defined in the YAML/layer
           default:   # Default layer parameters
             timeout: 30s
           output:    # Output layer parameters
             format: json
         whitelist:   # Allowed parameters by layer, only expose these parameters to the caller / schema
           default:   # Default layer whitelist
             - timeout
             - debug blacklist:   # Blocked parameters by layer, do not expose these parameters to the caller / schema
           default:   # Default layer blacklist
             - api_key
     
     files:          # Load specific tool files
       - path: ./tools/data/special-analyzer.yaml
         overrides:   # Force parameter values by layer
           default:   # Default layer overrides
             workers: 4
           output:    # Output layer overrides
             format: html
   ```

4. **Prompt Configuration**
   ```yaml
   prompts:
     directories:
       - path: ./prompts/system
         defaults:
           default:   # Default layer for prompts
             temperature: 0.7
   ```

### Understanding Parameter Layers

MCP tools use Glazed's parameter layer system to organize their parameters. Each tool can define multiple parameter layers. The `default` layer corresponds to the flags defined in the tool's YAML file, while other layers can be used for specific purposes. 

Currently, shell commands only support the `default` layer.

The configuration file allows you to control these layers through defaults, overrides, whitelists, and blacklists. For example:

```yaml
tools:
  directories:
    - path: ./tools/system
      defaults:            # defaults override the defaults defined in the YAML/layer
        default:           # Default layer settings (from YAML flags)
          metrics: [cpu, memory]
          format: text
        output:           # Output layer settings
          pretty: true
          colors: true
      whitelist:          # Whitelist exposes only these parameters to the caller / schema
        default:          # Default layer allowed params
          - metrics
          - format
```

Let's create a comprehensive configuration that demonstrates this:

```yaml
version: "1"
defaultProfile: all

profiles:
  all:
    description: "All available tools"
    tools:
      directories:
        - path: ./tools/system
          defaults:
            default:              # From YAML flags
              metrics: [cpu, memory]
              format: text
        - path: ./tools/data
        - path: ./tools/calendar
          defaults:
            default:             # From YAML flags
              calendar_id: "primary"
              timezone: "UTC"
            notification:        # Custom layer for notifications
              remind: true
              advance: "1h"
        - path: ./tools/web

  system:
    description: "System monitoring and maintenance tools"
    tools:
      directories:
        - path: ./tools/system

  data:
    description: "Data analysis and processing tools"
    tools:
      directories:
        - path: ./tools/data
          whitelist:
            default:             # From YAML flags
              - format
              - operations

  calendar:
    description: "Calendar management tools"
    tools:
      directories:
        - path: ./tools/calendar
          defaults:
            scheduling:         # Custom layer for scheduling
              buffer: "15m"
              working_hours: "9-17"

  web:
    description: "Web scraping and API tools"
    tools:
      directories:
        - path: ./tools/web
          defaults:
            default:             # From YAML flags
              timeout: 60
              format: json
            http:              # Custom layer for HTTP
              retry: 3
              backoff: "exponential"
            proxy:            # Custom layer for proxy settings
              enabled: true
              type: "socks5"
```

Key concepts:
1. **Profile Selection**: Each profile can have different:
   - Tool categories
   - Default parameters for each layer
   - Security controls per layer
   - Use case specific settings

2. **Parameter Management**:
   - `defaults`: Set suggested values for each layer
   - `overrides`: Force specific values for each layer
   - `whitelist`: Allow specific parameters per layer
   - `blacklist`: Block specific parameters per layer

3. **Security Considerations**:
   - Use whitelists for sensitive operations
   - Block dangerous parameters by layer
   - Control access to tools
   - Manage environment variables

**Exercise 5: Configuration Practice**
1. Create a combined profile that:
   - Includes both system monitoring and data analysis tools
   - Sets appropriate defaults for different parameter layers
   - Configures proper parameter restrictions per layer
2. Add a custom tool directory for specialized tools
3. Configure parameter validation for each layer

## Running the Server

Now we can start the MCP server with different tool sets based on our needs:

### File Watching

The server automatically watches configured repositories and files for changes. This means you can:
- Add or remove tools while the server is running
- Modify tool configurations in real-time
- Update tool implementations without restarts

File watching is enabled by default and can be controlled through the configuration:

```yaml
profiles:
  development:
    tools:
      directories:
        - path: ./tools/system
          watch: true  # Enable watching for this directory
        - path: ./tools/static
          watch: false # Disable watching for static tools
      files:
        - path: ./tools/special-tool.yaml
          watch: true  # Watch individual files too
```

When changes are detected:
1. The server reloads affected tools
2. New tools become immediately available
3. Removed tools are unregistered
4. Modified tools are updated in-place

### All Tools

Start the server with all available tools:

```bash
go-go-mcp start \
  --config-file config.yaml \
  --profile all \
  --transport sse \
  --port 3001

# In another terminal, watch tools being loaded
tail -f go-go-mcp.log

# Add a new tool while server is running
cp new-tool.yaml tools/system/
# Watch the log to see it being loaded
```

### System Monitoring

Start the server with just system monitoring tools:

```bash
go-go-mcp start \
  --config-file config.yaml \
  --profile system \
  --transport sse \
  --port 3001
```

### Direct Tool Interaction

The `server tools` commands allow you to interact with tools directly without starting a server:

```bash
# List tools in the system profile
go-go-mcp server tools list --profile system

# Expected output:
# NAME            DESCRIPTION
# system-monitor  Monitor system resources and performance
# disk-usage     Analyze disk space usage
# process-list   List and filter running processes

# Call system-monitor with JSON arguments
go-go-mcp server tools call system-monitor --json '{
  "format": "json",
  "metrics": ["cpu", "memory", "disk"],
  "watch": false
}'

# Call disk-usage with key-value arguments
go-go-mcp server tools call disk-usage \
  --args directory=/home,unit=GB,type=*.log

# Switch to calendar profile and list tools
go-go-mcp server tools list --profile calendar

# Expected output:
# NAME                 DESCRIPTION
# calendar-event      Manage calendar events and meetings
# calendar-availability Check calendar availability

# Create a calendar event
go-go-mcp server tools call calendar-event --json '{
  "title": "Team Meeting",
  "start_time": "2024-02-01 10:00",
  "duration": 60,
  "attendees": ["team@example.com"]
}'
```

This is particularly useful for:
- Testing tools during development
- Automation and scripting
- CI/CD pipelines
- Quick tool execution without server overhead

### Data Analysis

Start with data analysis tools:

```bash
go-go-mcp start \
  --config-file config.yaml \
  --profile data \
  --transport sse \
  --port 3001
```

### Calendar Tools

Start with calendar management tools:

```bash
go-go-mcp start \
  --config-file config.yaml \
  --profile calendar \
  --transport sse \
  --port 3001
```

The server will load only the tools relevant to the selected profile, making it easier to manage different use cases and maintain clear separation of concerns.

**Exercise 6: Profile Configuration**
1. Create a new profile that combines system monitoring and data analysis tools
2. Add custom defaults for the combined profile
3. Configure appropriate parameter whitelists
4. Test the profile with different tool combinations

## Testing with the Client

Now let's test our setup using the go-go-mcp client.

### List Available Tools

```bash
# List all available tools
go-go-mcp client \
  --transport sse \
  --server http://localhost:3001 \
  tools list

# Expected output:
# System Tools:
# - system-monitor: Monitor system resources and performance
# - disk-usage: Analyze disk space usage
#
# Data Tools:
# - data-analyzer: Analyze data files and generate reports
#
# Calendar Tools:
# - calendar-event: Manage calendar events and meetings
#
# Web Tools:
# - web-scraper: Fetch and process web content
```

### Testing System Monitoring Tools

```bash
# Monitor system resources in JSON format
go-go-mcp client \
  --transport sse \
  --server http://localhost:3001 \
  tools call system-monitor \
  --args '{"format":"json","metrics":["cpu","memory","disk"]}'

# Check disk usage with specific filters
go-go-mcp client \
  --transport sse \
  --server http://localhost:3001 \
  tools call disk-usage \
  --args '{"directory":"/home","unit":"GB","type":"*.log"}'
```

### Testing Calendar Tools

```bash
# Create a new calendar event
go-go-mcp client \
  --transport sse \
  --server http://localhost:3001 \
  tools call calendar-event \
  --args '{
    "title": "Team Meeting",
    "start_time": "2025-02-01 10:00",
    "duration": 60,
    "attendees": ["team@example.com"]
  }'

# Check calendar availability
go-go-mcp client \
  --transport sse \
  --server http://localhost:3001 \
  tools call calendar-availability \
  --args '{
    "date": "2025-02-01",
    "duration": 60
  }'
```

### Testing Data Analysis Tools

```bash
# Analyze a CSV file
go-go-mcp client \
  --transport sse \
  --server http://localhost:3001 \
  tools call data-analyzer \
  --args '{
    "input_file": "data.csv",
    "format": "csv",
    "operations": ["summary", "correlation"],
    "output_format": "html"
  }'

# Generate visualizations
go-go-mcp client \
  --transport sse \
  --server http://localhost:3001 \
  tools call data-visualizer \
  --args '{
    "input_file": "data.csv",
    "plot_type": "scatter",
    "x_column": "date",
    "y_column": "value"
  }'
```

### Testing Web Tools

```bash
# Fetch and process web content
go-go-mcp client \
  --transport sse \
  --server http://localhost:3001 \
  tools call web-scraper \
  --args '{
    "url": "https://example.com",
    "format": "json",
    "extract": ["title", "main_content"]
  }'

# Monitor a website
go-go-mcp client \
  --transport sse \
  --server http://localhost:3001 \
  tools call web-monitor \
  --args '{
    "url": "https://example.com",
    "interval": 300,
    "check": "status"
  }'
```

**Exercise 7: Tool Testing**
1. Create a test suite for each tool category
2. Test tool combinations (e.g., fetch web data and analyze it)
3. Test error handling and edge cases
4. Create a test report template

Remember to:
- Test each tool with various parameter combinations
- Verify error handling and validation
- Check output formats and data consistency
- Monitor resource usage during testing
- Document any issues or limitations found

For more details on specific topics, see:
- [MCP Protocol](01-mcp-protocol.md)
- [Shell Commands](02-shell-commands.md)
- [Configuration File](01-config-file.md)

## Next Steps

1. Create more specialized tools for your needs
2. Add authentication and authorization
3. Set up monitoring and logging
4. Create custom prompts
5. Integrate with your CI/CD pipeline

Remember to:
- Keep development and production configurations separate
- Use appropriate security measures in production
- Monitor server logs for issues
- Regularly test all tools
- Update configurations as needs change

**Final Exercise: Complete System**
1. Create a complete development environment with:
   - At least 3 custom tools
   - Proper configuration
   - Test suite
   - Documentation
2. Deploy to a staging environment
3. Document the deployment process

For more details on specific topics, see:
- [MCP Protocol](01-mcp-protocol.md)
- [Shell Commands](02-shell-commands.md)
- [Configuration File](01-config-file.md) 