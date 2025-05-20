# AmpCode MCP Server Integration

## Overview

This integration allows management of AmpCode MCP servers through the go-go-mcp user interface. AmpCode MCP servers are configured in the VS Code/Cursor settings file at `~/.config/Cursor/User/settings.json`.

## Implementation Details

1. **AmpCode Config Editor**: 
   - Added `pkg/config/ampcode.go` with an implementation of the `ServerConfigEditor` interface
   - Handles reading/writing to the VS Code/Cursor settings.json file
   - Supports VS Code's JSONC format (JSON with comments and trailing commas) using the `github.com/tailscale/hujson` library
   - **Creates properly formatted JSON** when writing back to the file
   - Preserves all other settings in the file while modifying just the MCP server settings

2. **UI Integration**:
   - Added "AmpCode Config" option to the main TUI menu
   - Added support for enabling/disabling MCP servers via the disabled flag
   - Server configuration follows the same patterns as Cursor and Claude configurations

3. **UI Definitions**:
   - Added `examples/ui/ampcode-mcp-servers.yaml` for web UI access to the AmpCode MCP server management

## Usage

### Terminal UI

To manage AmpCode MCP servers via the Terminal UI:

```bash
go-go-mcp ui
```

Then select "AmpCode Config" from the menu to view, add, edit, delete, enable, or disable servers.

### Web UI

To manage AmpCode MCP servers via the Web UI:

```bash
ui-server start examples/
```

Then navigate to http://localhost:8080/pages/ui/ampcode-mcp-servers to access the AmpCode MCP server management page.

## Configuration Format

The AmpCode MCP servers are stored in the VS Code/Cursor settings file under the `amp.mcpServers` section:

```jsonc
{
    // AmpCode MCP server configurations
    "amp.mcpServers": {
        "playwright": {
            "command": "npx",
            "args": [
                "@playwright/mcp@latest"
            ],
            "disabled": true, // Currently disabled
        },
        "blender": {
            "command": "uvx",
            "args": [
                "blender-mcp"
            ],
            "env": {},
            "disabled": false, // Currently enabled
        },
    },
    "update.releaseTrack": "prerelease",
    "amp.tools.disable": [
        "generate_hyper3d_model_via_images",
        "poll_rodin_job_status",
        // Other disabled tools
    ],
}
```

Each server has these properties:
- `command`: The executable to run
- `args`: An array of arguments
- `env`: (Optional) Environment variables
- `disabled`: Boolean flag indicating if the server is disabled