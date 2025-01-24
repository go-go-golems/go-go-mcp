package cursor

import (
	"github.com/go-go-golems/go-go-mcp/pkg/tools"
)

// RegisterCursorTools registers all cursor-related tools with the registry
func RegisterCursorTools(registry *tools.Registry) error {
	// Register conversation tools
	if err := RegisterGetConversationTool(registry); err != nil {
		return err
	}
	if err := RegisterFindConversationsTool(registry); err != nil {
		return err
	}

	// Register code analysis tools
	if err := RegisterExtractCodeBlocksTool(registry); err != nil {
		return err
	}
	if err := RegisterTrackFileModificationsTool(registry); err != nil {
		return err
	}

	// Register context tools
	if err := RegisterGetFileReferencesTool(registry); err != nil {
		return err
	}
	if err := RegisterGetConversationContextTool(registry); err != nil {
		return err
	}

	return nil
}
