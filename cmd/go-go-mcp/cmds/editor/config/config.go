package config

import (
	"fmt"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/go-go-mcp/pkg/core/configstore"
	"github.com/go-go-golems/go-go-mcp/pkg/mcp/types"
	"github.com/spf13/cobra"
)

// ConfigCommandOptions holds common options for config commands
type ConfigCommandOptions struct {
	Editor  string
	Editors []string
	Target  string
}

// EditorInfo holds information about supported editors
type EditorInfo struct {
	Name        string
	ConfigType  configstore.ConfigType
	Description string
}

// SupportedEditors defines all supported editors and their config types
var SupportedEditors = []EditorInfo{
	{Name: "claude", ConfigType: configstore.ConfigTypeClaude, Description: "Claude desktop configuration"},
	{Name: "cursor", ConfigType: configstore.ConfigTypeCursor, Description: "Cursor editor configuration"},
	{Name: "ampcode", ConfigType: configstore.ConfigTypeAmpCode, Description: "AmpCode editor configuration"},
	{Name: "amp", ConfigType: configstore.ConfigTypeAmp, Description: "Amp editor configuration"},
	{Name: "crush", ConfigType: configstore.ConfigTypeCrushLocal, Description: "Crush editor configuration"},
}

// ValidateEditor checks if the provided editor is supported
func ValidateEditor(editor string) error {
	for _, info := range SupportedEditors {
		if info.Name == editor {
			return nil
		}
	}

	var validEditors []string
	for _, info := range SupportedEditors {
		validEditors = append(validEditors, info.Name)
	}

	return fmt.Errorf("invalid editor '%s'. Supported editors: %v", editor, validEditors)
}

// ParseEditors parses a comma-separated list of editors and validates them
func ParseEditors(editorList string) ([]string, error) {
	if editorList == "" {
		return nil, fmt.Errorf("editor argument required")
	}

	editors := strings.Split(editorList, ",")
	var cleanEditors []string

	for _, editor := range editors {
		editor = strings.TrimSpace(editor)
		if editor == "" {
			continue
		}

		if err := ValidateEditor(editor); err != nil {
			return nil, err
		}

		cleanEditors = append(cleanEditors, editor)
	}

	if len(cleanEditors) == 0 {
		return nil, fmt.Errorf("no valid editors found in list")
	}

	return cleanEditors, nil
}

// ValidateEditors checks if all provided editors are supported
func ValidateEditors(editors []string) error {
	for _, editor := range editors {
		if err := ValidateEditor(editor); err != nil {
			return err
		}
	}
	return nil
}

// GetConfigType returns the config type for a given editor
func GetConfigType(editor string) (configstore.ConfigType, error) {
	for _, info := range SupportedEditors {
		if info.Name == editor {
			return info.ConfigType, nil
		}
	}
	return "", fmt.Errorf("unknown editor: %s", editor)
}

// NewStoreWithTarget creates a new store with target resolution
func NewStoreWithTarget(editor, target string) (configstore.Store, error) {
	if err := ValidateEditor(editor); err != nil {
		return nil, err
	}

	configType, err := GetConfigType(editor)
	if err != nil {
		return nil, err
	}

	store, err := configstore.NewStore(configType)
	if err != nil {
		return nil, fmt.Errorf("failed to create store for %s: %w", editor, err)
	}

	if target != "" {
		if err := store.ResolveTarget(target); err != nil {
			supportedTargets := store.GetSupportedTargets()
			return nil, fmt.Errorf("invalid target '%s' for %s. Supported targets: %v", target, editor, supportedTargets)
		}
	}

	return store, nil
}

// EditorStore represents a store paired with its editor name
type EditorStore struct {
	Editor string
	Store  configstore.Store
}

// NewStoresWithTarget creates stores for multiple editors with target resolution
func NewStoresWithTarget(editors []string, target string) ([]EditorStore, error) {
	if err := ValidateEditors(editors); err != nil {
		return nil, err
	}

	var stores []EditorStore
	for _, editor := range editors {
		store, err := NewStoreWithTarget(editor, target)
		if err != nil {
			return nil, fmt.Errorf("failed to create store for %s: %w", editor, err)
		}
		stores = append(stores, EditorStore{
			Editor: editor,
			Store:  store,
		})
	}

	return stores, nil
}

