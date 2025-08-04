package config

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/go-go-golems/go-go-mcp/pkg/helpers"
	"github.com/hpcloud/tail"
	"github.com/spf13/cobra"
)

// NewTailCommand creates the tail command for Claude (Claude-specific feature)
func NewTailCommand(editor string) *cobra.Command {
	var all bool
	var lines int

	cmd := &cobra.Command{
		Use:   "tail [server-names...]",
		Short: "Tail logs",
		Long: `Tail log files for Claude in real-time.
Without arguments, tails the main mcp.log file.
With server names, tails the corresponding mcp-server-XXX.log files.
Use --all to tail all log files.
Use --lines/-n to output the last N lines of each file before tailing.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if editor != "claude" {
				return fmt.Errorf("tail command is only supported for Claude")
			}

			xdgConfigPath, err := os.UserConfigDir()
			if err != nil {
				return fmt.Errorf("could not get user config directory: %w", err)
			}
			logDir := filepath.Join(xdgConfigPath, "Claude", "logs")

			// Create a context that can be cancelled
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Set up signal handling for graceful shutdown
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
			go func() {
				<-sigChan
				cancel()
			}()

			// Determine which files to tail
			var filesToTail []string
			if all {
				// Find all log files
				entries, err := os.ReadDir(logDir)
				if err != nil {
					return fmt.Errorf("could not read log directory: %w", err)
				}
				for _, entry := range entries {
					if !entry.IsDir() && strings.HasPrefix(entry.Name(), "mcp") && strings.HasSuffix(entry.Name(), ".log") {
						filesToTail = append(filesToTail, filepath.Join(logDir, entry.Name()))
					}
				}
			} else if len(args) == 0 {
				// Only tail main log file
				filesToTail = append(filesToTail, filepath.Join(logDir, "mcp.log"))
			} else {
				// Tail specified server logs
				for _, serverName := range args {
					filesToTail = append(filesToTail, filepath.Join(logDir, fmt.Sprintf("mcp-server-%s.log", serverName)))
				}
			}

			// Create a WaitGroup to wait for all tailers to finish
			var wg sync.WaitGroup

			// Start tailing each file
			for _, file := range filesToTail {
				wg.Add(1)
				go func(filename string) {
					defer wg.Done()

					// Find the starting position based on requested lines
					startPos := int64(0)
					if lines > 0 {
						var err error
						startPos, err = helpers.FindStartPosForLastNLines(filename, lines)
						if err != nil && !os.IsNotExist(err) {
							fmt.Fprintf(os.Stderr, "Error finding start position in %s: %v\n", filename, err)
						}
					}

					t, err := tail.TailFile(filename, tail.Config{
						Follow: true,
						ReOpen: true,
						Logger: tail.DiscardingLogger,
						Location: &tail.SeekInfo{
							Offset: startPos,
							Whence: io.SeekStart,
						},
					})
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error tailing %s: %v\n", filename, err)
						return
					}
					defer t.Cleanup()

					// Print the filename as a header
					fmt.Printf("==> %s <==\n", filename)

					// Read lines until context is cancelled
					for {
						select {
						case line, ok := <-t.Lines:
							if !ok {
								return
							}
							if line.Err != nil {
								fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", filename, line.Err)
								continue
							}
							fmt.Printf("%s: %s\n", filepath.Base(filename), line.Text)
						case <-ctx.Done():
							return
						}
					}
				}(file)
			}

			// Wait for all tailers to finish
			wg.Wait()

			return nil
		},
	}

	cmd.Flags().BoolVarP(&all, "all", "a", false, "Tail all log files")
	cmd.Flags().IntVarP(&lines, "lines", "n", 10, "Output the last N lines of each file before tailing")

	return cmd
}
