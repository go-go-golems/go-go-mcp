package main

import (
	"context"
	"fmt"
	"log"

	"github.com/go-go-golems/go-go-mcp/cmd/experiments/js-web-server/internal/repository"
)

func main() {
	// Create repository manager
	repos, err := repository.NewSQLiteRepositoryManager("data.sqlite")
	if err != nil {
		log.Fatalf("Failed to create repository manager: %v", err)
	}
	defer func() {
		if err := repos.Close(); err != nil {
			log.Printf("Failed to close repository: %v", err)
		}
	}()

	ctx := context.Background()

	// Get execution stats
	stats, err := repos.Executions().GetExecutionStats(ctx)
	if err != nil {
		log.Fatalf("Failed to get execution stats: %v", err)
	}

	fmt.Printf("Execution Statistics:\n")
	fmt.Printf("  Total Executions: %d\n", stats.TotalExecutions)
	fmt.Printf("  Successful: %d\n", stats.SuccessfulExecutions)
	fmt.Printf("  Failed: %d\n", stats.FailedExecutions)
	fmt.Printf("  By Source:\n")
	for source, count := range stats.ExecutionsBySource {
		fmt.Printf("    %s: %d\n", source, count)
	}

	// List recent executions
	fmt.Printf("\nRecent Executions:\n")
	filter := repository.ExecutionFilter{}
	pagination := repository.PaginationOptions{Limit: 5, Offset: 0}

	result, err := repos.Executions().ListExecutions(ctx, filter, pagination)
	if err != nil {
		log.Fatalf("Failed to list executions: %v", err)
	}

	for _, exec := range result.Executions {
		codePreview := exec.Code
		if len(codePreview) > 50 {
			codePreview = codePreview[:50] + "..."
		}
		fmt.Printf("  [%d] %s - %s\n", exec.ID, exec.SessionID[:8], codePreview)
		if exec.Error != nil {
			fmt.Printf("      Error: %s\n", *exec.Error)
		}
		if exec.Result != nil {
			fmt.Printf("      Result: %s\n", *exec.Result)
		}
		fmt.Printf("      Time: %s\n", exec.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Println()
	}
}
