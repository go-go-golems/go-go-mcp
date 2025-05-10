package scholarly_test

import (
	"context"
	"fmt"
	"os"

	"github.com/go-go-golems/go-go-mcp/pkg/scholarly"
	"github.com/go-go-golems/go-go-mcp/pkg/scholarly/querydsl"
)

func Example() {
	// Create a new query using the DSL
	query := querydsl.New().
		WithText("quantum computing").
		WithAuthor("John Smith").
		WithCategory("cs.AI").
		Between(2020, 2024).
		OnlyOA(true).
		Order(querydsl.SortRelevance).
		WithMaxResults(10)

	// Configure search options
	opts := scholarly.SearchOptions{
		// Optionally specify which providers to use (default: all)
		Providers: []scholarly.SearchProvider{
			scholarly.ProviderArxiv,
			scholarly.ProviderCrossref,
			scholarly.ProviderOpenAlex,
		},
		// Add your email for polite pool access
		Mailto: os.Getenv("SCHOLARLY_MAILTO"),
	}

	// Perform the search
	results, err := scholarly.Search(context.Background(), query, opts)
	if err != nil {
		fmt.Printf("Search failed: %v\n", err)
		return
	}

	// Process results
	fmt.Printf("Found %d results\n", len(results))
	for i, result := range results {
		if i >= 3 { // Just show first 3 for example
			break
		}
		fmt.Printf("\n%d. %s\n", i+1, result.Title)
		fmt.Printf("   Authors: %v\n", result.Authors)
		fmt.Printf("   Source: %s\n", result.SourceName)
		if result.DOI != "" {
			fmt.Printf("   DOI: %s\n", result.DOI)
		}
	}
}
