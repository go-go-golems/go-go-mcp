package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// ExecuteCmd represents the execute command
type ExecuteCmd struct {
	*cmds.CommandDescription
}

// ExecuteSettings holds the configuration for the execute command
type ExecuteSettings struct {
	URL   string `glazed.parameter:"url"`
	Input string `glazed.parameter:"input"`
}

// Ensure ExecuteCmd implements BareCommand
var _ cmds.BareCommand = &ExecuteCmd{}

// NewExecuteCmd creates a new execute command
func NewExecuteCmd() (*ExecuteCmd, error) {
	return &ExecuteCmd{
		CommandDescription: cmds.NewCommandDescription(
			"execute",
			cmds.WithShort("Execute JavaScript code on the server"),
			cmds.WithLong(`
Execute JavaScript code on a running js-web-server instance.

The input can be either:
- Direct JavaScript code string
- Path to a JavaScript file

The code will be executed in the server's JavaScript runtime with access to:
- Geppetto APIs (Conversation, ChatStepFactory)
- Database bindings (db.query, db.exec)
- HTTP route registration (app.get, app.post, etc.)
- Console logging and global state

Examples:
  execute "console.log('Hello World')"
  execute ./scripts/test.js
  execute --url http://localhost:8081 "globalState.counter++"
			`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"url",
					parameters.ParameterTypeString,
					parameters.WithHelp("Server URL"),
					parameters.WithDefault("http://localhost:8080"),
					parameters.WithShortFlag("u"),
				),
			),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"input",
					parameters.ParameterTypeString,
					parameters.WithHelp("JavaScript code to execute or path to JavaScript file"),
					parameters.WithRequired(true),
				),
			),
		),
	}, nil
}

// Run implements the BareCommand interface
func (c *ExecuteCmd) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	// Parse settings from layers
	s := &ExecuteSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "failed to parse execute settings")
	}

	// Determine if input is a file or direct code
	var code string
	if fileInfo, err := os.Stat(s.Input); err == nil && !fileInfo.IsDir() {
		// Input is a file
		data, err := os.ReadFile(s.Input)
		if err != nil {
			return errors.Wrapf(err, "failed to read file: %s", s.Input)
		}
		code = string(data)
		log.Info().Str("file", s.Input).Msg("Executing file")
	} else {
		// Input is direct code
		code = s.Input
		log.Info().Str("code", truncateCode(code, 100)).Msg("Executing code")
	}

	// Construct the execute endpoint URL
	executeURL := strings.TrimSuffix(s.URL, "/") + "/v1/execute"
	
	log.Debug().Str("url", executeURL).Msg("Sending request to server")
	
	// Send POST request to the server
	resp, err := http.Post(executeURL, "application/javascript", strings.NewReader(code))
	if err != nil {
		return errors.Wrapf(err, "failed to execute code on server: %s", executeURL)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "failed to read response")
	}

	// Output the results
	fmt.Printf("Status: %s\n", resp.Status)
	if len(body) > 0 {
		fmt.Printf("Response: %s\n", string(body))
	}

	// Return error if the server returned an error status
	if resp.StatusCode >= 400 {
		return fmt.Errorf("server returned error status: %s", resp.Status)
	}

	return nil
}

// truncateCode truncates code for logging purposes
func truncateCode(code string, maxLen int) string {
	if len(code) <= maxLen {
		return code
	}
	return code[:maxLen] + "..."
}
