package cmds

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/go-go-golems/go-go-mcp/pkg/config"
	"github.com/go-go-golems/go-go-mcp/pkg/helpers"
	"github.com/hpcloud/tail"
	"github.com/spf13/cobra"
)

func NewClaudeConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claude",
		Short: "Manage Claude desktop configuration",
		Long:  `Commands for managing the Claude desktop configuration file.`,
	}

	cmd.AddCommand(
		newClaudeConfigInitCommand(),
		newClaudeConfigEditCommand(),
		newClaudeConfigAddMCPServerCommand(),
		newClaudeConfigRemoveMCPServerCommand(),
		newClaudeConfigListServersCommand(),
		newClaudeConfigEnableServerCommand(),
		newClaudeConfigDisableServerCommand(),
		newClaudeConfigTailCommand(),
		newClaudeConfigAddGoGoServerCommand(),
	)

	return cmd
}

func newClaudeConfigInitCommand() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize Claude desktop configuration",
		Long:  `Creates a new Claude desktop configuration file if it doesn't exist.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			editor, err := config.NewClaudeDesktopEditor(configPath)
			if err != nil {
				return err
			}

			return editor.Save()
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to config file (default: $XDG_CONFIG_HOME/Claude/claude_desktop_config.json)")

	return cmd
}

func newClaudeConfigEditCommand() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit Claude desktop configuration",
		Long:  `Opens the Claude desktop configuration file in your default editor.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if configPath == "" {
				var err error
				configPath, err = config.GetDefaultClaudeDesktopConfigPath()
				if err != nil {
					return err
				}
			}

			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = "vi"
			}

			c := exec.Command(editor, configPath)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			return c.Run()
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to config file (default: $XDG_CONFIG_HOME/Claude/claude_desktop_config.json)")

	return cmd
}

func newClaudeConfigAddMCPServerCommand() *cobra.Command {
	var configPath string
	var env []string
	var overwrite bool

	cmd := &cobra.Command{
		Use:   "add-mcp-server NAME COMMAND [ARGS...]",
		Short: "Add or update an MCP server",
		Long: `Adds a new MCP server configuration or updates an existing one.
		
If a server with the same name already exists, the command will fail unless --overwrite is specified.`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			editor, err := config.NewClaudeDesktopEditor(configPath)
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

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to config file (default: $XDG_CONFIG_HOME/Claude/claude_desktop_config.json)")
	cmd.Flags().StringArrayVarP(&env, "env", "e", []string{}, "Environment variables in KEY=VALUE format (can be specified multiple times)")
	cmd.Flags().BoolVarP(&overwrite, "overwrite", "w", false, "Overwrite existing server if it exists")

	return cmd
}

func newClaudeConfigRemoveMCPServerCommand() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "remove-mcp-server NAME",
		Short: "Remove an MCP server",
		Long:  `Removes an MCP server configuration.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			editor, err := config.NewClaudeDesktopEditor(configPath)
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

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to config file (default: $XDG_CONFIG_HOME/Claude/claude_desktop_config.json)")

	return cmd
}

func newClaudeConfigListServersCommand() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "list-servers",
		Short: "List configured MCP servers",
		Long:  `Lists all configured MCP servers and their settings.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			editor, err := config.NewClaudeDesktopEditor(configPath)
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
				disabled := ""
				if editor.IsServerDisabled(name) {
					disabled = " (disabled)"
				}
				fmt.Printf("%s%s:\n", name, disabled)
				fmt.Printf("  Command: %s\n", server.Command)
				if len(server.Args) > 0 {
					fmt.Printf("  Args: %v\n", server.Args)
				}
				if len(server.Env) > 0 {
					fmt.Printf("  Environment:\n")
					for k, v := range server.Env {
						fmt.Printf("    %s: %s\n", k, v)
					}
				}
				fmt.Println()
			}

			// List disabled servers
			disabled := editor.ListDisabledServers()
			if len(disabled) > 0 {
				fmt.Println("Disabled servers:")
				for _, name := range disabled {
					fmt.Printf("  - %s\n", name)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to config file (default: $XDG_CONFIG_HOME/Claude/claude_desktop_config.json)")

	return cmd
}

func newClaudeConfigEnableServerCommand() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "enable-server NAME",
		Short: "Enable a disabled MCP server",
		Long:  `Enables a previously disabled MCP server configuration.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			editor, err := config.NewClaudeDesktopEditor(configPath)
			if err != nil {
				return err
			}

			name := args[0]
			if err := editor.EnableMCPServer(name); err != nil {
				return err
			}

			if err := editor.Save(); err != nil {
				return err
			}

			fmt.Printf("Successfully enabled MCP server '%s'\n", name)
			fmt.Printf("Configuration saved to: %s\n", editor.GetConfigPath())

			return nil
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to config file (default: $XDG_CONFIG_HOME/Claude/claude_desktop_config.json)")

	return cmd
}

func newClaudeConfigDisableServerCommand() *cobra.Command {
	var configPath string

	cmd := &cobra.Command{
		Use:   "disable-server NAME",
		Short: "Disable an MCP server",
		Long:  `Disables an MCP server configuration without removing it.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			editor, err := config.NewClaudeDesktopEditor(configPath)
			if err != nil {
				return err
			}

			name := args[0]
			if err := editor.DisableMCPServer(name); err != nil {
				return err
			}

			if err := editor.Save(); err != nil {
				return err
			}

			fmt.Printf("Successfully disabled MCP server '%s'\n", name)
			fmt.Printf("Configuration saved to: %s\n", editor.GetConfigPath())

			return nil
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to config file (default: $XDG_CONFIG_HOME/Claude/claude_desktop_config.json)")

	return cmd
}

