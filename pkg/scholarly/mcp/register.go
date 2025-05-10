package mcp

import (
	tool_registry "github.com/go-go-golems/go-go-mcp/pkg/tools/providers/tool-registry"
	"github.com/pkg/errors"
)

// RegisterScholarlyTools registers all scholarly tools with the provided registry
func RegisterScholarlyTools(registry *tool_registry.Registry) error {
	// Register all scholarly tools
	if err := registerSearchWorksTool(registry); err != nil {
		return errors.Wrap(err, "failed to register search_works tool")
	}

	if err := registerResolveDOITool(registry); err != nil {
		return errors.Wrap(err, "failed to register resolve_doi tool")
	}

	if err := registerSuggestKeywordsTool(registry); err != nil {
		return errors.Wrap(err, "failed to register suggest_keywords tool")
	}

	if err := registerGetMetricsTool(registry); err != nil {
		return errors.Wrap(err, "failed to register get_metrics tool")
	}

	if err := registerGetCitationsTool(registry); err != nil {
		return errors.Wrap(err, "failed to register get_citations tool")
	}

	if err := registerFindFullTextTool(registry); err != nil {
		return errors.Wrap(err, "failed to register find_full_text tool")
	}

	return nil
}
