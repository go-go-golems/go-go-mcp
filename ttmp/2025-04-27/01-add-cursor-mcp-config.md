### commands

// This example demonstrated an MCP server using the stdio format
// Cursor automatically runs this process for you
// This uses a Node.js server, ran with `npx`
{
  "mcpServers": {
    "server-name": {
      "command": "npx",
      "args": ["-y", "mcp-server"],
      "env": {
        "API_KEY": "value"
      }
    }
  }
}

### sse

// This example demonstrated an MCP server using the SSE format
// The user should manually setup and run the server
// This could be networked, to allow others to access it too
{
  "mcpServers": {
    "server-name": {
      "url": "http://localhost:3000/sse",
      "env": {
        "API_KEY": "value"
      }
    }
  }
}

### location

Project Configuration

For tools specific to a project, create a .cursor/mcp.json file in your project directory. This allows you to define MCP servers that are only available within that specific project.
Global Configuration

For tools that you want to use across all projects, create a \~/.cursor/mcp.json file in your home directory. This makes MCP servers available in all your Cursor workspaces.