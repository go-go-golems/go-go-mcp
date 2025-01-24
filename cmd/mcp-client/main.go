package main

import (
	clay "github.com/go-go-golems/clay/pkg"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/go-go-mcp/cmd/mcp-client/cmds"
	"github.com/spf13/cobra"
)

var (
	// Version information
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "none"
)

var rootCmd = &cobra.Command{
	Use:   "mcp-client",
	Short: "MCP client implementation in Go",
}

func main() {
	helpSystem, err := initRootCmd()
	cobra.CheckErr(err)

	err = initAllCommands(helpSystem)
	cobra.CheckErr(err)

	err = rootCmd.Execute()
	cobra.CheckErr(err)
}

func initRootCmd() (*help.HelpSystem, error) {
	helpSystem := help.NewHelpSystem()
	helpSystem.SetupCobraRootCommand(rootCmd)

	err := clay.InitViper("mcp", rootCmd)
	cobra.CheckErr(err)
	err = clay.InitLogger()
	cobra.CheckErr(err)

	return helpSystem, nil
}

func initAllCommands(_ *help.HelpSystem) error {
	// Add existing commands
	rootCmd.AddCommand(cmds.ToolsCmd)
	rootCmd.AddCommand(cmds.ResourcesCmd)
	rootCmd.AddCommand(cmds.PromptsCmd)

	return nil
}
