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