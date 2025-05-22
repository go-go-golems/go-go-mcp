package main

import (
	"context"
	"fmt"
	"log"

	"github.com/go-go-golems/go-go-mcp/pkg/embeddable"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/session"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "session-app",
		Short: "Application demonstrating session management",
	}

	err := embeddable.AddMCPCommand(rootCmd,
		embeddable.WithName("Session Demo MCP Server"),
		embeddable.WithVersion("1.0.0"),
		embeddable.WithServerDescription("Demonstration of session management capabilities"),
		embeddable.WithTool("get_counter", getCounterHandler,
			embeddable.WithDescription("Get the current counter value for this session"),
		),
		embeddable.WithTool("increment_counter", incrementCounterHandler,
			embeddable.WithDescription("Increment the counter for this session"),
			embeddable.WithIntArg("amount", "Amount to increment by", false),
		),
		embeddable.WithTool("reset_counter", resetCounterHandler,
			embeddable.WithDescription("Reset the counter for this session"),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func getCounterHandler(ctx context.Context, args map[string]interface{}) (*protocol.ToolResult, error) {
	// Session information is automatically available via context
	sess, ok := session.GetSessionFromContext(ctx)
	if !ok {
		return protocol.NewErrorToolResult(protocol.NewTextContent("No session found")), nil
	}

	// Get counter value from session
	counterVal, ok := sess.GetData("counter")
	counter := 0
	if ok {
		counter, _ = counterVal.(int)
	}

	return protocol.NewToolResult(
		protocol.WithText(fmt.Sprintf("Counter value: %d (Session: %s)", counter, sess.ID)),
	), nil
}

func incrementCounterHandler(ctx context.Context, args map[string]interface{}) (*protocol.ToolResult, error) {
	// Access session via context
	sess, ok := session.GetSessionFromContext(ctx)
	if !ok {
		return protocol.NewErrorToolResult(protocol.NewTextContent("No session found")), nil
	}

	// Get increment amount (default to 1)
	amount := 1
	if amountVal, ok := args["amount"].(float64); ok {
		amount = int(amountVal)
	}

	// Get current counter value
	counterVal, ok := sess.GetData("counter")
	counter := 0
	if ok {
		counter, _ = counterVal.(int)
	}

	// Increment and store back to session
	counter += amount
	sess.SetData("counter", counter)

	return protocol.NewToolResult(
		protocol.WithText(fmt.Sprintf("Counter incremented by %d to %d (Session: %s)",
			amount, counter, sess.ID)),
	), nil
}

func resetCounterHandler(ctx context.Context, args map[string]interface{}) (*protocol.ToolResult, error) {
	// Access session via context
	sess, ok := session.GetSessionFromContext(ctx)
	if !ok {
		return protocol.NewErrorToolResult(protocol.NewTextContent("No session found")), nil
	}

	// Reset counter to 0
	sess.SetData("counter", 0)

	return protocol.NewToolResult(
		protocol.WithText(fmt.Sprintf("Counter reset to 0 (Session: %s)", sess.ID)),
	), nil
}
