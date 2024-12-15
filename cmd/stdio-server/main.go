package main

import (
	"io"

	"github.com/go-go-golems/go-mcp/pkg/server"
	"github.com/rs/zerolog/log"
)

func main() {
	srv := server.NewServer()

	// Register the simple provider
	simpleProvider := NewSimpleProvider()
	srv.GetRegistry().RegisterPromptProvider(simpleProvider)

	if err := srv.Start(); err != nil && err != io.EOF {
		log.Fatal().Err(err).Msg("Server error")
	}
}
