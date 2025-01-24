package cmds

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version information
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "none"
)

// VersionCmd handles the "version" command
var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("mcp-client version %s\n", Version)
		fmt.Printf("  Build time: %s\n", BuildTime)
		fmt.Printf("  Git commit: %s\n", GitCommit)
	},
}
