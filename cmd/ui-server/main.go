package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "ui-server [directory]",
	Short: "Start a UI server that renders YAML UI definitions",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		dir := args[0]
		port, _ := cmd.Flags().GetInt("port")

		server := NewServer(dir)
		return server.Start(ctx, port)
	},
}

func init() {
	rootCmd.Flags().IntP("port", "p", 8080, "Port to run the server on")
}
