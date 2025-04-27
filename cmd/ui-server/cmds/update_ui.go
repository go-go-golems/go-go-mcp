package cmds

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"gopkg.in/yaml.v3"
)

// UpdateUISettings contains the settings for the update-ui command
type UpdateUISettings struct {
	Host string `glazed.parameter:"host"`
	Port int    `glazed.parameter:"port"`
	File string `glazed.parameter:"file"`
}

// UpdateUICommand is a command to update the UI with a YAML file
type UpdateUICommand struct {
	*cmds.CommandDescription
}

// NewUpdateUICommand creates a new update-ui command
func NewUpdateUICommand() (cmds.BareCommand, error) {
	cmd := &UpdateUICommand{
		CommandDescription: cmds.NewCommandDescription(
			"update-ui",
			cmds.WithShort("Update the UI with a YAML file"),
			cmds.WithLong("Send a YAML file to the UI server to update the UI"),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"file",
					parameters.ParameterTypeString,
					parameters.WithHelp("Path to the YAML file"),
					parameters.WithRequired(true),
				),
			),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"host",
					parameters.ParameterTypeString,
					parameters.WithHelp("Host of the UI server"),
					parameters.WithDefault("localhost"),
				),
				parameters.NewParameterDefinition(
					"port",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Port of the UI server"),
					parameters.WithDefault(8080),
				),
			),
		),
	}

	return cmd, nil
}

// Run implements the BareCommand interface
func (c *UpdateUICommand) Run(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
) error {
	// Initialize settings from parsed layers
	s := &UpdateUISettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return err
	}

	// Read the YAML file
	yamlData, err := os.ReadFile(s.File)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse YAML to map
	var data map[string]interface{}
	if err := yaml.Unmarshal(yamlData, &data); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Convert to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to convert to JSON: %w", err)
	}

	// Send to API
	url := fmt.Sprintf("http://%s:%d/api/ui-update", s.Host, s.Port)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			if err == nil {
				err = fmt.Errorf("failed to close response body: %w", closeErr)
			}
		}
	}()

	// Check response
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned error: %s - %s", resp.Status, string(body))
	}

	fmt.Println("UI updated successfully")
	return nil
}