func newClaudeConfigTailCommand() *cobra.Command {
	var all bool
	var lines int

	cmd := &cobra.Command{
		Use:   "tail [server-names...]",
		Short: "Tail Claude log files",
		Long: `Tail the Claude log files in real-time.
Without arguments, tails the main mcp.log file.
With server names, tails the corresponding mcp-server-XXX.log files.
Use --all to tail all log files.
Use --lines/-n to output the last N lines of each file before tailing.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			xdgConfigPath, err := os.UserConfigDir()
			if err != nil {
				return fmt.Errorf("could not get user config directory: %w", err)
			}
			logDir := filepath.Join(xdgConfigPath, "Claude", "logs")

			// Create a context that can be cancelled
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Set up signal handling for graceful shutdown
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
			go func() {
				<-sigChan
				cancel()
			}()

			// Determine which files to tail
			var filesToTail []string
			if all {
				// Find all log files
				entries, err := os.ReadDir(logDir)
				if err != nil {
					return fmt.Errorf("could not read log directory: %w", err)
				}
				for _, entry := range entries {
					if !entry.IsDir() && strings.HasPrefix(entry.Name(), "mcp") && strings.HasSuffix(entry.Name(), ".log") {
						filesToTail = append(filesToTail, filepath.Join(logDir, entry.Name()))
					}
				}
			} else if len(args) == 0 {
				// Only tail main log file
				filesToTail = append(filesToTail, filepath.Join(logDir, "mcp.log"))
			} else {
				// Tail specified server logs
				for _, serverName := range args {
					filesToTail = append(filesToTail, filepath.Join(logDir, fmt.Sprintf("mcp-server-%s.log", serverName)))
				}
			}

			// Create a WaitGroup to wait for all tailers to finish
			var wg sync.WaitGroup

			// Start tailing each file
			for _, file := range filesToTail {
				wg.Add(1)
				go func(filename string) {
					defer wg.Done()

					// Find the starting position based on requested lines
					startPos := int64(0)
					if lines > 0 {
						var err error
						startPos, err = helpers.FindStartPosForLastNLines(filename, lines)
						if err != nil && !os.IsNotExist(err) {
							fmt.Fprintf(os.Stderr, "Error finding start position in %s: %v\n", filename, err)
						}
					}

					t, err := tail.TailFile(filename, tail.Config{
						Follow: true,
						ReOpen: true,
						Logger: tail.DiscardingLogger,
						Location: &tail.SeekInfo{
							Offset: startPos,
							Whence: io.SeekStart,
						},
					})
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error tailing %s: %v\n", filename, err)
						return
					}
					defer t.Cleanup()

					// Print the filename as a header
					fmt.Printf("==> %s <==\n", filename)

					// Read lines until context is cancelled
					for {
						select {
						case line, ok := <-t.Lines:
							if !ok {
								return
							}
							if line.Err != nil {
								fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", filename, line.Err)
								continue
							}
							fmt.Printf("%s: %s\n", filepath.Base(filename), line.Text)
						case <-ctx.Done():
							return
						}
					}
				}(file)
			}

			// Wait for all tailers to finish
			wg.Wait()

			return nil
		},
	}

	cmd.Flags().BoolVarP(&all, "all", "a", false, "Tail all log files")
	cmd.Flags().IntVarP(&lines, "lines", "n", 10, "Output the last N lines of each file before tailing")

	return cmd
}

func newClaudeConfigAddGoGoServerCommand() *cobra.Command {
	var configPath string
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
			editor, err := config.NewClaudeDesktopEditor(configPath)
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

	cmd.Flags().StringVarP(&configPath, "config", "c", "", "Path to config file (default: $XDG_CONFIG_HOME/Claude/claude_desktop_config.json)")
	cmd.Flags().StringArrayVarP(&env, "env", "e", []string{}, "Environment variables in KEY=VALUE format (can be specified multiple times)")
	cmd.Flags().BoolVarP(&overwrite, "overwrite", "w", false, "Overwrite existing server if it exists")
	cmd.Flags().StringArrayVar(&additionalArgs, "args", []string{}, "Additional arguments to pass to the server command")

	return cmd
}
