package cmds

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"text/template"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

// ShellCommandDescription represents the YAML structure for shell commands
type ShellCommandDescription struct {
	Name          string                            `yaml:"name"`
	Short         string                            `yaml:"short"`
	Long          string                            `yaml:"long,omitempty"`
	Flags         []*parameters.ParameterDefinition `yaml:"flags,omitempty"`
	Arguments     []*parameters.ParameterDefinition `yaml:"arguments,omitempty"`
	Layers        []layers.ParameterLayer           `yaml:"layers,omitempty"`
	ShellScript   string                            `yaml:"shell-script,omitempty"`
	Command       []string                          `yaml:"command,omitempty"`
	Cwd           string                            `yaml:"cwd,omitempty"`
	Environment   map[string]string                 `yaml:"environment,omitempty"`
	CaptureStderr bool                              `yaml:"capture-stderr,omitempty"`
}

// ShellCommand is the runtime representation of a shell command
type ShellCommand struct {
	*cmds.CommandDescription
	ShellScript   string
	Command       []string
	Cwd           string
	Environment   map[string]string
	CaptureStderr bool
}

var _ cmds.WriterCommand = &ShellCommand{}

type ShellCommandOption func(*ShellCommand)

func WithShellScript(script string) ShellCommandOption {
	return func(c *ShellCommand) {
		c.ShellScript = script
	}
}

func WithCommand(cmd []string) ShellCommandOption {
	return func(c *ShellCommand) {
		c.Command = cmd
	}
}

func WithCwd(cwd string) ShellCommandOption {
	return func(c *ShellCommand) {
		c.Cwd = cwd
	}
}

func WithEnvironment(env map[string]string) ShellCommandOption {
	return func(c *ShellCommand) {
		c.Environment = env
	}
}

func WithCaptureStderr(capture bool) ShellCommandOption {
	return func(c *ShellCommand) {
		c.CaptureStderr = capture
	}
}

// NewShellCommand creates a new ShellCommand with the given options
func NewShellCommand(
	description *cmds.CommandDescription,
	options ...ShellCommandOption,
) (*ShellCommand, error) {
	ret := &ShellCommand{
		CommandDescription: description,
		Environment:        make(map[string]string),
	}

	for _, option := range options {
		option(ret)
	}

	if ret.ShellScript == "" && len(ret.Command) == 0 {
		return nil, fmt.Errorf("either shell script or command must be specified")
	}

	return ret, nil
}

type templateData struct {
	Args map[string]interface{}
	Env  map[string]string
}

// processTemplate handles template processing for both command arguments and environment variables
func (c *ShellCommand) processTemplate(
	templateStr string,
	args map[string]interface{},
) (string, error) {
	data := templateData{
		Args: args,
		Env:  c.Environment,
	}

	tmpl, err := template.New("shell").Parse(templateStr)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse template")
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", errors.Wrap(err, "failed to execute template")
	}

	return buf.String(), nil
}

// ExecuteCommand handles the actual command execution
func (c *ShellCommand) ExecuteCommand(
	ctx context.Context,
	args map[string]interface{},
	w io.Writer,
) error {
	var cmd *exec.Cmd

	if c.ShellScript != "" {
		// Process script template
		script, err := c.processTemplate(c.ShellScript, args)
		if err != nil {
			return errors.Wrap(err, "failed to process shell script template")
		}

		fmt.Fprintln(os.Stderr, script)
		// Debug log the processed script
		log.Debug().Str("script", script).Msg("executing shell script")

		// Create temporary script file
		tmpFile, err := os.CreateTemp("", "shell-*.sh")
		if err != nil {
			return errors.Wrap(err, "failed to create temporary script file")
		}
		defer os.Remove(tmpFile.Name())

		if _, err := tmpFile.WriteString(script); err != nil {
			return errors.Wrap(err, "failed to write script to temporary file")
		}
		if err := tmpFile.Close(); err != nil {
			return errors.Wrap(err, "failed to close temporary file")
		}

		// Make the script executable
		if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
			return errors.Wrap(err, "failed to make script executable")
		}

		cmd = exec.CommandContext(ctx, "bash", tmpFile.Name())
	} else {
		// Process command template
		processedArgs := make([]string, len(c.Command))
		for i, arg := range c.Command {
			processed, err := c.processTemplate(arg, args)
			if err != nil {
				return errors.Wrapf(err, "failed to process command argument template: %s", arg)
			}
			processedArgs[i] = processed
		}

		cmd = exec.CommandContext(ctx, processedArgs[0], processedArgs[1:]...)
	}

	// Setup working directory
	if c.Cwd != "" {
		cmd.Dir = c.Cwd
	}

	// Process and set environment variables
	if len(c.Environment) > 0 {
		env := os.Environ()
		for k, v := range c.Environment {
			processed, err := c.processTemplate(v, args)
			if err != nil {
				return errors.Wrapf(err, "failed to process environment variable template: %s", k)
			}
			env = append(env, fmt.Sprintf("%s=%s", k, processed))
		}
		cmd.Env = env
	}

	// Setup output streams
	cmd.Stdout = w
	if c.CaptureStderr {
		cmd.Stderr = w
	} else {
		cmd.Stderr = os.Stderr
	}

	return cmd.Run()
}

