package config_provider

import (
	"os"
	"path/filepath"

	"github.com/go-go-golems/clay/pkg/repositories"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/go-go-mcp/pkg"
	"github.com/go-go-golems/go-go-mcp/pkg/config"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/pkg/errors"
)

// ConfigPromptProvider implements pkg.PromptProvider interface
type ConfigPromptProvider struct {
	repository     *repositories.Repository
	pinocchioFiles map[string]*protocol.Prompt
	promptConfigs  map[string]*config.SourceConfig
}

func NewConfigPromptProvider(config_ *config.Config, profile string) (*ConfigPromptProvider, error) {
	if _, ok := config_.Profiles[profile]; !ok {
		return nil, errors.Errorf("profile %s not found", profile)
	}

	directories := []repositories.Directory{}

	profileConfig := config_.Profiles[profile]

	// Load directories using Clay's repository system
	for _, dir := range profileConfig.Prompts.Directories {
		absPath, err := filepath.Abs(dir.Path)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get absolute path for %s", dir.Path)
		}

		directories = append(directories, repositories.Directory{
			FS:            os.DirFS(absPath),
			RootDirectory: ".",
			Name:          dir.Path,
		})
	}

	provider := &ConfigPromptProvider{
		repository:     repositories.NewRepository(repositories.WithDirectories(directories...)),
		pinocchioFiles: make(map[string]*protocol.Prompt),
		promptConfigs:  make(map[string]*config.SourceConfig),
	}

	if profileConfig.Prompts == nil {
		return provider, nil
	}

	helpSystem := help.NewHelpSystem()
	// Load repository commands
	if err := provider.repository.LoadCommands(helpSystem); err != nil {
		return nil, errors.Wrap(err, "failed to load repository commands")
	}

	// Load individual Pinocchio files
	for _, file := range profileConfig.Prompts.Files {
		absPath, err := filepath.Abs(file.Path)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to get absolute path for %s", file.Path)
		}

		prompt, err := loadPinocchioFile(absPath)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to load pinocchio file from %s", file.Path)
		}

		provider.pinocchioFiles[prompt.Name] = prompt
		provider.promptConfigs[prompt.Name] = &file
	}

	return provider, nil
}

// ListPrompts implements pkg.PromptProvider interface
func (p *ConfigPromptProvider) ListPrompts(cursor string) ([]protocol.Prompt, string, error) {
	var prompts []protocol.Prompt

	// Get prompts from repositories
	repoCommands := p.repository.CollectCommands([]string{}, true)
	for _, cmd := range repoCommands {
		prompt := convertCommandToPrompt(cmd)
		prompts = append(prompts, prompt)
	}

	// Add Pinocchio files
	for _, prompt := range p.pinocchioFiles {
		prompts = append(prompts, *prompt)
	}

	// Handle cursor-based pagination if needed
	if cursor != "" {
		for i, prompt := range prompts {
			if prompt.Name == cursor && i+1 < len(prompts) {
				return prompts[i+1:], "", nil
			}
		}
		return nil, "", nil
	}

	return prompts, "", nil
}

// GetPrompt implements pkg.PromptProvider interface
func (p *ConfigPromptProvider) GetPrompt(name string, arguments map[string]string) (*protocol.PromptMessage, error) {
	// Try repositories first
	if cmd, ok := p.repository.GetCommand(name); ok {
		return p.executeRepositoryPrompt(cmd, arguments)
	}

	// Try Pinocchio files
	if prompt, ok := p.pinocchioFiles[name]; ok {
		return p.executePinocchioPrompt(prompt, arguments)
	}

	return nil, pkg.ErrPromptNotFound
}

func (p *ConfigPromptProvider) executeRepositoryPrompt(cmd cmds.Command, arguments map[string]string) (*protocol.PromptMessage, error) {
	// Convert arguments and apply parameter configuration
	if config, ok := p.promptConfigs[cmd.Description().Name]; ok {
		args := make(map[string]interface{})
		for k, v := range arguments {
			args[k] = v
		}
		_ = config
		for k, v := range args {
			if str, ok := v.(string); ok {
				arguments[k] = str
			}
		}
	}

	// Build message content from arguments
	content := "Please help me with the following task:\n\n"
	for key, value := range arguments {
		content += key + ": " + value + "\n"
	}

	return &protocol.PromptMessage{
		Role: "user",
		Content: protocol.PromptContent{
			Type: "text",
			Text: content,
		},
	}, nil
}

func (p *ConfigPromptProvider) executePinocchioPrompt(prompt *protocol.Prompt, arguments map[string]string) (*protocol.PromptMessage, error) {
	// Validate required arguments
	for _, arg := range prompt.Arguments {
		if arg.Required {
			if _, ok := arguments[arg.Name]; !ok {
				return nil, errors.Errorf("missing required argument: %s", arg.Name)
			}
		}
	}

	// Apply parameter configuration
	if config, ok := p.promptConfigs[prompt.Name]; ok {
		args := make(map[string]interface{})
		for k, v := range arguments {
			args[k] = v
		}
		_ = config
		for k, v := range args {
			if str, ok := v.(string); ok {
				arguments[k] = str
			}
		}
	}

	// Build message content from arguments
	// XXX this doesn't really make sense, but whatever, placeholder
	content := prompt.Description + "\n\n"
	for key, value := range arguments {
		content += key + ": " + value + "\n"
	}

	return &protocol.PromptMessage{
		Role: "user",
		Content: protocol.PromptContent{
			Type: "text",
			Text: content,
		},
	}, nil
}

func loadPinocchioFile(path string) (*protocol.Prompt, error) {
	// TODO: Implement Pinocchio file loading
	// For now, return a simple prompt
	return &protocol.Prompt{
		Name:        filepath.Base(path),
		Description: "Pinocchio prompt from " + path,
		Arguments: []protocol.PromptArgument{
			{
				Name:        "context",
				Description: "Additional context",
				Required:    false,
			},
		},
	}, nil
}

func convertCommandToPrompt(cmd cmds.Command) protocol.Prompt {
	prompt := protocol.Prompt{
		Name:        cmd.Description().Name,
		Description: cmd.Description().Short + "\n" + cmd.Description().Long,
		Arguments:   make([]protocol.PromptArgument, 0),
	}

	flagsAndArguments := cmd.Description().GetDefaultFlags().ToList()
	flagsAndArguments = append(flagsAndArguments, cmd.Description().GetDefaultArguments().ToList()...)

	// Convert parameters to arguments
	for _, param := range flagsAndArguments {
		arg := protocol.PromptArgument{
			Name:        param.Name,
			Description: param.Help,
			Required:    param.Required,
		}
		prompt.Arguments = append(prompt.Arguments, arg)
	}

	return prompt
}
