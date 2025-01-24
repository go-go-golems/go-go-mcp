# Converting Cobra Commands to Glazed Commands

This tutorial explains how to convert traditional Cobra commands into Glazed commands with parameter layers. We'll use the transformation of the MCP client's prompts commands as an example.

## Overview

The process involves:
1. Creating parameter layers for settings
2. Converting command flags to parameter definitions
3. Implementing Glazed command interfaces
4. Updating command initialization and registration

## Step 1: Identify Command Settings 

First, identify all settings and flags used by your command. Group related settings into structs with glazed tags:

```go
// Before: Cobra flags
var (
    transport string
    server string
    command []string
)

// After: Settings struct with glazed tags
type ClientSettings struct {
    Transport string   `glazed.parameter:"transport"`
    Server    string   `glazed.parameter:"server"`
    Command   []string `glazed.parameter:"command"`
}
```

## OPTIONAL: Step 2: Create Parameter Layers (optional, only when asked for explicitly)

Create a new layer file (e.g., `layers/client.go`) to define your parameter layers:

```go
package layers

import (
    glazed_layers "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

const ClientLayerSlug = "my-layer"

func NewClientParameterLayer() (glazed_layers.ParameterLayer, error) {
    return glazed_layers.NewParameterLayer(ClientLayerSlug, "Layer Description",
        glazed_layers.WithParameterDefinitions(
            parameters.NewParameterDefinition(
                "transport",
                parameters.ParameterTypeString,
                parameters.WithHelp("Transport type"),
                parameters.WithDefault("command"),
            ),
            // Add more parameters...
        ),
    )
}
```

## Step 3: Convert Command Structure

Transform your Cobra command into a Glazed command:

```go
// Before: Cobra command
var myCmd = &cobra.Command{
    Use: "my-command",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Command logic
    },
}

// After: Glazed command
type MyCommand struct {
    *cmds.CommandDescription
}

type MyCommandSettings struct {
    // Command-specific settings
    Args string `glazed.parameter:"args"`
}

func NewMyCommand() (*MyCommand, error) {
    // Create required layers
    glazedParameterLayer, err := settings.NewGlazedParameterLayers()
    if err != nil {
        return nil, err
    }

    clientLayer, err := layers.NewClientParameterLayer()
    if err != nil {
        return nil, err
    }

    return &MyCommand{
        CommandDescription: cmds.NewCommandDescription(
            "my-command",
            cmds.WithShort("Command description"),
            cmds.WithFlags(
                parameters.NewParameterDefinition(
                    "args",
                    parameters.ParameterTypeString,
                    parameters.WithHelp("Arguments"),
                    parameters.WithDefault(""),
                ),
            ),
            cmds.WithLayersList(
                glazedParameterLayer,
                clientLayer,
            ),
        ),
    }, nil
}
```

## Step 4: Implement Command Interface

Choose between GlazeCommand (structured output) or WriterCommand (formatted text):

```go
// For structured data output:
func (c *MyCommand) RunIntoGlazeProcessor(
    ctx context.Context,
    parsedLayers *glazed_layers.ParsedLayers,
    gp middlewares.Processor,
) error {
    // Initialize settings
    s := &MyCommandSettings{}
    if err := parsedLayers.InitializeStruct(glazed_layers.DefaultSlug, s); err != nil {
        return err
    }

    // Your command logic here
    row := types.NewRow(
        types.MRP("field1", "value1"),
        types.MRP("field2", "value2"),
    )
    return gp.AddRow(ctx, row)
}

// For text output:
func (c *MyCommand) RunIntoWriter(
    ctx context.Context,
    parsedLayers *glazed_layers.ParsedLayers,
    w io.Writer,
) error {
    s := &MyCommandSettings{}
    if err := parsedLayers.InitializeStruct(glazed_layers.DefaultSlug, s); err != nil {
        return err
    }

    // Your command logic here
    _, err := fmt.Fprintf(w, "Output: %s\n", result)
    return err
}
```

## Step 5: Update Command Registration

Update the command initialization to use Glazed builders:

```go
func init() {
    myCmd, err := NewMyCommand()
    cobra.CheckErr(err)

    // For GlazeCommand:
    cobraCmd, err := cli.BuildCobraCommandFromGlazeCommand(myCmd)
    // OR for WriterCommand:
    cobraCmd, err := cli.BuildCobraCommandFromWriterCommand(myCmd)
    cobra.CheckErr(err)

    rootCmd.AddCommand(cobraCmd)
}
```

