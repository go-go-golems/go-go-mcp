package cmds

import (
	"io"
	"io/fs"
	"os"
	"strings"

	glazed_cmds "github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/alias"
	"github.com/go-go-golems/glazed/pkg/cmds/loaders"
	"github.com/pkg/errors"
)

type ShellCommandLoader struct{}

var _ loaders.CommandLoader = &ShellCommandLoader{}

func (l *ShellCommandLoader) LoadCommands(
	fs_ fs.FS,
	filePath string,
	options []glazed_cmds.CommandDescriptionOption,
	aliasOptions []alias.Option,
) ([]glazed_cmds.Command, error) {
	f, err := fs_.Open(filePath)
	if err != nil {
		return nil, errors.Wrapf(err, "could not open file %s", filePath)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			if err == nil {
				err = errors.Wrapf(closeErr, "could not close file %s", filePath)
			}
		}
	}()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read file %s", filePath)
	}

	cmd, err := LoadShellCommandFromYAML(data)
	if err != nil {
		return nil, errors.Wrapf(err, "could not load shell command from file %s", filePath)
	}

	// Apply any additional options
	for _, opt := range options {
		opt(cmd.CommandDescription)
	}

	return []glazed_cmds.Command{cmd}, nil
}

func (l *ShellCommandLoader) GetFileExtensions() []string {
	return []string{".yaml", ".yml"}
}

func (l *ShellCommandLoader) GetName() string {
	return "shell"
}

func (l *ShellCommandLoader) IsFileSupported(f fs.FS, fileName string) bool {
	return strings.HasSuffix(fileName, ".yaml") || strings.HasSuffix(fileName, ".yml")
}

func LoadShellCommand(path string) (*ShellCommand, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read file %s", path)
	}

	return LoadShellCommandFromYAML(data)
}
