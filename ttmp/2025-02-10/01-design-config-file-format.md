Config file format:

- different profiles
- each profile has a list of sources (tools, prompts)
- sources have a set of configuration options to override parameters, a bit like in parka
  - in fact, use parka's Defaults, Overrides. We also only expose the DefaultSlug anyway (see @handlers.go in parka)

- tool sources can be:
  - individual files
  - directories
  - external go-go commands that are read out on the CLI

- prompts can be
  - individual pinocchio files
  - directories with pinocchio files
  - calling pinocchio 

# Detailed YAML Structure

The configuration file follows this structure, which has been implemented in `pkg/config/types.go`:

```yaml
# Version of the config format
version: "1"

# Default profile to use if none specified
defaultProfile: default

# Profile definitions
profiles:
  default:
    description: "Default profile with standard tools and prompts"
    tools:
      # Directory-based tools using Clay's repository system
      directories:
        - path: ./tools
          # Parameter filtering using Parka's system
          defaults:
            default:  # layer name
              timeout: 30s
              verbose: true
          overrides:
            default:
              model: gpt-4-turbo
          blacklist:
            default:
              - system_prompt
          whitelist:
            default:
              - param1
              - param2
        
      # Individual tool files
      files:
        - path: ./tools/special-tool.yaml
          defaults:
            default:
              memory: 512MB
          
      # External commands (not yet implemented)
      external_commands:
        - command: external-mcp-tool
          args: ["--config", "tool-config.yaml"]
          path: /usr/local/bin  # where to find the command
    
    prompts:
      # Directory-based prompts using Clay's repository
      directories:
        - path: ./prompts
          defaults:
            default:
              temperature: 0.7
          
      # Individual Pinocchio files
      files:
        - path: ./prompts/custom-prompt.pinocchio
          overrides:
            default:
              temperature: 0.9

      # Pinocchio integration (not yet implemented)
      pinocchio:
        command: pinocchio
        args: ["--config", "pinocchio.yaml"]
        path: /usr/local/bin

  development:
    description: "Development profile with additional debug tools"
    tools:
      directories:
        - path: ./dev-tools
          defaults:
            default:
              debug: true
              verbose: true
    
    prompts:
      directories:
        - path: ./dev-prompts
          overrides:
            default:
              temperature: 1.0

  production:
    description: "Locked down production profile"
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

## Key Features

1. **Profile System**
   - Multiple named profiles for different environments
   - Default profile selection
   - Each profile can have its own tools and prompts

2. **Tool Sources**
   - Directory-based tools using Clay's repository system
   - Individual tool files
   - External commands (planned)

3. **Prompt Sources**
   - Directory-based prompts using Clay's repository
   - Individual Pinocchio files
   - Pinocchio integration (planned)

4. **Parameter Management**
   - Uses Parka's parameter filtering system
   - Supports defaults, overrides, blacklists, and whitelists
   - Layer-based parameter organization

5. **Security Features**
   - Parameter blacklisting for sensitive values
   - Strict whitelisting for production environments
   - Profile-based access control

## Key Design Points

1. **Version Control**
   - The format includes a version number for future compatibility
   - Currently at version "1"
   - Allows for future format changes while maintaining backward compatibility

2. **Profile System**
   - Multiple named profiles for different environments/use cases
   - Default profile selection via `defaultProfile`
   - Each profile is independent and complete
   - No profile inheritance (removed from initial design)

3. **Repository Integration**
   - Uses Clay's repository system for both tools and prompts
   - Supports directory-based loading with Clay's Directory structure
   - Handles individual files separately from repositories
   - Converts Clay commands to MCP protocol format

4. **Parameter Management**
   - Integrates with Parka's parameter filtering system
   - Uses a single default layer for simplicity
   - Supports four types of parameter modifications:
     - Defaults: Set default values for parameters
     - Overrides: Force specific parameter values
     - Blacklist: Prevent certain parameters from being used
     - Whitelist: Only allow specific parameters

5. **Security Considerations**
   - Parameter blacklisting for sensitive values
   - Strict whitelisting option for production environments
   - Separate profiles for different security contexts
   - Path validation for file and directory sources

6. **Extensibility**
   - Planned support for external commands
   - Pinocchio integration for advanced prompt handling
   - Modular provider design for future extensions
   - Cursor-based pagination for large collections

7. **Implementation Details**
   - Uses Go's standard YAML parsing
   - Implements both pkg.ToolProvider and pkg.PromptProvider interfaces
   - Proper error handling and validation
   - Support for streaming command output

Then we can have a web UI that allows us to edit profiles / create new ones.

# Implementation Plan

## Core Types

The core types have been implemented in `pkg/config/types.go`:

```go
// Config represents the root configuration
type Config struct {
    Version        string              `yaml:"version"`
    DefaultProfile string              `yaml:"defaultProfile"`
    Profiles       map[string]*Profile `yaml:"profiles"`
}

