# Add Environment File Support

## Overview

This document outlines the implementation plan for adding environment file support to the MCP server and shell commands. This enhancement will allow users to:

1. Pass environment variables via a file using the `--env-file` flag to the server command
2. Define environment files directly in the command YAML definitions
3. Load these environment variables into both the server and the executed commands

## Current Implementation

Currently, environment variables can be set:
- Directly in the command YAML via the `environment` map
- Through environment variable templates that reference command arguments

However, there's no built-in support for loading environment variables from external files.

## Proposed Enhancement

We will implement the following features:

1. **Server-level environment file flag**:
   - Add a global `--env-file` flag to load environment variables from a file
   - Make these variables available to all executed commands

2. **Command-specific environment files**:
   - Add an `env-file` field in the shell command YAML definition
   - Support multiple environment files per command

3. **Environment variable loading and precedence**:
   - System environment variables (lowest precedence)
   - Server-level environment file variables
   - Command-specific environment file variables
   - Command-specific environment map (highest precedence)

## Implementation Locations

The changes will be implemented in:

1. `cmd/go-go-mcp/cmds/server/server.go` - For adding the global `--env-file` flag
2. `pkg/cmds/cmd.go` - For extending the `ShellCommand` struct and updating environment handling
3. `pkg/cmds/loader.go` - For parsing the new `env-file` field from YAML

## Implementation Details

### 1. Server-Level Environment File Support

Add a global flag to `ServerCmd` in `cmd/go-go-mcp/cmds/server/server.go`:

```go
type ServerOptions struct {
    EnvFile string
}

var serverOptions = &ServerOptions{}

func init() {
    ServerCmd.PersistentFlags().StringVar(&serverOptions.EnvFile, "env-file", "", "Path to environment file")
    // ... existing initialization ...
}
```

### 2. Extend ShellCommand Structure

Update the `ShellCommandDescription` and `ShellCommand` structs in `pkg/cmds/cmd.go`:

```go
type ShellCommandDescription struct {
    // ... existing fields ...
    EnvFiles []string `yaml:"env-files,omitempty"`
}

type ShellCommand struct {
    // ... existing fields ...
    EnvFiles []string
}
```

### 3. Environment File Loading Function

Create a utility function to load environment variables from a file:

```go
// loadEnvFile reads a file containing environment variables (KEY=VALUE format)
// and returns them as a map
func loadEnvFile(filePath string) (map[string]string, error) {
    env := make(map[string]string)
    
    // Implementation details...
    
    return env, nil
}
```

### 4. Update Environment Variable Handling

Modify the `ExecuteCommand` method to incorporate environment variables from files:

```go
// ExecuteCommand handles the actual command execution
func (c *ShellCommand) ExecuteCommand(
    ctx context.Context,
    args map[string]interface{},
    w io.Writer,
    serverEnvFile string,
) error {
    // ... existing code ...
    
    // Build environment with proper precedence
    env := os.Environ()
    
    // 1. Add server-level env file variables if specified
    if serverEnvFile != "" {
        serverEnv, err := loadEnvFile(serverEnvFile)
        if err != nil {
            log.Warn().Err(err).Str("file", serverEnvFile).Msg("Failed to load server environment file")
        } else {
            for k, v := range serverEnv {
                env = append(env, fmt.Sprintf("%s=%s", k, v))
            }
        }
    }
    
    // 2. Add command-specific env file variables
    for _, envFile := range c.EnvFiles {
        cmdEnv, err := loadEnvFile(envFile)
        if err != nil {
            log.Warn().Err(err).Str("file", envFile).Msg("Failed to load command environment file")
            continue
        }
        
        for k, v := range cmdEnv {
            env = append(env, fmt.Sprintf("%s=%s", k, v))
        }
    }
    
    // 3. Add command-specific environment variables (highest precedence)
    for k, v := range c.Environment {
        processed, err := c.processTemplate(v, args)
        if err != nil {
            log.Error().Err(err).Str("environment_variable", k).Msg("failed to process environment variable template")
            return errors.Wrapf(err, "failed to process environment variable template: %s", k)
        }
        env = append(env, fmt.Sprintf("%s=%s", k, processed))
    }
    
    cmd.Env = env
    
    // ... rest of existing code ...
}
```

### 5. Update Command Execution Flow

Modify the execution flow to pass the server environment file to commands:

```go
// RunIntoWriter implements the WriterCommand interface
func (c *ShellCommand) RunIntoWriter(
    ctx context.Context,
    parsedLayers *layers.ParsedLayers,
    w io.Writer,
) error {
    // Get arguments from parsed layers
    args := parsedLayers.GetDefaultParameterLayer().Parameters.ToMap()
    
    // Get server env file from context
    serverEnvFile := ""
    if v := ctx.Value("serverEnvFile"); v != nil {
        if envFile, ok := v.(string); ok {
            serverEnvFile = envFile
        }
    }
    
    log.Debug().Interface("args", args).Msg("executing command")
    err := c.ExecuteCommand(ctx, args, w, serverEnvFile)
    if err != nil {
        log.Error().Interface("args", args).Err(err).Msg("failed to execute command")
        return errors.Wrap(err, "failed to execute command")
    }
    
    return nil
}
```

### 6. Update YAML Loader

Modify the YAML loader in `pkg/cmds/loader.go` to handle the new `env-files` field:

```go
func LoadShellCommandFromYAML(data []byte) (*ShellCommand, error) {
    var desc ShellCommandDescription
    if err := yaml.Unmarshal(data, &desc); err != nil {
        return nil, errors.Wrap(err, "failed to unmarshal YAML")
    }
    
    cmdDesc := cmds.NewCommandDescription(
        desc.Name,
        cmds.WithShort(desc.Short),
        cmds.WithLong(desc.Long),
        cmds.WithFlags(desc.Flags...),
        cmds.WithArguments(desc.Arguments...),
        cmds.WithLayersList(desc.Layers...),
    )
    
    return NewShellCommand(
        cmdDesc,
        WithShellScript(desc.ShellScript),
        WithCommand(desc.Command),
        WithCwd(desc.Cwd),
        WithEnvironment(desc.Environment),
        WithCaptureStderr(desc.CaptureStderr),
        WithEnvFiles(desc.EnvFiles),
    )
}
```

## Environment File Format

We will support a simple format for environment files:

```
# Comments start with #
KEY1=VALUE1
KEY2=VALUE2

# Empty lines are ignored
KEY3="VALUE WITH SPACES"
```

## Usage Examples

### Command Line Usage

```bash
# Start server with global environment file
go-go-mcp server start --env-file=./server.env
```

### Command YAML Definition

```yaml
name: example-command
short: An example command
shell-script: |
  echo "Running with environment variables"
  echo "DATABASE_URL: $DATABASE_URL"
env-files:
  - ./common.env
  - ./specific.env
environment:
  CUSTOM_VAR: "value from direct setting"
```

## Security Considerations

- Environment files may contain sensitive information (passwords, API keys, etc.)
- Files should have appropriate permissions (readable only by the intended users)
- Consider implementing support for encrypted environment files in the future 