## Best Practices

1. **Layer Organization**:
   - Group related parameters into their own layers
   - Use meaningful layer slugs
   - Consider reusability across commands

2. **Settings Structs**:
   - Use clear, descriptive field names
   - Add proper glazed tags
   - Group related settings together

3. **Parameter Definitions**:
   - Provide helpful descriptions
   - Set appropriate defaults
   - Use correct parameter types

4. **Error Handling**:
   - Check layer initialization errors
   - Provide meaningful error messages
   - Clean up resources properly

## Example: Converting Prompts Command

Here's how we converted the MCP client's prompts execute command:

```go
// Before:
var executePromptCmd = &cobra.Command{
    Use:   "execute [prompt-name]",
    Short: "Execute a specific prompt",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        client, err := helpers.CreateClient(cmd)
        if err != nil {
            return err
        }
        defer client.Close(cmd.Context())

        promptArgMap := make(map[string]string)
        if promptArgs != "" {
            if err := json.Unmarshal([]byte(promptArgs), &promptArgMap); err != nil {
                return fmt.Errorf("invalid prompt arguments JSON: %w", err)
            }
        }

        message, err := client.GetPrompt(cmd.Context(), args[0], promptArgMap)
        if err != nil {
            return err
        }

        fmt.Printf("Role: %s\n", message.Role)
        fmt.Printf("Content: %s\n", message.Content.Text)
        return nil
    },
}

// After:
type ExecutePromptCommand struct {
    *cmds.CommandDescription
}

type ExecutePromptSettings struct {
    Args       string `glazed.parameter:"args"`
    PromptName string `glazed.parameter:"prompt-name"`
}

func NewExecutePromptCommand() (*ExecutePromptCommand, error) {
    clientLayer, err := layers.NewClientParameterLayer()
    if err != nil {
        return nil, errors.Wrap(err, "could not create client parameter layer")
    }

    return &ExecutePromptCommand{
        CommandDescription: cmds.NewCommandDescription(
            "execute",
            cmds.WithShort("Execute a specific prompt"),
            cmds.WithFlags(
                parameters.NewParameterDefinition(
                    "args",
                    parameters.ParameterTypeString,
                    parameters.WithHelp("Prompt arguments as JSON string"),
                    parameters.WithDefault(""),
                ),
            ),
            cmds.WithArguments(
                parameters.NewParameterDefinition(
                    "prompt-name",
                    parameters.ParameterTypeString,
                    parameters.WithHelp("Name of the prompt to execute"),
                    parameters.WithRequired(true),
                ),
            ),
            cmds.WithLayersList(
                clientLayer,
            ),
        ),
    }, nil
}

func (c *ExecutePromptCommand) RunIntoWriter(
    ctx context.Context,
    parsedLayers *glazed_layers.ParsedLayers,
    w io.Writer,
) error {
    s := &ExecutePromptSettings{}
    if err := parsedLayers.InitializeStruct(glazed_layers.DefaultSlug, s); err != nil {
        return err
    }

    client, err := helpers.CreateClientFromSettings(parsedLayers)
    if err != nil {
        return err
    }
    defer client.Close(ctx)

    promptArgMap := make(map[string]string)
    if s.Args != "" {
        if err := json.Unmarshal([]byte(s.Args), &promptArgMap); err != nil {
            return fmt.Errorf("invalid prompt arguments JSON: %w", err)
        }
    }

    message, err := client.GetPrompt(ctx, s.PromptName, promptArgMap)
    if err != nil {
        return err
    }

    _, err = fmt.Fprintf(w, "Role: %s\nContent: %s\n", message.Role, message.Content.Text)
    return err
}
```

## Benefits

1. **Structured Settings**: Better organization of command parameters
2. **Reusable Layers**: Share common settings across commands
3. **Type Safety**: Early detection of configuration errors
4. **Consistent Interface**: Uniform handling of command execution
5. **Better Output Control**: Flexible output formatting options

## Common Pitfalls

1. **Layer Naming**: Avoid generic names, be specific about layer purpose
2. **Parameter Types**: Use appropriate types (e.g., StringList for arrays)
3. **Error Handling**: Don't forget to check layer initialization errors
4. **Resource Cleanup**: Always close/cleanup resources in commands
5. **Documentation**: Keep help messages and documentation up to date

Remember to update tests and documentation when converting commands, and consider backward compatibility if needed.