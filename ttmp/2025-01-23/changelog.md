## Convert Resources Commands to Glazed

Converted the resources list and read commands to use the Glazed framework for better parameter handling and output formatting.

- Implemented ListResourcesCommand with structured output using GlazeProcessor
- Implemented ReadResourceCommand with formatted text output using Writer
- Added proper parameter layers and settings structs 

## Convert Tools Commands to Glazed

Converted the tools list and call commands to use the Glazed framework for better parameter handling and output formatting.

- Implemented ListToolsCommand with structured output using GlazeProcessor
- Implemented CallToolCommand with formatted text output using Writer
- Added proper parameter layers and settings structs
- Improved error handling and output formatting for tool results 

## Add Key-Value Arguments to Tool Call Command

Enhanced the tool call command to support both JSON string and key-value pair arguments:

- Added key-value parameter type for simpler argument passing
- JSON string arguments take precedence over key-value pairs
- Improved argument handling in settings struct 

## Rename Tool Call Command Parameters

Renamed tool call command parameters for better clarity:

- Renamed --args to --json for JSON string input
- Renamed --key-value to --args for key-value pairs
- Updated help text to reflect new parameter names 

## Add Schema Output to Tools List Command

Enhanced the tools list command to include tool schemas:

- Added schema field to output rows
- Schema is formatted as indented JSON for readability 

## Improve Schema Output in Tools List Command

Enhanced schema output handling in tools list command:

- Added JSON validation step for schemas
- Improved error handling for malformed schemas
- Ensured consistent JSON formatting 

# Shell Command Implementation

Added a new ShellCommand type that allows executing shell commands and scripts with templated arguments.

- Added ShellCommand type with support for both shell scripts and command lists
- Implemented template processing for command arguments and environment variables
- Added working directory and stderr capture options
- Included YAML loading support for shell command definitions 

# Add Shell Command Examples

Added example YAML configurations for the ShellCommand type:

- backup-db: Example of using AWS CLI with environment variables
- process-logs: Example of shell script for log processing
- docker-compose: Example of running docker-compose with templated environment
- git-sync: Example of complex shell script for git operations 

# Add Run Command Support

Added support for running shell commands directly from YAML files:

- Added run-command functionality to main.go
- Created ShellCommandLoader for loading YAML command definitions
- Support for running commands with all their flags and arguments
- Similar to sqleton's run-command implementation 

# Add Shell Commands Documentation

Added comprehensive documentation for shell commands:

- Detailed explanation of command structure and types
- Multiple examples covering various use cases
- Best practices and debugging tips
- Common issues and solutions 