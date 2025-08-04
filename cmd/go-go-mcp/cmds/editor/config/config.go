package config

import (
	"fmt"

	"github.com/go-go-golems/go-go-mcp/pkg/core/configstore"
	"github.com/spf13/cobra"
)

// ConfigCommandOptions holds common options for config commands
type ConfigCommandOptions struct {
	Editor string
	Target string
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

// NewConfigCommand creates the main config command that delegates to editor-specific subcommands
func NewConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config <editor> <verb> [options]",
		Short: "Unified configuration management for MCP editors",
		Long: `Manage MCP server configurations across different editors.

Supported editors: claude, cursor, ampcode, amp, crush

Available verbs:
  add          Add MCP server
  remove       Remove MCP server  
  list         List servers
  enable       Enable server
  disable      Disable server
  edit         Edit config file
  init         Initialize config
  tail         Tail logs (Claude-specific)
  add-go-go    Add go-go-mcp server

Examples:
  mcp editor config claude add myserver /path/to/command
  mcp editor config cursor list --target global
  mcp editor config amp remove myserver`,
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
	
	// Add subcommands for each editor
	for _, editorInfo := range SupportedEditors {
		editorCmd := createEditorCommand(editorInfo)
		cmd.AddCommand(editorCmd)
	}
	
	return cmd
}

// createEditorCommand creates a subcommand for a specific editor
func createEditorCommand(editorInfo EditorInfo) *cobra.Command {
	cmd := &cobra.Command{
		Use:   editorInfo.Name,
		Short: fmt.Sprintf("Manage %s", editorInfo.Description),
		Long:  fmt.Sprintf("Commands for managing %s.", editorInfo.Description),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("verb argument required. Available verbs: add, remove, list, enable, disable, edit, init, tail, add-go-go")
			}
			return fmt.Errorf("unknown verb '%s' for editor '%s'", args[0], editorInfo.Name)
		},
	}
	
	// Add verb subcommands
	cmd.AddCommand(NewAddCommand(editorInfo.Name))
	cmd.AddCommand(NewRemoveCommand(editorInfo.Name))
	cmd.AddCommand(NewListCommand(editorInfo.Name))
	cmd.AddCommand(NewEnableCommand(editorInfo.Name))
	cmd.AddCommand(NewDisableCommand(editorInfo.Name))
	cmd.AddCommand(NewEditCommand(editorInfo.Name))
	cmd.AddCommand(NewInitCommand(editorInfo.Name))
	cmd.AddCommand(NewAddGoGoCommand(editorInfo.Name))
	
	// Add tail command only for Claude (Claude-specific feature)
	if editorInfo.Name == "claude" {
		cmd.AddCommand(NewTailCommand(editorInfo.Name))
	}
	
	return cmd
}

// getSupportedEditorNames returns a slice of supported editor names
func getSupportedEditorNames() []string {
	var names []string
	for _, info := range SupportedEditors {
		names = append(names, info.Name)
	}
	return names
}