// Profile represents a named configuration profile
type Profile struct {
    Description string         `yaml:"description"`
    Tools       *ToolSources   `yaml:"tools"`
    Prompts     *PromptSources `yaml:"prompts"`
}

// Common source configuration for both tools and prompts
type SourceConfig struct {
    Path      string          `yaml:"path"`
    Defaults  LayerParameters `yaml:"defaults,omitempty"`
    Overrides LayerParameters `yaml:"overrides,omitempty"`
    Blacklist ParameterFilter `yaml:"blacklist,omitempty"`
    Whitelist ParameterFilter `yaml:"whitelist,omitempty"`
}

// LayerParameters maps layer names to their parameter settings
type LayerParameters map[string]map[string]interface{}

// ParameterFilter maps layer names to lists of parameter names
type ParameterFilter map[string][]string

// ToolSources configures where tools are loaded from
type ToolSources struct {
    Directories      []SourceConfig `yaml:"directories,omitempty"`
    Files            []SourceConfig `yaml:"files,omitempty"`
    ExternalCommands []struct {
        Command      string   `yaml:"command"`
        Args         []string `yaml:"args,omitempty"`
        SourceConfig `yaml:",inline"`
    } `yaml:"external_commands,omitempty"`
}

// PromptSources configures where prompts are loaded from
type PromptSources struct {
    Directories []SourceConfig `yaml:"directories,omitempty"`
    Files       []SourceConfig `yaml:"files,omitempty"`
    Pinocchio   *struct {
        Command      string   `yaml:"command"`
        Args         []string `yaml:"args,omitempty"`
        SourceConfig `yaml:",inline"`
    } `yaml:"pinocchio,omitempty"`
}
```

## Provider Implementation

The providers have been implemented with a focus on error handling and validation:

### Tool Provider

```go
type ConfigToolProvider struct {
    repository    *repositories.Repository
    shellCommands map[string]cmds.Command
    toolConfigs   map[string]*SourceConfig
    helpSystem    *help.HelpSystem
    debug         bool
    tracingDir    string
}

type ConfigToolProviderOption func(*ConfigToolProvider) error
```

Key features:
- Error-returning functional options for configuration
- Strict validation of profiles and paths
- Proper error propagation from file loading
- Integration with Clay's help system
- Support for both directory and file-based tools

### Error Handling

The provider implements comprehensive error handling:

1. **Profile Validation**:
   ```go
   if !ok {
       return errors.Errorf("profile %s not found", profile)
   }
   ```

2. **Path Resolution**:
   ```go
   absPath, err := filepath.Abs(dir.Path)
   if err != nil {
       return errors.Wrapf(err, "failed to get absolute path for %s", dir.Path)
   }
   ```

3. **Command Loading**:
   ```go
   shellCmd, err := mcp_cmds.LoadShellCommand(absPath)
   if err != nil {
       return errors.Wrapf(err, "failed to load shell command from %s", file.Path)
   }
   ```

### Integration Points

The configuration system integrates with several key components:

1. **Clay Repository System**:
   - Directory-based tool loading
   - Command collection and execution
   - Help system integration

2. **Parka Parameter Management**:
   - Parameter filtering
   - Defaults and overrides
   - Layer-based organization

3. **Server Integration**:
   - Profile-based provider creation
   - Optional configuration loading
   - Command-line integration

## Usage Example

```yaml
version: "1"
defaultProfile: development

profiles:
  development:
    description: "Development environment configuration"
    tools:
      directories:
        - path: ./dev-tools
          defaults:
            default:
              debug: true
          blacklist:
            default:
              - sensitive_param
    prompts:
      directories:
        - path: ./dev-prompts
          overrides:
            default:
              temperature: 0.9
```

Command-line usage:
```bash
go-go-mcp server start --config-file config.yaml --profile development
```

## Error Handling Examples

1. **Missing Profile**:
```
Error: profile 'production' not found
```

2. **Invalid Directory**:
```
Error: failed to get absolute path for ./tools: no such file or directory
```

3. **Invalid Tool File**:
```
Error: failed to load shell command from special-tool.yaml: invalid YAML format
```

## Next Steps

1. **External Commands**:
   - Implement command execution
   - Add timeout and security controls
   - Support streaming output

2. **Pinocchio Integration**:
   - Add proper file loading
   - Support parameter passing
   - Handle execution errors

3. **Web UI**:
   - Profile management interface
   - Configuration editor
   - Real-time validation

4. **Testing**:
   - Unit tests for configuration loading
   - Integration tests for providers
   - Error handling test cases