package mcp

import (
	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/common"
)

// SearchResponse represents the response from a scholarly search
type SearchResponse struct {
	Results []common.SearchResult `json:"results"`
}
