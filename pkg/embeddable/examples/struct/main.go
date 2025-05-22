package main

import (
	"context"
	"fmt"
	"log"

	"github.com/go-go-golems/go-go-mcp/pkg/embeddable"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	tool_registry "github.com/go-go-golems/go-go-mcp/pkg/tools/providers/tool-registry"
	"github.com/spf13/cobra"
)

type DatabaseService struct {
	connectionString string
}

type QueryArgs struct {
	Query string `json:"query" description:"SQL query to execute"`
	Limit int    `json:"limit,omitempty" description:"Maximum number of rows to return"`
}

func (db *DatabaseService) ExecuteQuery(ctx context.Context, args QueryArgs) (*protocol.ToolResult, error) {
	// Implementation here
	return protocol.NewToolResult(
		protocol.WithText(fmt.Sprintf("Executed query: %s (limit: %d)", args.Query, args.Limit)),
	), nil
}

type MathService struct{}

type AddArgs struct {
	A int `json:"a" description:"First number"`
	B int `json:"b" description:"Second number"`
}

func (m *MathService) Add(ctx context.Context, args AddArgs) (*protocol.ToolResult, error) {
	result := args.A + args.B
	return protocol.NewToolResult(
		protocol.WithText(fmt.Sprintf("%d + %d = %d", args.A, args.B, result)),
	), nil
}

func (m *MathService) GetPi(ctx context.Context) (*protocol.ToolResult, error) {
	return protocol.NewToolResult(
		protocol.WithText("π ≈ 3.14159"),
	), nil
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "struct-app",
		Short: "Application demonstrating struct-based tool registration",
	}

	dbService := &DatabaseService{connectionString: "postgresql://localhost:5432/mydb"}
	mathService := &MathService{}

	config := embeddable.NewServerConfig()
	config.Name = "Struct-based MCP Server"
	config.Version = "1.0.0"
	config.Description = "Demonstration of struct-based tool registration"

	// Register struct-based tools
	err := embeddable.RegisterStructTool(config, "execute_query", dbService, "ExecuteQuery")
	if err != nil {
		log.Fatal("Failed to register execute_query tool:", err)
	}

	err = embeddable.RegisterStructTool(config, "add", mathService, "Add")
	if err != nil {
		log.Fatal("Failed to register add tool:", err)
	}

	err = embeddable.RegisterStructTool(config, "get_pi", mathService, "GetPi")
	if err != nil {
		log.Fatal("Failed to register get_pi tool:", err)
	}

	// Add MCP command with the configured tools
	err = embeddable.AddMCPCommand(rootCmd,
		embeddable.WithToolRegistry(config.GetToolProvider().(*tool_registry.Registry)),
		embeddable.WithName(config.Name),
		embeddable.WithVersion(config.Version),
		embeddable.WithServerDescription(config.Description),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
