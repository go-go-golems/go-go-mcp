package cmds

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/glazed/pkg/helpers/templating"
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
	Debug         bool                              `yaml:"debug,omitempty"`
	SaveScriptDir string                            `yaml:"save-script-dir,omitempty"`
}

// ShellCommand is the runtime representation of a shell command
type ShellCommand struct {
	*cmds.CommandDescription
	ShellScript   string
	Command       []string
	Cwd           string
	Environment   map[string]string
	CaptureStderr bool
	Debug         bool
	SaveScriptDir string
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

func WithSaveScriptDir(dir string) ShellCommandOption {
	return func(c *ShellCommand) {
		c.SaveScriptDir = dir
	}
}

func WithDebug(debug bool) ShellCommandOption {
	return func(c *ShellCommand) {
		c.Debug = debug
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

	tmpl := templating.CreateTemplate("shell")
	tmpl, err := tmpl.Parse(templateStr)
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

	// Create temp JSON file with args
	argsJSON, err := json.MarshalIndent(args, "", "  ")
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal args to JSON")
		return errors.Wrap(err, "failed to marshal args to JSON")
	}

	log.Debug().Str("args_json", string(argsJSON)).Msg("args JSON")
	log.Debug().Str("save-script-dir", c.SaveScriptDir).Msg("save-script-dir")

	// Write args to temp file
	argsTmpFile, err := os.CreateTemp("", "args-*.json")
	if err != nil {
		log.Error().Err(err).Msg("failed to create temporary args file")
		return errors.Wrap(err, "failed to create temporary args file")
	}
	defer func() {
		if removeErr := os.Remove(argsTmpFile.Name()); removeErr != nil {
			log.Warn().Err(removeErr).Str("file", argsTmpFile.Name()).Msg("Failed to remove temporary args file")
		}
		log.Info().Str("args_file", argsTmpFile.Name()).Msg("removed temporary args file")
	}()

	if _, err := argsTmpFile.Write(argsJSON); err != nil {
		log.Error().Str("file", argsTmpFile.Name()).Err(err).Msg("failed to write args to temporary file")
		return errors.Wrap(err, "failed to write args to temporary file")
	}
	if err := argsTmpFile.Close(); err != nil {
		log.Error().Str("file", argsTmpFile.Name()).Err(err).Msg("failed to close temporary args file")
		return errors.Wrap(err, "failed to close temporary args file")
	}

	log.Info().Str("args_file", argsTmpFile.Name()).Msg("created temporary args file")

	if c.SaveScriptDir != "" {
		scriptFile := fmt.Sprintf("%s/shell-%s.args.json", c.SaveScriptDir, time.Now().Format("20060102-150405"))
		if err := os.WriteFile(scriptFile, argsJSON, 0644); err != nil {
			log.Warn().Err(err).Str("args_file", scriptFile).Msg("failed to write args file")
		}
		log.Info().Str("args_file", scriptFile).Msg("wrote args file")
	}

	if c.ShellScript != "" {
		// Process script template
		script, err := c.processTemplate(c.ShellScript, args)
		if err != nil {
			log.Error().Err(err).Str("shell_script", c.ShellScript).Msg("failed to process shell script template")
			return errors.Wrap(err, "failed to process shell script template")
		}

		if c.Debug {
			// Debug log the processed script
			log.Debug().Str("script", script).Msg("executing shell script")
		}

		// Create temporary script file
		tmpFile, err := os.CreateTemp("", "shell-*.sh")
		if err != nil {
			log.Error().Err(err).Msg("failed to create temporary script file")
			return errors.Wrap(err, "failed to create temporary script file")
		}
		log.Info().Str("debug_file", tmpFile.Name()).Msg("created temporary script file")
		defer func() {
			if removeErr := os.Remove(tmpFile.Name()); removeErr != nil {
				log.Warn().Err(removeErr).Str("file", tmpFile.Name()).Msg("Failed to remove temporary file")
			}
			log.Info().Str("debug_file", tmpFile.Name()).Msg("removed temporary script file")
		}()

		if _, err := tmpFile.WriteString(script); err != nil {
			log.Error().Str("file", tmpFile.Name()).Err(err).Msg("failed to write script to temporary file")
			return errors.Wrap(err, "failed to write script to temporary file")
		}
		if err := tmpFile.Close(); err != nil {
			log.Error().Str("file", tmpFile.Name()).Err(err).Msg("failed to close temporary file")
			return errors.Wrap(err, "failed to close temporary file")
		}

		// Make the script executable
		if err := os.Chmod(tmpFile.Name(), 0755); /* #nosec G302 */ err != nil {
			log.Error().Str("file", tmpFile.Name()).Err(err).Msg("failed to make script executable")
			return errors.Wrap(err, "failed to make script executable")
		}

		if c.SaveScriptDir != "" {
			// Create SaveScriptDir if it doesn't exist
			if _, err := os.Stat(c.SaveScriptDir); os.IsNotExist(err) {
				log.Info().Str("dir", c.SaveScriptDir).Msg("creating save script directory")
				if err := os.MkdirAll(c.SaveScriptDir, 0755); err != nil {
					log.Warn().Err(err).Str("dir", c.SaveScriptDir).Msg("failed to create save script directory")
					return errors.Wrapf(err, "failed to create save script directory: %s", c.SaveScriptDir)
				}
			}

			log.Info().Str("dir", c.SaveScriptDir).Msg("saving script and args to debug directory")

			// Save both script and args to debug directory with timestamp
			timestamp := time.Now().Format("20060102-150405")
			scriptFile := fmt.Sprintf("%s/shell-%s.sh", c.SaveScriptDir, timestamp)
			argsFile := fmt.Sprintf("%s/shell-%s.args.json", c.SaveScriptDir, timestamp)

			if err := os.WriteFile(scriptFile, []byte(script), 0644); err != nil {
				log.Warn().Err(err).Str("script_file", scriptFile).Msg("failed to write debug script file")
			} else {
				log.Info().Str("script_file", scriptFile).Msg("wrote debug script file")
			}

			if err := os.WriteFile(argsFile, argsJSON, 0644); err != nil {
				log.Warn().Err(err).Str("args_file", argsFile).Msg("failed to write debug args file")
			} else {
				log.Info().Str("args_file", argsFile).Msg("wrote debug args file")
			}
		}

		cmd = exec.CommandContext(ctx, "/bin/bash", tmpFile.Name())
		log.Debug().Str("command", cmd.String()).Msg("command")
	} else {
		// Process command template
		processedArgs := make([]string, len(c.Command))
		for i, arg := range c.Command {
			processed, err := c.processTemplate(arg, args)
			if err != nil {
				log.Error().Err(err).Str("command_argument", arg).Msg("failed to process command argument template")
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
	env := os.Environ()
	// Add the arguments JSON file path to environment
	env = append(env, fmt.Sprintf("MCP_ARGUMENTS_JSON_PATH=%s", argsTmpFile.Name()))
	log.Debug().Str("arguments_json_path", argsTmpFile.Name()).Msg("added arguments JSON path to environment")

	if len(c.Environment) > 0 {
		for k, v := range c.Environment {
			processed, err := c.processTemplate(v, args)
			if err != nil {
				log.Error().Err(err).Str("environment_variable", k).Msg("failed to process environment variable template")
				return errors.Wrapf(err, "failed to process environment variable template: %s", k)
			}
			env = append(env, fmt.Sprintf("%s=%s", k, processed))
		}
	}
	cmd.Env = env

	// Setup output streams
	cmd.Stdout = w
	if c.CaptureStderr {
		log.Debug().Msg("capturing stderr")
		cmd.Stderr = w
	} else {
		log.Debug().Msg("not capturing stderr")
		cmd.Stderr = os.Stderr
	}

	log.Info().Str("command", fmt.Sprintf("%v", cmd.Args)).Msg("executing command")

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

	log.Debug().Interface("args", args).Msg("executing command")
	err := c.ExecuteCommand(ctx, args, w)
	if err != nil {
		log.Error().Interface("args", args).Err(err).Msg("failed to execute command")
		return errors.Wrap(err, "failed to execute command")
	}

	return nil
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
		WithSaveScriptDir(desc.SaveScriptDir),
		WithDebug(desc.Debug),
	)
}
