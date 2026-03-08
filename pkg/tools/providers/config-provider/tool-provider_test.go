package config_provider

import (
	"context"
	"encoding/json"
	"io"
	"testing"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/go-go-mcp/pkg/config"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/stretchr/testify/require"
)

type probeFieldResult struct {
	Present bool   `json:"present"`
	Value   string `json:"value,omitempty"`
}

type configProbeCommand struct {
	*cmds.CommandDescription
	fieldNames []string
}

func newConfigProbeCommand(name string, fieldNames ...string) *configProbeCommand {
	return &configProbeCommand{
		CommandDescription: cmds.NewCommandDescription(
			name,
			cmds.WithFlags(
				fields.New(
					"message",
					fields.TypeString,
					fields.WithDefault("schema-message"),
				),
				fields.New(
					"format",
					fields.TypeString,
					fields.WithDefault("schema-format"),
				),
				fields.New(
					"tone",
					fields.TypeString,
					fields.WithDefault("schema-tone"),
				),
				fields.New(
					"hidden",
					fields.TypeString,
					fields.WithDefault("schema-hidden"),
				),
				fields.New(
					"excluded",
					fields.TypeString,
					fields.WithDefault("schema-excluded"),
				),
			),
		),
		fieldNames: fieldNames,
	}
}

func (c *configProbeCommand) RunIntoWriter(_ context.Context, parsedValues *values.Values, w io.Writer) error {
	payload := map[string]probeFieldResult{}
	for _, fieldName := range c.fieldNames {
		fv, ok := parsedValues.GetField(schema.DefaultSlug, fieldName)
		if !ok {
			payload[fieldName] = probeFieldResult{Present: false}
			continue
		}

		value, _ := fv.GetInterfaceValue()
		stringValue, _ := value.(string)
		payload[fieldName] = probeFieldResult{
			Present: true,
			Value:   stringValue,
		}
	}

	return json.NewEncoder(w).Encode(payload)
}

func TestExecuteCommandAppliesConfigPrecedence(t *testing.T) {
	cmd := newConfigProbeCommand("config-probe-precedence", "message", "format", "tone")
	provider := &ConfigToolProvider{
		toolConfigs: map[string]*config.SourceConfig{
			cmd.Description().Name: {
				Defaults: config.LayerParameters{
					schema.DefaultSlug: {
						"message": "config-message",
						"format":  "config-format",
						"tone":    "config-tone",
					},
				},
				Overrides: config.LayerParameters{
					schema.DefaultSlug: {
						"message": "forced-message",
					},
				},
			},
		},
	}

	result, err := provider.executeCommand(context.Background(), cmd, map[string]interface{}{
		"format": "user-format",
	})
	require.NoError(t, err)
	require.False(t, result.IsError)

	got := decodeProbeOutput(t, result)
	require.Equal(t, probeFieldResult{Present: true, Value: "forced-message"}, got["message"])
	require.Equal(t, probeFieldResult{Present: true, Value: "user-format"}, got["format"])
	require.Equal(t, probeFieldResult{Present: true, Value: "config-tone"}, got["tone"])
}

func TestExecuteCommandAppliesWhitelistAndBlacklist(t *testing.T) {
	cmd := newConfigProbeCommand("config-probe-filters", "message", "hidden", "excluded")
	provider := &ConfigToolProvider{
		toolConfigs: map[string]*config.SourceConfig{
			cmd.Description().Name: {
				Whitelist: config.ParameterFilter{
					schema.DefaultSlug: {"message", "hidden"},
				},
				Blacklist: config.ParameterFilter{
					schema.DefaultSlug: {"hidden"},
				},
			},
		},
	}

	result, err := provider.executeCommand(context.Background(), cmd, map[string]interface{}{
		"message":  "visible-message",
		"hidden":   "secret-message",
		"excluded": "should-not-survive",
	})
	require.NoError(t, err)
	require.False(t, result.IsError)

	got := decodeProbeOutput(t, result)
	require.Equal(t, probeFieldResult{Present: true, Value: "visible-message"}, got["message"])
	require.Equal(t, probeFieldResult{Present: false}, got["hidden"])
	require.Equal(t, probeFieldResult{Present: false}, got["excluded"])
}

func decodeProbeOutput(t *testing.T, result *protocol.ToolResult) map[string]probeFieldResult {
	t.Helper()

	require.NotNil(t, result)
	require.NotEmpty(t, result.Content)
	require.Equal(t, "text", result.Content[0].Type)

	ret := map[string]probeFieldResult{}
	err := json.Unmarshal([]byte(result.Content[0].Text), &ret)
	require.NoError(t, err)

	return ret
}