// RunIntoWriter implements the WriterCommand interface
func (c *ShellCommand) RunIntoWriter(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
	w io.Writer,
) error {
	// Get arguments from parsed layers
	args := parsedLayers.GetDefaultParameterLayer().Parameters.ToMap()

	return c.ExecuteCommand(ctx, args, w)
}

// LoadShellCommandFromYAML creates a new ShellCommand from YAML data
func LoadShellCommandFromYAML(data []byte) (*ShellCommand, error) {
	var desc ShellCommandDescription
	if err := yaml.Unmarshal(data, &desc); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal YAML")
	}

	cmdDesc := cmds.NewCommandDescription(
		desc.Name,
		cmds.WithShort(desc.Short),
		cmds.WithLong(desc.Long),
		cmds.WithFlags(desc.Flags...),
		cmds.WithArguments(desc.Arguments...),
		cmds.WithLayersList(desc.Layers...),
	)

	return NewShellCommand(
		cmdDesc,
		WithShellScript(desc.ShellScript),
		WithCommand(desc.Command),
		WithCwd(desc.Cwd),
		WithEnvironment(desc.Environment),
		WithCaptureStderr(desc.CaptureStderr),
	)
}

// JsonSchemaProperty represents a property in the JSON Schema
type JsonSchemaProperty struct {
	Type                 string                         `json:"type"`
	Description          string                         `json:"description,omitempty"`
	Enum                 []string                       `json:"enum,omitempty"`
	Default              interface{}                    `json:"default,omitempty"`
	Items                *JsonSchemaProperty            `json:"items,omitempty"`
	Required             bool                           `json:"-"`
	Properties           map[string]*JsonSchemaProperty `json:"properties,omitempty"`
	AdditionalProperties *JsonSchemaProperty            `json:"additionalProperties,omitempty"`
}

// CommandJsonSchema represents the root JSON Schema for a command
type CommandJsonSchema struct {
	Type        string                         `json:"type"`
	Description string                         `json:"description,omitempty"`
	Properties  map[string]*JsonSchemaProperty `json:"properties"`
	Required    []string                       `json:"required,omitempty"`
}

