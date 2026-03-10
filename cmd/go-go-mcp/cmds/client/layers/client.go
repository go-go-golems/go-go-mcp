package layers

import (
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
)

type ClientSettings struct {
	Transport string   `glazed:"transport"`
	Server    string   `glazed:"server"`
	Command   []string `glazed:"command"`
}

const ClientLayerSlug = "mcp-client"

func NewClientParameterLayer() (schema.Section, error) {
	return schema.NewSection(ClientLayerSlug, "MCP Client Settings",
		schema.WithFields(
			fields.New(
				"transport",
				fields.TypeString,
				fields.WithHelp("Transport type (command or sse)"),
				fields.WithDefault("command"),
			),
			fields.New(
				"server",
				fields.TypeString,
				fields.WithHelp("Server URL for SSE transport"),
				fields.WithDefault("http://localhost:3001"),
			),
			fields.New(
				"command",
				fields.TypeStringList,
				fields.WithHelp("Command and arguments for command transport (starts go-go-mcp in stdio mode per default)"),
				fields.WithDefault([]string{"mcp", "server", "start", "--transport", "stdio"}),
			),
		),
	)
}
