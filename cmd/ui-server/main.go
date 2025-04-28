package main

import (
	"fmt"
	"os"

	clay "github.com/go-go-golems/clay/pkg"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/help"
	uicmds "github.com/go-go-golems/go-go-mcp/cmd/ui-server/cmds"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ui-server",
	Short: "Start a UI server that renders YAML UI definitions",
	Long: `A server that renders UI definitions from YAML files.
The server watches for changes in the specified directory and automatically reloads pages.`,
}

func main() {
	helpSystem := help.NewHelpSystem()
	helpSystem.SetupCobraRootCommand(rootCmd)

	startCmd, err := NewStartCommand()
	cobra.CheckErr(err)
	err = clay.InitViper("ui-server", rootCmd)
	cobra.CheckErr(err)

	cobraStartCmd, err := cli.BuildCobraCommandFromBareCommand(startCmd)
	cobra.CheckErr(err)

	rootCmd.AddCommand(cobraStartCmd)

	// Add update-ui command
	updateUICmd, err := uicmds.NewUpdateUICommand()
	cobra.CheckErr(err)
	cobraUpdateUICmd, err := cli.BuildCobraCommandFromBareCommand(updateUICmd)
	cobra.CheckErr(err)
	rootCmd.AddCommand(cobraUpdateUICmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