// NewConfigCommand creates the main config command that delegates to editor-specific subcommands
func NewConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config <editor[,editor2,...]> <verb> [options]",
		Short: "Unified configuration management for MCP editors",
		Long: `Manage MCP server configurations across different editors.

Supported editors: claude, cursor, ampcode, amp, crush

You can specify multiple editors using comma-separated lists to perform 
operations across multiple editors simultaneously.

Available verbs:
  add          Add MCP server
  remove       Remove MCP server  
  list         List servers
  enable       Enable server
  disable      Disable server
  copy         Copy MCP server configuration
  edit         Edit config file
  init         Initialize config
  tail         Tail logs (Claude-specific)
  add-go-go    Add go-go-mcp server

Examples:
  # Single editor (existing functionality)
  mcp editor config claude add myserver /path/to/command
  mcp editor config cursor list --target global
  mcp editor config amp remove myserver

  # Multiple editors (new functionality)
  mcp editor config claude,cursor,amp add myserver /path/to/command
  mcp editor config claude,cursor list --target global
  mcp editor config cursor,amp remove myserver`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("editor argument required. Supported editors: %v", getSupportedEditorNames())
			}

			editor := args[0]
			if err := ValidateEditor(editor); err != nil {
				return err
			}

			if len(args) == 1 {
				return fmt.Errorf("verb argument required after editor '%s'", editor)
			}

			return fmt.Errorf("unknown verb '%s' for editor '%s'", args[1], editor)
		},
	}

	// Add subcommands for each editor (single-editor commands)
	for _, editorInfo := range SupportedEditors {
		editorCmd, err := createEditorCommand(editorInfo)
		if err != nil {
			panic(fmt.Sprintf("Failed to create editor command for %s: %v", editorInfo.Name, err))
		}
		cmd.AddCommand(editorCmd)
	}

	// Add multi-editor verb commands using Glazed dual commands
	
	// Multi-editor add command (dual mode)
	multiAddCmd, err := NewAddCommandDual()
	if err != nil {
		panic(fmt.Sprintf("Failed to create multi-editor add command: %v", err))
	}
	cobraMultiAddCmd, err := cli.BuildCobraCommand(multiAddCmd)
	if err != nil {
		panic(fmt.Sprintf("Failed to build multi-editor add command: %v", err))
	}
	// Override Use to match multi-editor format
	cobraMultiAddCmd.Use = "add EDITORS NAME COMMAND [ARGS...]"
	cmd.AddCommand(cobraMultiAddCmd)

	// Multi-editor list command (single mode Glazed)
	cmd.AddCommand(NewMultiListCommandGlazed())
	
	// Multi-editor remove command (dual mode)
	multiRemoveCmd, err := NewRemoveCommandDual()
	if err != nil {
		panic(fmt.Sprintf("Failed to create multi-editor remove command: %v", err))
	}
	cobraMultiRemoveCmd, err := cli.BuildCobraCommand(multiRemoveCmd)
	if err != nil {
		panic(fmt.Sprintf("Failed to build multi-editor remove command: %v", err))
	}
	// Override Use to match multi-editor format
	cobraMultiRemoveCmd.Use = "remove EDITORS NAME"
	cmd.AddCommand(cobraMultiRemoveCmd)
	
	// Multi-editor enable command (dual mode)
	multiEnableCmd, err := NewEnableCommandDual()
	if err != nil {
		panic(fmt.Sprintf("Failed to create multi-editor enable command: %v", err))
	}
	cobraMultiEnableCmd, err := cli.BuildCobraCommand(multiEnableCmd)
	if err != nil {
		panic(fmt.Sprintf("Failed to build multi-editor enable command: %v", err))
	}
	// Override Use to match multi-editor format
	cobraMultiEnableCmd.Use = "enable EDITORS NAME"
	cmd.AddCommand(cobraMultiEnableCmd)
	
	// Multi-editor disable command (dual mode)
	multiDisableCmd, err := NewDisableCommandDual()
	if err != nil {
		panic(fmt.Sprintf("Failed to create multi-editor disable command: %v", err))
	}
	cobraMultiDisableCmd, err := cli.BuildCobraCommand(multiDisableCmd)
	if err != nil {
		panic(fmt.Sprintf("Failed to build multi-editor disable command: %v", err))
	}
	// Override Use to match multi-editor format
	cobraMultiDisableCmd.Use = "disable EDITORS NAME"
	cmd.AddCommand(cobraMultiDisableCmd)
	
	// Multi-editor copy command (dual mode)
	multiCopyCmd, err := NewCopyCommandDual()
	if err != nil {
		panic(fmt.Sprintf("Failed to create multi-editor copy command: %v", err))
	}
	cobraMultiCopyCmd, err := cli.BuildCobraCommand(multiCopyCmd)
	if err != nil {
		panic(fmt.Sprintf("Failed to build multi-editor copy command: %v", err))
	}
	// Override Use to match multi-editor format (copy has its own format)
	cobraMultiCopyCmd.Use = "copy FROM_EDITOR TO_EDITOR SERVER_NAME"
	cmd.AddCommand(cobraMultiCopyCmd)

	return cmd
}

