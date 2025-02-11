---
Title: Go Go MCP config file
Slug: config-file
Short: Learn how to create a configuration file for go-go-mcp.
Topics:
  - config
  - tools
  - prompts
Commands:
  - start
Flags:
  - config-file
  - profile
  - repositories
  - debug
  - tracing-dir
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

This tutorial will guide you through creating and using configuration files in go-go-mcp. We'll cover all major features and provide practical examples.

## Table of Contents

1. [Basic Configuration](#basic-configuration)
2. [Profiles](#profiles)
3. [Tool Configuration](#tool-configuration)
4. [Prompt Configuration](#prompt-configuration)
5. [Parameter Management](#parameter-management)
6. [Advanced Features](#advanced-features)
7. [Troubleshooting](#troubleshooting)

## Basic Configuration

Let's start with a minimal configuration file:

```yaml
version: "1"
defaultProfile: default
profiles:
  default:
    description: "Basic configuration"
    tools:
      directories:
        - path: ./tools
```

Save this as `config.yaml` and run:
```bash
go-go-mcp start --config-file config.yaml
```

This will:
1. Load tools from the `./tools` directory
2. Use default parameter settings
3. Start the server with the default profile

## Profiles

Profiles allow you to maintain different configurations for different environments:

```yaml
version: "1"
defaultProfile: development

profiles:
  development:
    description: "Development environment"
    tools:
      directories:
        - path: ./dev-tools
          defaults:
            default:
              debug: true
              verbose: true

  production:
    description: "Production environment"
    tools:
      directories:
        - path: /opt/go-go-mcp/tools
          whitelist:
            default:
              - timeout
              - retries
          overrides:
            default:
              timeout: 5s
              retries: 3
```

To use a specific profile:
```bash
go-go-mcp start --config-file config.yaml --profile production
```

## Tool Configuration

### Directory-Based Tools

Load tools from directories:

```yaml
tools:
  directories:
    - path: ./tools
      defaults:
        default:
          timeout: 30s
      blacklist:
        default:
          - sensitive_param
    
    - path: ./more-tools
      overrides:
        default:
          debug: true
```

### Individual Tool Files

Load specific tool files:

```yaml
tools:
  files:
    - path: ./special-tool.yaml
      defaults:
        default:
          memory: 512MB
    
    - path: ./custom-tool.yaml
      overrides:
        default:
          workers: 4
```

### External Commands (Planned)

Future support for external commands:

```yaml
tools:
  external_commands:
    - command: external-tool
      args: ["--config", "tool-config.yaml"]
      path: /usr/local/bin
```

## Prompt Configuration

### Directory-Based Prompts

Load prompts from directories:

```yaml
prompts:
  directories:
    - path: ./prompts
      defaults:
        default:
          temperature: 0.7
          max_tokens: 1000
```

### Individual Prompt Files

Load specific Pinocchio files:

```yaml
prompts:
  files:
    - path: ./custom-prompt.pinocchio
      overrides:
        default:
          temperature: 0.9
```

### Pinocchio Integration (Planned)

Future support for Pinocchio:

```yaml
prompts:
  pinocchio:
    command: pinocchio
    args: ["--config", "pinocchio.yaml"]
    path: /usr/local/bin
```

## Parameter Management

### Defaults

Set default parameter values:

```yaml
defaults:
  default:  # layer name
    timeout: 30s
    retries: 3
    debug: false
```

### Overrides

Force specific parameter values:

```yaml
overrides:
  default:
    model: gpt-4-turbo
    max_tokens: 2000
```

### Blacklist

Prevent certain parameters from being used:

```yaml
blacklist:
  default:
    - system_prompt
    - api_key
```

### Whitelist

Only allow specific parameters:

```yaml
whitelist:
  default:
    - timeout
    - retries
    - debug
```

## Advanced Features

### Complete Example

Here's a comprehensive configuration example:

```yaml
version: "1"
defaultProfile: development

profiles:
  development:
    description: "Development environment with debug tools"
    tools:
      directories:
        - path: ./dev-tools
          defaults:
            default:
              debug: true
              verbose: true
          blacklist:
            default:
              - api_key
      
      files:
        - path: ./special-debug.yaml
          overrides:
            default:
              memory: 1GB
    
    prompts:
      directories:
        - path: ./dev-prompts
          defaults:
            default:
              temperature: 0.9
              max_tokens: 2000

  production:
    description: "Production environment with strict controls"
    tools:
      directories:
        - path: /opt/go-go-mcp/tools
          whitelist:
            default:
              - timeout
              - retries
          overrides:
            default:
              timeout: 5s
              retries: 3
      
      files:
        - path: /opt/go-go-mcp/tools/special.yaml
    
    prompts:
      directories:
        - path: /opt/go-go-mcp/prompts
          overrides:
            default:
              temperature: 0.7
              max_tokens: 1000
```

## Troubleshooting

### Common Errors

1. **Profile Not Found**
   ```
   Error: profile 'staging' not found
   ```
   Solution: Check that the profile name matches exactly what's in your config file.

2. **Directory Not Found**
   ```
   Error: failed to get absolute path for ./tools: no such file or directory
   ```
   Solution: Ensure the directory exists and the path is correct relative to where you run the command.

3. **Invalid Tool File**
   ```
   Error: failed to load shell command from special-tool.yaml: invalid YAML format
   ```
   Solution: Verify the YAML syntax in your tool file.

### Best Practices

1. **Path Management**
   - Use relative paths for development
   - Use absolute paths for production
   - Validate paths before deployment

2. **Parameter Control**
   - Use blacklists for security-sensitive parameters
   - Use whitelists in production for strict control
   - Set sensible defaults for optional parameters

3. **Profile Organization**
   - Keep development and production profiles separate
   - Use descriptive profile names
   - Document profile purposes in descriptions

4. **Error Handling**
   - Check error messages carefully
   - Validate configuration before deployment
   - Keep backups of working configurations 