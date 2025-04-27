# Add JSON Command Arguments Storage to Shell Commands

## Overview

This document outlines the plan to enhance the shell command functionality to store command-line arguments and flags as a JSON file. The path to this JSON file will be passed to executed scripts via the `MCP_INPUT_ARGS` environment variable.

## Current Implementation

Currently, the `ShellCommand` in `pkg/cmds/cmd.go` processes arguments and passes them to shell scripts through:
1. Template rendering in the script content
2. Environment variables that are explicitly defined in the command configuration

## Proposed Enhancement

We will modify the `ExecuteCommand` function to:
1. Serialize the entire `args` map (containing all parsed flags/arguments) to a JSON file
2. Store this JSON file in a temporary location
3. Pass the path to this file via the `MCP_INPUT_ARGS` environment variable to executed scripts

## Implementation Locations

The changes will be implemented in:
- `pkg/cmds/cmd.go` - Specifically in the `ExecuteCommand` method

## Implementation Details

1. **Serialize Arguments to JSON**:
   - Take the `args` map (containing all parsed command-line arguments)
   - Convert it to a JSON string

2. **Create Temporary File**:
   - Create a temporary file with a meaningful name (e.g., `args-*.json`)
   - Write the JSON content to this file
   - Ensure proper file permissions

3. **Set Environment Variable**:
   - Add the `MCP_INPUT_ARGS` environment variable to the command's environment
   - Set its value to the absolute path of the temporary JSON file

4. **Cleanup**:
   - Ensure the temporary file is properly cleaned up after script execution

## Benefits

1. **Structured Data Access**: Scripts can parse the JSON file to access all command arguments in a structured format
2. **Avoids Template Complexity**: Complex argument structures can be passed without complicated templating logic
3. **Preserves Original Data Types**: JSON maintains the original data types of arguments

## Usage Example

After implementation, scripts can access all command arguments by reading and parsing the JSON file specified in the `MCP_INPUT_ARGS` environment variable:

```bash
# Inside a shell script
if [ -n "$MCP_INPUT_ARGS" ]; then
  # Parse the JSON file
  args=$(cat "$MCP_INPUT_ARGS")
  
  # Example: Use jq to extract specific arguments
  value=$(echo "$args" | jq -r '.some_flag')
  
  echo "Got value: $value"
fi
```

## Security Considerations

- The temporary file will be created with appropriate permissions to prevent unauthorized access
- The file will be deleted after script execution to avoid leaving sensitive data on disk 