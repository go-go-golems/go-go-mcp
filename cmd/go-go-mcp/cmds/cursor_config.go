package cmds

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/go-go-golems/go-go-mcp/pkg/config"
	"github.com/spf13/cobra"
)

func NewCursorConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cursor",
		Short: "Manage Cursor MCP configuration",
		Long:  `Commands for managing the Cursor MCP configuration file (global and project-specific).`,
	}

	cmd.AddCommand(
		newCursorConfigInitCommand(),
		newCursorConfigEditCommand(),
		newCursorConfigAddMCPServerCommand(),
		newCursorConfigAddMCPServerSSECommand(),
		newCursorConfigRemoveMCPServerCommand(),
		newCursorConfigListServersCommand(),
		newCursorConfigAddGoGoServerCommand(),
	)

	return cmd
}

func newCursorConfigInitCommand() *cobra.Command {
	var configPath string
	var projectDir string
	var global bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize Cursor MCP configuration",
		Long:  `Creates a new Cursor MCP configuration file if it doesn't exist.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if configPath == "" {
				var err error
				if global {
					configPath, err = config.GetGlobalCursorMCPConfigPath()
					if err != nil {
						return err
					}
				} else {
					if projectDir == "" {
						var err error
						projectDir, err = os.Getwd()
						if err != nil {
							return fmt.Errorf("could not get current directory: %w", err)
						}
					}
					configPath = config.GetProjectCursorMCPConfigPath(projectDir)
				}
			}

			editor, err := config.NewCursorMCPEditor(configPath)
			if err != nil {
				return err
			}

			if err := editor.Save(); err != nil {
				return err
			}

			fmt.Printf("Initialized Cursor MCP configuration at: %s\n", configPath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to config file")
	cmd.Flags().StringVarP(&projectDir, "project-dir", "p", "", "Project directory (defaults to current directory)")
	cmd.Flags().BoolVarP(&global, "global", "g", true, "Use global configuration (~/.cursor/mcp.json)")

	return cmd
}

func newCursorConfigEditCommand() *cobra.Command {
	var configPath string
	var projectDir string
	var global bool

	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit Cursor MCP configuration",
		Long:  `Opens the Cursor MCP configuration file in your default editor.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if configPath == "" {
				var err error
				if global {
					configPath, err = config.GetGlobalCursorMCPConfigPath()
					if err != nil {
						return err
					}
				} else {
					if projectDir == "" {
						var err error
						projectDir, err = os.Getwd()
						if err != nil {
							return fmt.Errorf("could not get current directory: %w", err)
						}
					}
					configPath = config.GetProjectCursorMCPConfigPath(projectDir)
				}
			}

			// Create the file if it doesn't exist
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				editor, err := config.NewCursorMCPEditor(configPath)
				if err != nil {
					return err
				}
				if err := editor.Save(); err != nil {
					return err
				}
			}

			editorCmd := os.Getenv("EDITOR")
			if editorCmd == "" {
				editorCmd = "vi"
			}

			c := exec.Command(editorCmd, configPath)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			return c.Run()
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to config file")
	cmd.Flags().StringVarP(&projectDir, "project-dir", "p", "", "Project directory (defaults to current directory)")
	cmd.Flags().BoolVarP(&global, "global", "g", true, "Use global configuration (~/.cursor/mcp.json)")

	return cmd
}

