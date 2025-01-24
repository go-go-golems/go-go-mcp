# Loading Commands from YAML and Repositories

This guide explains how to load commands from YAML files and repositories in go-go-mcp.

## Command Loaders

Command loaders are interfaces that know how to load commands from files. They implement the `CommandLoader` interface:

```go
type CommandLoader interface {
    LoadCommands(fs fs.FS, filePath string, options []CommandDescriptionOption, aliasOptions []AliasOption) ([]Command, error)
    GetFileExtensions() []string
    GetName() string
    IsFileSupported(fs fs.FS, fileName string) bool
}
```

### Shell Command Loader Example

Here's an example of a command loader for shell commands:

```go
type ShellCommandLoader struct{}

func (l *ShellCommandLoader) LoadCommands(
    fs_ fs.FS,
    filePath string,
    options []glazed_cmds.CommandDescriptionOption,
    aliasOptions []alias.Option,
) ([]glazed_cmds.Command, error) {
    // Read file contents
    f, err := fs_.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer f.Close()

    data, err := io.ReadAll(f)
    if err != nil {
        return nil, err
    }

    // Load command from YAML
    cmd, err := LoadShellCommandFromYAML(data)
    if err != nil {
        return nil, err
    }

    // Apply any additional options
    for _, opt := range options {
        opt(cmd.CommandDescription)
    }

    return []glazed_cmds.Command{cmd}, nil
}
```

## Loading from Single Files

To load commands from individual YAML files:

```go
loader := &ShellCommandLoader{}
data, err := os.ReadFile("command.yaml")
if err != nil {
    return err
}

commands, err := loader.LoadCommands(os.DirFS("."), "command.yaml", nil, nil)
if err != nil {
    return err
}
```

## Loading from Repositories

Repositories provide a way to load commands from multiple directories and files. Here's how to use them:

### 1. Create Directory Definitions

```go
directories := []repositories.Directory{
    {
        FS:               os.DirFS("/path/to/commands"),
        RootDirectory:    ".",
        RootDocDirectory: "doc",
        WatchDirectory:   "/path/to/commands",
        Name:             "my-commands",
        SourcePrefix:     "file",
    },
}
```

### 2. Create and Configure Repository

```go
repo := repositories.NewRepository(
    repositories.WithDirectories(directories...),
    repositories.WithCommandLoader(loader),
)
```

### 3. Load Commands

```go
helpSystem := help.NewHelpSystem()
err := repo.LoadCommands(helpSystem)
if err != nil {
    return err
}

// Get all loaded commands
commands := repo.CollectCommands([]string{}, true)
```

## Loading from Multiple Input Sources

You can also load commands from a mix of files and directories:

```go
commands, err := repositories.LoadCommandsFromInputs(
    &ShellCommandLoader{},
    []string{
        "path/to/command.yaml",
        "path/to/commands/directory",
    },
)
```

## Best Practices

1. **Error Handling**: Always check for errors when loading commands and provide meaningful error messages.

2. **File System Abstraction**: Use `fs.FS` interface for file operations to support different file system implementations.

3. **Documentation**: Include documentation directory in your repositories for command help and examples.

4. **Validation**: Validate loaded commands to ensure they meet your application's requirements.

## Example: Loading Shell Commands in MCP Server

Here's a complete example of loading shell commands in an MCP server:

```go
// Create loader and repository directories
loader := &shell.ShellCommandLoader{}
directories := []repositories.Directory{}

// Add directories from configuration
for _, repoPath := range repositoryPaths {
    dir := os.ExpandEnv(repoPath)
    if fi, err := os.Stat(dir); os.IsNotExist(err) || !fi.IsDir() {
        log.Warn().Str("path", dir).Msg("Repository directory does not exist")
        continue
    }
    
    directories = append(directories, repositories.Directory{
        FS:               os.DirFS(dir),
        RootDirectory:    ".",
        RootDocDirectory: "doc",
        WatchDirectory:   dir,
        Name:             dir,
        SourcePrefix:     "file",
    })
}

// Create and load repository
if len(directories) > 0 {
    repo := repositories.NewRepository(
        repositories.WithDirectories(directories...),
        repositories.WithCommandLoader(loader),
    )

    helpSystem := help.NewHelpSystem()
    err := repo.LoadCommands(helpSystem)
    if err != nil {
        log.Error().Err(err).Msg("Error loading commands")
        return err
    }

    // Get all commands
    commands := repo.CollectCommands([]string{}, true)
    log.Info().Int("count", len(commands)).Msg("Loaded commands")
}
```

## Command YAML Format

Commands are defined in YAML files. Here's an example shell command:

```yaml
name: example-command
short: A simple example command
long: |
  A longer description of what this command does
  and how to use it.
flags:
  - name: input
    type: string
    help: Input file path
    required: true
  - name: verbose
    type: bool
    help: Enable verbose output
    default: false
command:
  - echo
  - "Processing {{ .Args.input }}"
environment:
  DEBUG: "{{ if .Args.verbose }}1{{ else }}0{{ end }}"
```

## Troubleshooting

Common issues and solutions:

1. **File Not Found**: Ensure repository paths exist and are accessible.
2. **YAML Parsing Errors**: Validate YAML syntax and required fields.
3. **Command Loading Failures**: Check command loader implementation and file support.
4. **Missing Dependencies**: Verify all required packages are imported. 