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

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/go-go-mcp/pkg/helpers"
	"github.com/hpcloud/tail"
)

// TailCommand is a Glazed BareCommand for tailing log files (Claude-specific)
type TailCommand struct {
	*cmds.CommandDescription
}

// TailSettings holds the parameters for the tail command
type TailSettings struct {
	Editors     string   `glazed.parameter:"editors"`
	ServerNames []string `glazed.parameter:"server-names"`
	All         bool     `glazed.parameter:"all"`
	Lines       int      `glazed.parameter:"lines"`
}

// Ensure interface is implemented
var _ cmds.BareCommand = &TailCommand{}

// Run implements BareCommand interface
func (c *TailCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	settings := &TailSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, settings); err != nil {
		return fmt.Errorf("failed to parse settings: %w", err)
	}

	// Parse editors from first argument
	editors, err := ParseEditors(settings.Editors)
	if err != nil {
		return err
	}

	// Validate that only Claude is supported
	for _, editor := range editors {
		if editor != "claude" {
			return fmt.Errorf("tail command is only supported for Claude, got: %s", editor)
		}
	}

	// Since we validated all editors are Claude, we can proceed
	// (They should all be the same anyway)

	xdgConfigPath, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("could not get user config directory: %w", err)
	}
	logDir := filepath.Join(xdgConfigPath, "Claude", "logs")

	// Create a context that can be cancelled
	tailCtx, cancel := context.WithCancel(ctx)
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
	if settings.All {
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
	} else if len(settings.ServerNames) == 0 {
		// Only tail main log file
		filesToTail = append(filesToTail, filepath.Join(logDir, "mcp.log"))
	} else {
		// Tail specified server logs
		for _, serverName := range settings.ServerNames {
			filesToTail = append(filesToTail, filepath.Join(logDir, fmt.Sprintf("mcp-server-%s.log", serverName)))
		}
	}

	if len(filesToTail) == 0 {
		return fmt.Errorf("no log files found to tail")
	}

	fmt.Printf("Tailing %d log file(s) for Claude...\n", len(filesToTail))

	// Create a WaitGroup to wait for all tailers to finish
	var wg sync.WaitGroup

	// Start tailing each file
	for _, file := range filesToTail {
		wg.Add(1)
		go func(filename string) {
			defer wg.Done()

			// Find the starting position based on requested lines
			startPos := int64(0)
			if settings.Lines > 0 {
				var err error
				startPos, err = helpers.FindStartPosForLastNLines(filename, settings.Lines)
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
				case <-tailCtx.Done():
					return
				}
			}
		}(file)
	}

	// Wait for all tailers to finish
	wg.Wait()

	return nil
}

// NewTailCommand creates a new tail command
func NewTailCommand() (cmds.BareCommand, error) {
	cmd := &TailCommand{
		CommandDescription: cmds.NewCommandDescription(
			"tail",
			cmds.WithShort("Tail logs"),
			cmds.WithLong("Tail log files for Claude in real-time.\n\nThe tail command is only supported for Claude editors.\n\nWithout server names, tails the main mcp.log file.\nWith server names, tails the corresponding mcp-server-XXX.log files.\nUse --all to tail all log files.\nUse --lines/-n to output the last N lines of each file before tailing.\n\nExamples:\n```\n# Tail the main log file\nmcp editor config claude tail\n\n# Tail specific server logs with line history\nmcp editor config claude tail server1 server2 --lines 20\n\n# Tail all log files\nmcp editor config claude tail --all\n\n# Tail with more line history\nmcp editor config claude tail --lines 50\n```"),

			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"editors",
					parameters.ParameterTypeString,
					parameters.WithHelp("Editor (must be claude)"),
					parameters.WithRequired(true),
				),
				parameters.NewParameterDefinition(
					"server-names",
					parameters.ParameterTypeStringList,
					parameters.WithHelp("Specific server names to tail logs for"),
					parameters.WithRequired(false),
				),
			),

			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"all",
					parameters.ParameterTypeBool,
					parameters.WithHelp("Tail all log files"),
					parameters.WithShortFlag("a"),
					parameters.WithDefault(false),
				),
				parameters.NewParameterDefinition(
					"lines",
					parameters.ParameterTypeInteger,
					parameters.WithHelp("Output the last N lines of each file before tailing"),
					parameters.WithShortFlag("n"),
					parameters.WithDefault(10),
				),
			),
		),
	}

	return cmd, nil
}