// createEditorCommand creates a subcommand for a specific editor
func createEditorCommand(editorInfo EditorInfo) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   editorInfo.Name,
		Short: fmt.Sprintf("Manage %s", editorInfo.Description),
		Long:  fmt.Sprintf("Commands for managing %s.", editorInfo.Description),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("verb argument required. Available verbs: add, remove, list, enable, disable, copy, edit, init, tail, add-go-go")
			}
			return fmt.Errorf("unknown verb '%s' for editor '%s'", args[0], editorInfo.Name)
		},
	}

	// Add verb subcommands using new Glazed dual commands
	
	// Add command (dual mode)
	addCmd, err := NewAddCommandDual()
	if err != nil {
		return nil, fmt.Errorf("failed to create add command: %w", err)
	}
	cobraAddCmd, err := cli.BuildCobraCommand(addCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to build add command: %w", err)
	}
	// Override Use to match single editor format
	cobraAddCmd.Use = "add NAME COMMAND [ARGS...]"
	// Add editor context to command
	originalAddRunE := cobraAddCmd.RunE
	if originalAddRunE != nil {
		cobraAddCmd.RunE = func(cmd *cobra.Command, args []string) error {
			// Set the editors flag to this specific editor
			if err := cmd.Flags().Set("editors", editorInfo.Name); err != nil {
				return err
			}
			return originalAddRunE(cmd, args)
		}
	}
	cmd.AddCommand(cobraAddCmd)
	
	// Remove command (dual mode)
	removeCmd, err := NewRemoveCommandDual()
	if err != nil {
		return nil, fmt.Errorf("failed to create remove command: %w", err)
	}
	cobraRemoveCmd, err := cli.BuildCobraCommand(removeCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to build remove command: %w", err)
	}
	// Override Use to match single editor format
	cobraRemoveCmd.Use = "remove NAME"
	// Add editor context to command
	originalRemoveRunE := cobraRemoveCmd.RunE
	if originalRemoveRunE != nil {
		cobraRemoveCmd.RunE = func(cmd *cobra.Command, args []string) error {
			// Set the editors flag to this specific editor
			if err := cmd.Flags().Set("editors", editorInfo.Name); err != nil {
				return err
			}
			return originalRemoveRunE(cmd, args)
		}
	}
	cmd.AddCommand(cobraRemoveCmd)
	
	// List command (single mode Glazed)
	cmd.AddCommand(NewListCommandGlazed(editorInfo.Name))
	
	// Enable command (dual mode)
	enableCmd, err := NewEnableCommandDual()
	if err != nil {
		return nil, fmt.Errorf("failed to create enable command: %w", err)
	}
	cobraEnableCmd, err := cli.BuildCobraCommand(enableCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to build enable command: %w", err)
	}
	// Override Use to match single editor format
	cobraEnableCmd.Use = "enable NAME"
	// Add editor context to command
	originalEnableRunE := cobraEnableCmd.RunE
	if originalEnableRunE != nil {
		cobraEnableCmd.RunE = func(cmd *cobra.Command, args []string) error {
			// Set the editors flag to this specific editor
			if err := cmd.Flags().Set("editors", editorInfo.Name); err != nil {
				return err
			}
			return originalEnableRunE(cmd, args)
		}
	}
	cmd.AddCommand(cobraEnableCmd)
	
	// Disable command (dual mode)
	disableCmd, err := NewDisableCommandDual()
	if err != nil {
		return nil, fmt.Errorf("failed to create disable command: %w", err)
	}
	cobraDisableCmd, err := cli.BuildCobraCommand(disableCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to build disable command: %w", err)
	}
	// Override Use to match single editor format
	cobraDisableCmd.Use = "disable NAME"
	// Add editor context to command
	originalDisableRunE := cobraDisableCmd.RunE
	if originalDisableRunE != nil {
		cobraDisableCmd.RunE = func(cmd *cobra.Command, args []string) error {
			// Set the editors flag to this specific editor
			if err := cmd.Flags().Set("editors", editorInfo.Name); err != nil {
				return err
			}
			return originalDisableRunE(cmd, args)
		}
	}
	cmd.AddCommand(cobraDisableCmd)
	
	// Copy command (dual mode)
	copyCmd, err := NewCopyCommandDual()
	if err != nil {
		return nil, fmt.Errorf("failed to create copy command: %w", err)
	}
	cobraCopyCmd, err := cli.BuildCobraCommand(copyCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to build copy command: %w", err)
	}
	// Override Use to match single editor format
	cobraCopyCmd.Use = "copy FROM_EDITOR TO_EDITOR SERVER_NAME"
	// Add editor context to command (for copy, the context is different as it involves multiple editors)
	cmd.AddCommand(cobraCopyCmd)
	// Create and add Edit command
	editCmd, err := NewEditCommand()
	if err != nil {
		return nil, fmt.Errorf("failed to create edit command: %w", err)
	}
	cobraEditCmd, err := cli.BuildCobraCommandFromBareCommand(editCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to build edit command: %w", err)
	}
	cmd.AddCommand(cobraEditCmd)

	// Create and add Init command
	initCmd, err := NewInitCommand()
	if err != nil {
		return nil, fmt.Errorf("failed to create init command: %w", err)
	}
	cobraInitCmd, err := cli.BuildCobraCommandFromBareCommand(initCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to build init command: %w", err)
	}
	cmd.AddCommand(cobraInitCmd)

	cmd.AddCommand(NewAddGoGoCommand(editorInfo.Name))

	// Add tail command only for Claude (Claude-specific feature)
	if editorInfo.Name == "claude" {
		tailCmd, err := NewTailCommand()
		if err != nil {
			return nil, fmt.Errorf("failed to create tail command: %w", err)
		}
		cobraTailCmd, err := cli.BuildCobraCommandFromBareCommand(tailCmd)
		if err != nil {
			return nil, fmt.Errorf("failed to build tail command: %w", err)
		}
		cmd.AddCommand(cobraTailCmd)
	}

	return cmd, nil
}