// parameterTypeToJsonSchema converts a parameter definition to a JSON schema property
func parameterTypeToJsonSchema(param *parameters.ParameterDefinition) (*JsonSchemaProperty, error) {
	prop := &JsonSchemaProperty{
		Description: param.Help,
		Required:    param.Required,
	}

	if param.Default != nil {
		prop.Default = *param.Default
	}

	switch param.Type {
	// Basic types
	case parameters.ParameterTypeString:
		prop.Type = "string"

	case parameters.ParameterTypeInteger:
		prop.Type = "integer"

	case parameters.ParameterTypeFloat:
		prop.Type = "number"

	case parameters.ParameterTypeBool:
		prop.Type = "boolean"

	case parameters.ParameterTypeDate:
		prop.Type = "string"
		// Add format for date strings
		prop.Properties = map[string]*JsonSchemaProperty{
			"format": {Type: "string", Default: "date"},
		}

	// List types
	case parameters.ParameterTypeStringList:
		prop.Type = "array"
		prop.Items = &JsonSchemaProperty{Type: "string"}

	case parameters.ParameterTypeIntegerList:
		prop.Type = "array"
		prop.Items = &JsonSchemaProperty{Type: "integer"}

	case parameters.ParameterTypeFloatList:
		prop.Type = "array"
		prop.Items = &JsonSchemaProperty{Type: "number"}

	// Choice types
	case parameters.ParameterTypeChoice:
		prop.Type = "string"
		prop.Enum = param.Choices

	case parameters.ParameterTypeChoiceList:
		prop.Type = "array"
		prop.Items = &JsonSchemaProperty{Type: "string"}
		prop.Items.Enum = param.Choices

	// File types
	case parameters.ParameterTypeFile:
		prop.Type = "object"
		prop.Properties = map[string]*JsonSchemaProperty{
			"path":    {Type: "string", Description: "Path to the file"},
			"content": {Type: "string", Description: "File content"},
		}

	case parameters.ParameterTypeFileList:
		prop.Type = "array"
		prop.Items = &JsonSchemaProperty{
			Type: "object",
			Properties: map[string]*JsonSchemaProperty{
				"path":    {Type: "string", Description: "Path to the file"},
				"content": {Type: "string", Description: "File content"},
			},
		}

	// Key-value type
	case parameters.ParameterTypeKeyValue:
		prop.Type = "object"
		prop.Properties = map[string]*JsonSchemaProperty{
			"key":   {Type: "string"},
			"value": {Type: "string"},
		}

	// File-based parameter types
	case parameters.ParameterTypeStringFromFile:
		prop.Type = "string"

	case parameters.ParameterTypeStringFromFiles:
		prop.Type = "array"
		prop.Items = &JsonSchemaProperty{Type: "string"}

	case parameters.ParameterTypeObjectFromFile:
		prop.Type = "object"
		prop.AdditionalProperties = &JsonSchemaProperty{Type: "string"}

	case parameters.ParameterTypeObjectListFromFile:
		prop.Type = "array"
		prop.Items = &JsonSchemaProperty{
			Type:                 "object",
			AdditionalProperties: &JsonSchemaProperty{Type: "string"},
		}

	case parameters.ParameterTypeObjectListFromFiles:
		prop.Type = "array"
		prop.Items = &JsonSchemaProperty{
			Type:                 "object",
			AdditionalProperties: &JsonSchemaProperty{Type: "string"},
		}

	case parameters.ParameterTypeStringListFromFile:
		prop.Type = "array"
		prop.Items = &JsonSchemaProperty{Type: "string"}

	case parameters.ParameterTypeStringListFromFiles:
		prop.Type = "array"
		prop.Items = &JsonSchemaProperty{Type: "string"}

	default:
		return nil, fmt.Errorf("unsupported parameter type: %s", param.Type)
	}

	return prop, nil
}

// ToJsonSchema converts a ShellCommand to a JSON Schema representation
func ToJsonSchema(desc *cmds.CommandDescription) (*CommandJsonSchema, error) {
	schema := &CommandJsonSchema{
		Type:        "object",
		Description: fmt.Sprintf("%s\n\n%s", desc.Short, desc.Long),
		Properties:  make(map[string]*JsonSchemaProperty),
		Required:    []string{},
	}

	// Process flags
	err := desc.GetDefaultFlags().ForEachE(func(flag *parameters.ParameterDefinition) error {
		prop, err := parameterTypeToJsonSchema(flag)
		if err != nil {
			return fmt.Errorf("error processing flag %s: %w", flag.Name, err)
		}
		schema.Properties[flag.Name] = prop
		if flag.Required {
			schema.Required = append(schema.Required, flag.Name)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Process arguments
	err = desc.GetDefaultArguments().ForEachE(func(arg *parameters.ParameterDefinition) error {
		prop, err := parameterTypeToJsonSchema(arg)
		if err != nil {
			return fmt.Errorf("error processing argument %s: %w", arg.Name, err)
		}
		schema.Properties[arg.Name] = prop
		if arg.Required {
			schema.Required = append(schema.Required, arg.Name)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return schema, nil
}
