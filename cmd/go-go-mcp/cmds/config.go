package cmds

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-go-golems/go-go-mcp/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type Profile struct {
	Description string        `yaml:"description"`
	Tools       ToolSources   `yaml:"tools"`
	Prompts     PromptSources `yaml:"prompts"`
}

type ToolSources struct {
	Directories []DirectorySource `yaml:"directories,omitempty"`
	Files       []FileSource      `yaml:"files,omitempty"`
}

type PromptSources struct {
	Directories []DirectorySource `yaml:"directories,omitempty"`
	Files       []FileSource      `yaml:"files,omitempty"`
}

type DirectorySource struct {
	Path      string                            `yaml:"path"`
	Defaults  map[string]map[string]interface{} `yaml:"defaults,omitempty"`
	Overrides map[string]map[string]interface{} `yaml:"overrides,omitempty"`
	Whitelist map[string][]string               `yaml:"whitelist,omitempty"`
	Blacklist map[string][]string               `yaml:"blacklist,omitempty"`
}

type FileSource struct {
	Path      string                            `yaml:"path"`
	Defaults  map[string]map[string]interface{} `yaml:"defaults,omitempty"`
	Overrides map[string]map[string]interface{} `yaml:"overrides,omitempty"`
}

type Config struct {
	Version        string             `yaml:"version"`
	DefaultProfile string             `yaml:"defaultProfile"`
	Profiles       map[string]Profile `yaml:"profiles"`
}

func NewConfigGroupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage go-go-mcp configuration",
		Long:  `Commands for managing go-go-mcp configuration files and profiles.`,
	}

	cmd.AddCommand(NewConfigInitCommand())
	cmd.AddCommand(NewConfigEditCommand())
	cmd.AddCommand(NewConfigListProfilesCommand())
	cmd.AddCommand(NewConfigShowProfileCommand())
	cmd.AddCommand(NewConfigAddToolCommand())
	cmd.AddCommand(NewConfigAddProfileCommand())
	cmd.AddCommand(NewConfigDuplicateProfileCommand())
	cmd.AddCommand(NewConfigSetDefaultProfileCommand())

	return cmd
}

func NewConfigInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a new configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			configFile := viper.ConfigFileUsed()
			configFile, err := config.GetProfilesPath(configFile)
			if err != nil {
				return fmt.Errorf("could not get profiles path: %w", err)
			}

			// Create directory if it doesn't exist
			configDir := filepath.Dir(configFile)
			if err := os.MkdirAll(configDir, 0755); err != nil {
				return fmt.Errorf("could not create config directory: %w", err)
			}

			// Check if file already exists
			if _, err := os.Stat(configFile); err == nil {
				return fmt.Errorf("config file already exists at %s", configFile)
			}

			// Create initial config with documentation
			initialContent := `# Go Go MCP Configuration File
#
# This file contains configuration for different profiles, tools, and prompts.
# Each profile can have its own set of tools and prompts with specific parameter settings.
#
# You can manage this file using the 'go-go-mcp config' commands:
# - edit: Edit this file in your default editor
# - list-profiles: List all available profiles
# - show-profile: Show full configuration of a profile
# - add-tool: Add tool directory or file to a profile
# - add-profile: Create a new profile
# - duplicate-profile: Create a new profile by duplicating an existing one
# - set-default-profile: Set the default profile

version: "1"
defaultProfile: development

profiles:
  development:
    description: "Development environment with debug tools"
    tools:
      directories:
        - path: ./dev-tools
          defaults:
            default:
              debug: true
              verbose: true
          blacklist:
            default:
              - api_key
      
      files:
        - path: ./special-debug.yaml
    
    prompts:
      directories:
        - path: ./dev-prompts

  production:
    description: "Production environment with strict controls"
    tools:
      directories:
        - path: /opt/go-go-mcp/tools
          whitelist:
            default:
              - timeout
              - retries
          overrides:
            default:
              timeout: 5s
              retries: 3
      
      files:
        - path: /opt/go-go-mcp/tools/special.yaml
    
    prompts:
      directories:
        - path: /opt/go-go-mcp/prompts
          overrides:
            default:
              temperature: 0.7
              max_tokens: 1000

# Parameter Management:
#
# Each tool can have multiple parameter layers, and each layer can have its own settings:
# - defaults: Set default values for parameters
# - overrides: Force specific parameter values
# - whitelist: Only allow specific parameters
# - blacklist: Prevent certain parameters from being used
#
# Common layer names:
# - default: Main parameter layer used by most tools
# - output: Parameters related to output formatting
# - format: Parameters controlling data format options
#
# For more information, see the documentation:
# go-go-mcp help config-file
`

			if err := os.WriteFile(configFile, []byte(initialContent), 0644); err != nil {
				return fmt.Errorf("could not write config file: %w", err)
			}

			fmt.Printf("Created new config file at %s\n", configFile)
			return nil
		},
	}
}

func NewConfigEditCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Edit the configuration file in your default editor",
		RunE: func(cmd *cobra.Command, args []string) error {
			configFile := viper.ConfigFileUsed()
			configFile, err := config.GetProfilesPath(configFile)
			if err != nil {
				return fmt.Errorf("could not get profiles path: %w", err)
			}

			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = "vim"
			}

			editCmd := exec.Command(editor, configFile)
			editCmd.Stdin = os.Stdin
			editCmd.Stdout = os.Stdout
			editCmd.Stderr = os.Stderr

			return editCmd.Run()
		},
	}
}

func NewConfigListProfilesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list-profiles",
		Short: "List all available profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			configFile := viper.ConfigFileUsed()
			configFile, err := config.GetProfilesPath(configFile)
			if err != nil {
				return fmt.Errorf("could not get profiles path: %w", err)
			}

			editor, err := config.NewConfigEditor(configFile)
			if err != nil {
				return fmt.Errorf("could not create config editor: %w", err)
			}

			defaultProfile, err := editor.GetDefaultProfile()
			if err != nil {
				return fmt.Errorf("could not get default profile: %w", err)
			}

			profiles, err := editor.GetProfiles()
			if err != nil {
				return fmt.Errorf("could not get profiles: %w", err)
			}

			fmt.Printf("Default profile: %s\n\n", defaultProfile)
			fmt.Println("Available profiles:")
			for name, desc := range profiles {
				fmt.Printf("- %s: %s\n", name, desc)
			}

			return nil
		},
	}
}

func NewConfigShowProfileCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show-profile [profile-name]",
		Short: "Show the full configuration of a profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configFile := viper.ConfigFileUsed()
			configFile, err := config.GetProfilesPath(configFile)
			if err != nil {
				return fmt.Errorf("could not get profiles path: %w", err)
			}

			editor, err := config.NewConfigEditor(configFile)
			if err != nil {
				return fmt.Errorf("could not create config editor: %w", err)
			}

			profile, err := editor.GetProfile(args[0])
			if err != nil {
				return fmt.Errorf("could not get profile: %w", err)
			}

			data, err := yaml.Marshal(profile)
			if err != nil {
				return fmt.Errorf("could not marshal profile: %w", err)
			}

			fmt.Printf("Profile: %s\n\n%s", args[0], string(data))
			return nil
		},
	}
}

func NewConfigAddToolCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-tool [profile-name] [--dir path | --file path]",
		Short: "Add a tool directory or file to a profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, _ := cmd.Flags().GetString("dir")
			file, _ := cmd.Flags().GetString("file")

			if (dir == "" && file == "") || (dir != "" && file != "") {
				return fmt.Errorf("exactly one of --dir or --file must be specified")
			}

			configFile := viper.ConfigFileUsed()
			configFile, err := config.GetProfilesPath(configFile)
			if err != nil {
				return fmt.Errorf("could not get profiles path: %w", err)
			}

			editor, err := config.NewConfigEditor(configFile)
			if err != nil {
				return fmt.Errorf("could not create config editor: %w", err)
			}

			if dir != "" {
				err = editor.AddToolDirectory(args[0], dir, map[string]interface{}{})
				if err != nil {
					return fmt.Errorf("could not add tool directory: %w", err)
				}
			} else {
				err = editor.AddToolFile(args[0], file)
				if err != nil {
					return fmt.Errorf("could not add tool file: %w", err)
				}
			}

			if err := editor.Save(); err != nil {
				return fmt.Errorf("could not save config file: %w", err)
			}

			fmt.Printf("Added tool to profile %s\n", args[0])
			return nil
		},
	}

	cmd.Flags().String("dir", "", "Directory path to add")
	cmd.Flags().String("file", "", "File path to add")

	return cmd
}

func NewConfigAddProfileCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "add-profile [profile-name] [description]",
		Short: "Add a new profile",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			configFile := viper.ConfigFileUsed()
			configFile, err := config.GetProfilesPath(configFile)
			if err != nil {
				return fmt.Errorf("could not get profiles path: %w", err)
			}

			editor, err := config.NewConfigEditor(configFile)
			if err != nil {
				return fmt.Errorf("could not create config editor: %w", err)
			}

			err = editor.AddProfile(args[0], args[1])
			if err != nil {
				return fmt.Errorf("could not add profile: %w", err)
			}

			if err := editor.Save(); err != nil {
				return fmt.Errorf("could not save config file: %w", err)
			}

			fmt.Printf("Added new profile: %s\n", args[0])
			return nil
		},
	}
}

func NewConfigDuplicateProfileCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "duplicate-profile [source-profile] [new-profile] [description]",
		Short: "Duplicate an existing profile with a new name",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			configFile := viper.ConfigFileUsed()
			configFile, err := config.GetProfilesPath(configFile)
			if err != nil {
				return fmt.Errorf("could not get profiles path: %w", err)
			}

			editor, err := config.NewConfigEditor(configFile)
			if err != nil {
				return fmt.Errorf("could not create config editor: %w", err)
			}

			err = editor.DuplicateProfile(args[0], args[1], args[2])
			if err != nil {
				return fmt.Errorf("could not duplicate profile: %w", err)
			}

			if err := editor.Save(); err != nil {
				return fmt.Errorf("could not save config file: %w", err)
			}

			fmt.Printf("Duplicated profile %s to %s\n", args[0], args[1])
			return nil
		},
	}
}

func NewConfigSetDefaultProfileCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set-default-profile [profile-name]",
		Short: "Set the default profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configFile := viper.ConfigFileUsed()
			configFile, err := config.GetProfilesPath(configFile)
			if err != nil {
				return fmt.Errorf("could not get profiles path: %w", err)
			}

			editor, err := config.NewConfigEditor(configFile)
			if err != nil {
				return fmt.Errorf("could not create config editor: %w", err)
			}

			err = editor.SetDefaultProfile(args[0])
			if err != nil {
				return fmt.Errorf("could not set default profile: %w", err)
			}

			if err := editor.Save(); err != nil {
				return fmt.Errorf("could not save config file: %w", err)
			}

			fmt.Printf("Set default profile to %s\n", args[0])
			return nil
		},
	}
}
