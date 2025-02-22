package main

import (
	"fmt"
	"os"

	clay "github.com/go-go-golems/clay/pkg"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ui-server",
	Short: "Start a UI server that renders YAML UI definitions",
	Long: `A server that renders UI definitions from YAML files.
The server watches for changes in the specified directory and automatically reloads pages.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// reinitialize the logger because we can now parse --log-level and co
		// from the command line flag
		err := clay.InitLogger()
		cobra.CheckErr(err)
	},
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

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