func newCursorConfigAddMCPServerCommand() *cobra.Command {
	var configPath string
	var projectDir string
	var global bool
	var env []string
	var overwrite bool

	cmd := &cobra.Command{
		Use:   "add-mcp-server NAME COMMAND [ARGS...]",
		Short: "Add or update an MCP server (stdio format)",
		Long: `Adds a new MCP server configuration or updates an existing one using the stdio format.
		
If a server with the same name already exists, the command will fail unless --overwrite is specified.`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if configPath == "" {
				var err error
				if global {
					configPath, err = config.GetGlobalCursorMCPConfigPath()
					if err != nil {
						return err
					}
				} else {
					if projectDir == "" {
						var err error
						projectDir, err = os.Getwd()
						if err != nil {
							return fmt.Errorf("could not get current directory: %w", err)
						}
					}
					configPath = config.GetProjectCursorMCPConfigPath(projectDir)
				}
			}

			editor, err := config.NewCursorMCPEditor(configPath)
			if err != nil {
				return err
			}

			name := args[0]
			command := args[1]
			cmdArgs := args[2:]

			// Parse environment variables
			envMap := make(map[string]string)
			for _, e := range env {
				parts := strings.SplitN(e, "=", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid environment variable format: %s (expected KEY=VALUE)", e)
				}
				envMap[parts[0]] = parts[1]
			}

			if err := editor.AddMCPServer(name, command, cmdArgs, envMap, overwrite); err != nil {
				return err
			}

			if err := editor.Save(); err != nil {
				return err
			}

			// Print success message with configuration details
			action := "Added"
			if overwrite {
				action = "Updated"
			}
			fmt.Printf("Successfully %s MCP server '%s':\n", action, name)
			fmt.Printf("  Command: %s\n", command)
			if len(cmdArgs) > 0 {
				fmt.Printf("  Args: %v\n", cmdArgs)
			}
			if len(envMap) > 0 {
				fmt.Printf("  Environment:\n")
				for k, v := range envMap {
					fmt.Printf("    %s: %s\n", k, v)
				}
			}
			fmt.Printf("\nConfiguration saved to: %s\n", editor.GetConfigPath())

			return nil
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to config file")
	cmd.Flags().StringVarP(&projectDir, "project-dir", "p", "", "Project directory (defaults to current directory)")
	cmd.Flags().BoolVarP(&global, "global", "g", true, "Use global configuration (~/.cursor/mcp.json)")
	cmd.Flags().StringArrayVarP(&env, "env", "e", []string{}, "Environment variables in KEY=VALUE format (can be specified multiple times)")
	cmd.Flags().BoolVarP(&overwrite, "overwrite", "w", false, "Overwrite existing server if it exists")

	return cmd
}

func newCursorConfigAddMCPServerSSECommand() *cobra.Command {
	var configPath string
	var projectDir string
	var global bool
	var env []string
	var overwrite bool

	cmd := &cobra.Command{
		Use:   "add-mcp-server-sse NAME URL",
		Short: "Add or update an MCP server (SSE format)",
		Long: `Adds a new MCP server configuration or updates an existing one using the SSE format.
		
If a server with the same name already exists, the command will fail unless --overwrite is specified.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if configPath == "" {
				var err error
				if global {
					configPath, err = config.GetGlobalCursorMCPConfigPath()
					if err != nil {
						return err
					}
				} else {
					if projectDir == "" {
						var err error
						projectDir, err = os.Getwd()
						if err != nil {
							return fmt.Errorf("could not get current directory: %w", err)
						}
					}
					configPath = config.GetProjectCursorMCPConfigPath(projectDir)
				}
			}

			editor, err := config.NewCursorMCPEditor(configPath)
			if err != nil {
				return err
			}

			name := args[0]
			url := args[1]

			// Parse environment variables
			envMap := make(map[string]string)
			for _, e := range env {
				parts := strings.SplitN(e, "=", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid environment variable format: %s (expected KEY=VALUE)", e)
				}
				envMap[parts[0]] = parts[1]
			}

			if err := editor.AddMCPServerSSE(name, url, envMap, overwrite); err != nil {
				return err
			}

			if err := editor.Save(); err != nil {
				return err
			}

			// Print success message with configuration details
			action := "Added"
			if overwrite {
				action = "Updated"
			}
			fmt.Printf("Successfully %s MCP server '%s':\n", action, name)
			fmt.Printf("  URL: %s\n", url)
			if len(envMap) > 0 {
				fmt.Printf("  Environment:\n")
				for k, v := range envMap {
					fmt.Printf("    %s: %s\n", k, v)
				}
			}
			fmt.Printf("\nConfiguration saved to: %s\n", editor.GetConfigPath())

			return nil
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to config file")
	cmd.Flags().StringVarP(&projectDir, "project-dir", "p", "", "Project directory (defaults to current directory)")
	cmd.Flags().BoolVarP(&global, "global", "g", true, "Use global configuration (~/.cursor/mcp.json)")
	cmd.Flags().StringArrayVarP(&env, "env", "e", []string{}, "Environment variables in KEY=VALUE format (can be specified multiple times)")
	cmd.Flags().BoolVarP(&overwrite, "overwrite", "w", false, "Overwrite existing server if it exists")

	return cmd
}

func newCursorConfigRemoveMCPServerCommand() *cobra.Command {
	var configPath string
	var projectDir string
	var global bool

	cmd := &cobra.Command{
		Use:   "remove-mcp-server NAME",
		Short: "Remove an MCP server",
		Long:  `Removes an MCP server configuration.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if configPath == "" {
				var err error
				if global {
					configPath, err = config.GetGlobalCursorMCPConfigPath()
					if err != nil {
						return err
					}
				} else {
					if projectDir == "" {
						var err error
						projectDir, err = os.Getwd()
						if err != nil {
							return fmt.Errorf("could not get current directory: %w", err)
						}
					}
					configPath = config.GetProjectCursorMCPConfigPath(projectDir)
				}
			}

			editor, err := config.NewCursorMCPEditor(configPath)
			if err != nil {
				return err
			}

			if err := editor.RemoveMCPServer(args[0]); err != nil {
				return err
			}

			if err := editor.Save(); err != nil {
				return err
			}

			fmt.Printf("Successfully removed MCP server '%s'\n", args[0])
			fmt.Printf("Configuration saved to: %s\n", editor.GetConfigPath())

			return nil
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to config file")
	cmd.Flags().StringVarP(&projectDir, "project-dir", "p", "", "Project directory (defaults to current directory)")
	cmd.Flags().BoolVarP(&global, "global", "g", true, "Use global configuration (~/.cursor/mcp.json)")

	return cmd
}

func newCursorConfigListServersCommand() *cobra.Command {
	var configPath string
	var projectDir string
	var global bool

	cmd := &cobra.Command{
		Use:   "list-servers",
		Short: "List configured MCP servers",
		Long:  `Lists all configured MCP servers and their settings.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if configPath == "" {
				var err error
				if global {
					configPath, err = config.GetGlobalCursorMCPConfigPath()
					if err != nil {
						return err
					}
				} else {
					if projectDir == "" {
						var err error
						projectDir, err = os.Getwd()
						if err != nil {
							return fmt.Errorf("could not get current directory: %w", err)
						}
					}
					configPath = config.GetProjectCursorMCPConfigPath(projectDir)
				}
			}

			// Check if the file exists
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				fmt.Printf("Configuration file does not exist: %s\n", configPath)
				return nil
			}

			editor, err := config.NewCursorMCPEditor(configPath)
			if err != nil {
				return err
			}

			servers := editor.ListServers()
			if len(servers) == 0 {
				fmt.Println("No MCP servers configured.")
				fmt.Printf("Configuration file: %s\n", editor.GetConfigPath())
				return nil
			}

			fmt.Printf("Configured MCP servers in %s:\n\n", editor.GetConfigPath())
			for name, server := range servers {
				fmt.Printf("%s:\n", name)
				if server.Command != "" {
					fmt.Printf("  Command: %s\n", server.Command)
					if len(server.Args) > 0 {
						fmt.Printf("  Args: %v\n", server.Args)
					}
				} else if server.URL != "" {
					fmt.Printf("  URL: %s\n", server.URL)
				}
				if len(server.Env) > 0 {
					fmt.Printf("  Environment:\n")
					for k, v := range server.Env {
						fmt.Printf("    %s: %s\n", k, v)
					}
				}
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to config file")
	cmd.Flags().StringVarP(&projectDir, "project-dir", "p", "", "Project directory (defaults to current directory)")
	cmd.Flags().BoolVarP(&global, "global", "g", true, "Use global configuration (~/.cursor/mcp.json)")

	return cmd
}

func newCursorConfigAddGoGoServerCommand() *cobra.Command {
	var configPath string
	var projectDir string
	var global bool
	var env []string
	var overwrite bool
	var additionalArgs []string

	cmd := &cobra.Command{
		Use:   "add-go-go-server NAME PROFILE [ARGS...]",
		Short: "Add an MCP server using go-go-mcp server",
		Long: `Adds a new MCP server configuration that uses go-go-mcp server with the specified profile.
		
This is a shortcut for adding a server with command "mcp server start --profile PROFILE" and additional args.
If a server with the same name already exists, the command will fail unless --overwrite is specified.`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if configPath == "" {
				var err error
				if global {
					configPath, err = config.GetGlobalCursorMCPConfigPath()
					if err != nil {
						return err
					}
				} else {
					if projectDir == "" {
						var err error
						projectDir, err = os.Getwd()
						if err != nil {
							return fmt.Errorf("could not get current directory: %w", err)
						}
					}
					configPath = config.GetProjectCursorMCPConfigPath(projectDir)
				}
			}

			editor, err := config.NewCursorMCPEditor(configPath)
			if err != nil {
				return err
			}

			name := args[0]
			profile := args[1]
			cmdArgs := append([]string{"server", "start", "--profile", profile}, args[2:]...)
			cmdArgs = append(cmdArgs, additionalArgs...)

			// Parse environment variables
			envMap := make(map[string]string)
			for _, e := range env {
				parts := strings.SplitN(e, "=", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid environment variable format: %s (expected KEY=VALUE)", e)
				}
				envMap[parts[0]] = parts[1]
			}

			// Find the path to the mcp binary
			mcpPath, err := exec.LookPath("mcp")
			if err != nil {
				return fmt.Errorf("could not find mcp executable in PATH: %w", err)
			}

			if err := editor.AddMCPServer(name, mcpPath, cmdArgs, envMap, overwrite); err != nil {
				return err
			}

			if err := editor.Save(); err != nil {
				return err
			}

			// Print success message with configuration details
			action := "Added"
			if overwrite {
				action = "Updated"
			}
			fmt.Printf("Successfully %s go-go-mcp server '%s':\n", action, name)
			fmt.Printf("  Command: %s\n", mcpPath)
			fmt.Printf("  Profile: %s\n", profile)
			if len(cmdArgs) > 4 {
				fmt.Printf("  Additional Args: %v\n", cmdArgs[4:])
			}
			if len(envMap) > 0 {
				fmt.Printf("  Environment:\n")
				for k, v := range envMap {
					fmt.Printf("    %s: %s\n", k, v)
				}
			}
			fmt.Printf("\nConfiguration saved to: %s\n", editor.GetConfigPath())

			return nil
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to config file")
	cmd.Flags().StringVarP(&projectDir, "project-dir", "p", "", "Project directory (defaults to current directory)")
	cmd.Flags().BoolVarP(&global, "global", "g", true, "Use global configuration (~/.cursor/mcp.json)")
	cmd.Flags().StringArrayVarP(&env, "env", "e", []string{}, "Environment variables in KEY=VALUE format (can be specified multiple times)")
	cmd.Flags().BoolVarP(&overwrite, "overwrite", "w", false, "Overwrite existing server if it exists")
	cmd.Flags().StringArrayVar(&additionalArgs, "args", []string{}, "Additional arguments to pass to the server command")

	return cmd
}
