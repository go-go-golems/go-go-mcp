package main

import (
	"io"

	"github.com/go-go-golems/go-mcp/pkg/prompts"
	"github.com/go-go-golems/go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-mcp/pkg/resources"
	"github.com/go-go-golems/go-mcp/pkg/server"
	"github.com/go-go-golems/go-mcp/pkg/tools"
	"github.com/rs/zerolog/log"
)

func main() {
	srv := server.NewServer()

	// Create registries
	promptRegistry := prompts.NewRegistry()
	resourceRegistry := resources.NewRegistry()
	toolRegistry := tools.NewRegistry()

	// Register a simple prompt directly
	promptRegistry.RegisterPrompt(protocol.Prompt{
		Name:        "simple",
		Description: "A simple prompt that can take optional context and topic arguments",
		Arguments: []protocol.PromptArgument{
			{
				Name:        "context",
				Description: "Additional context to consider",
				Required:    false,
			},
			{
				Name:        "topic",
				Description: "Specific topic to focus on",
				Required:    false,
			},
		},
	})

	// Register registries with the server
	srv.GetRegistry().RegisterPromptProvider(promptRegistry)
	srv.GetRegistry().RegisterResourceProvider(resourceRegistry)
	srv.GetRegistry().RegisterToolProvider(toolRegistry)

	if err := srv.Start(); err != nil && err != io.EOF {
		log.Fatal().Err(err).Msg("Server error")
	}
}