// getSupportedEditorNames returns a slice of supported editor names
func getSupportedEditorNames() []string {
	var names []string
	for _, info := range SupportedEditors {
		names = append(names, info.Name)
	}
	return names
}

// printServerDetails prints the details of a server configuration
func printServerDetails(server types.CommonServer, indent string) {
	if server.URL != "" {
		transportType := "HTTP"
		if server.IsSSE {
			transportType = "SSE"
		}
		fmt.Printf("%sTransport: %s\n", indent, transportType)
		fmt.Printf("%sURL: %s\n", indent, server.URL)
	} else {
		fmt.Printf("%sTransport: stdio\n", indent)
		fmt.Printf("%sCommand: %s\n", indent, server.Command)
		if len(server.Args) > 0 {
			fmt.Printf("%sArgs: %v\n", indent, server.Args)
		}
	}

	if len(server.Env) > 0 {
		fmt.Printf("%sEnvironment:\n", indent)
		for k, v := range server.Env {
			fmt.Printf("%s  %s: %s\n", indent, k, v)
		}
	}
}

// validateTransportCompatibility checks if the source server's transport is compatible with the target editor
func validateTransportCompatibility(sourceEditor, targetEditor string, server types.CommonServer) error {
	// Claude only supports stdio transport
	if targetEditor == "claude" && server.URL != "" {
		return fmt.Errorf("claude only supports stdio transport, but source server uses URL-based transport (%s)", server.URL)
	}

	// Add more compatibility checks as needed for other editors
	return nil
}

// NewAddGoGoCommand creates the add-go-go command for a specific editor
func NewAddGoGoCommand(editor string) *cobra.Command {
	var target string
	var env []string
	var overwrite bool
	var additionalArgs []string

	cmd := &cobra.Command{
		Use:   "add-go-go NAME PROFILE [ARGS...]",
		Short: "Add go-go-mcp server",
		Long:  fmt.Sprintf("Add a new MCP server configuration that uses go-go-mcp server for %s.", editor),
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := NewStoreWithTarget(editor, target)
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

			server := types.CommonServer{
				Name:    name,
				Command: "go-go-mcp",
				Args:    cmdArgs,
				Env:     envMap,
			}

			if err := store.AddServer(server, overwrite); err != nil {
				return fmt.Errorf("failed to add server: %w", err)
			}

			if err := store.Save(); err != nil {
				return fmt.Errorf("failed to save configuration: %w", err)
			}

			action := "Added"
			if overwrite {
				action = "Added (overwritten)"
			}

			fmt.Printf("Successfully %s go-go-mcp server '%s' with profile '%s' to %s:\n", action, name, profile, editor)
			fmt.Printf("  Command: %s %s\n", server.Command, strings.Join(server.Args, " "))
			if len(server.Env) > 0 {
				fmt.Printf("  Environment:\n")
				for k, v := range server.Env {
					fmt.Printf("    %s: %s\n", k, v)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&target, "target", "t", "", "Target configuration (e.g., global, cwd)")
	cmd.Flags().StringArrayVarP(&env, "env", "e", []string{}, "Environment variables in KEY=VALUE format")
	cmd.Flags().BoolVarP(&overwrite, "overwrite", "w", false, "Overwrite existing server if it exists")
	cmd.Flags().StringArrayVar(&additionalArgs, "args", []string{}, "Additional arguments to pass to the server command")

	return cmd
}
