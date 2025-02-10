package layers

import (
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

type ClientSettings struct {
	Transport string   `glazed.parameter:"transport"`
	Server    string   `glazed.parameter:"server"`
	Command   []string `glazed.parameter:"command"`
}

const ClientLayerSlug = "mcp-client"

func NewClientParameterLayer() (layers.ParameterLayer, error) {
	return layers.NewParameterLayer(ClientLayerSlug, "MCP Client Settings",
		layers.WithParameterDefinitions(
			parameters.NewParameterDefinition(
				"transport",
				parameters.ParameterTypeString,
				parameters.WithHelp("Transport type (command or sse)"),
				parameters.WithDefault("command"),
			),
			parameters.NewParameterDefinition(
				"server",
				parameters.ParameterTypeString,
				parameters.WithHelp("Server URL for SSE transport"),
				parameters.WithDefault("http://localhost:3001"),
			),
			parameters.NewParameterDefinition(
				"command",
				parameters.ParameterTypeStringList,
				parameters.WithHelp("Command and arguments for command transport"),
				parameters.WithDefault([]string{"mcp-server", "start", "--transport", "stdio"}),
			),
		),
	)
}
