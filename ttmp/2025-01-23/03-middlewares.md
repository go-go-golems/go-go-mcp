I'll help you write a comprehensive tutorial on using middlewares for parameter parsing in Glazed. Let me break this down into clear sections with examples.

# Tutorial: Parameter Parsing with Glazed Middlewares

## Introduction

Glazed provides a flexible system for handling command parameters through layers and middlewares. This tutorial will show you how to:

1. Define parameter layers
2. Create and use middlewares to populate parameters
3. Handle different parameter sources
4. Chain multiple middlewares together

## 1. Basic Setup

First, let's create a basic parameter layer structure:

```go
package main

import (
    "github.com/go-go-golems/glazed/pkg/cmds/layers"
    "github.com/go-go-golems/glazed/pkg/cmds/parameters"
    "github.com/go-go-golems/glazed/pkg/cmds/middlewares"
)

func main() {
    // Create a new parameter layer
    layer, err := layers.NewParameterLayer(
        "config",
        "Configuration Options",
        layers.WithParameterDefinitions(
            parameters.NewParameterDefinition(
                "host",
                parameters.ParameterTypeString,
                parameters.WithDefault("localhost"),
                parameters.WithHelp("Server hostname"),
            ),
            parameters.NewParameterDefinition(
                "port",
                parameters.ParameterTypeInteger,
                parameters.WithDefault(8080),
                parameters.WithHelp("Server port"),
            ),
        ),
    )
    if err != nil {
        panic(err)
    }

    // Create parameter layers container
    parameterLayers := layers.NewParameterLayers(
        layers.WithLayers(layer),
    )
}
```

## 2. Using SetFromDefaults Middleware

The simplest middleware is `SetFromDefaults`, which populates parameters with their default values:

```go
func useDefaultsMiddleware() {
    // Create empty parsed layers
    parsedLayers := layers.NewParsedLayers()

    // Create and execute the middleware
    middleware := middlewares.SetFromDefaults(
        parameters.WithParseStepSource(parameters.SourceDefaults),
    )

    err := middlewares.ExecuteMiddlewares(
        parameterLayers,
        parsedLayers,
        middleware,
    )
    if err != nil {
        panic(err)
    }

    // Access the parsed values
    configLayer, _ := parsedLayers.Get("config")
    hostValue, _ := configLayer.GetParameter("host")
    // hostValue will be "localhost"
}
```

## 3. Using UpdateFromMap Middleware

The `UpdateFromMap` middleware allows you to update parameters from a map structure:

```go
func useMapMiddleware() {
    parsedLayers := layers.NewParsedLayers()

    // Define the update map
    updateMap := map[string]map[string]interface{}{
        "config": {
            "host": "example.com",
            "port": 9090,
        },
    }

    err := middlewares.ExecuteMiddlewares(
        parameterLayers,
        parsedLayers,
        middlewares.UpdateFromMap(updateMap),
    )
    if err != nil {
        panic(err)
    }

    // Access the updated values
    configLayer, _ := parsedLayers.Get("config")
    hostValue, _ := configLayer.GetParameter("host")
    // hostValue will be "example.com"
}
```

## 4. Using UpdateFromMapAsDefault Middleware

This middleware only sets values if they haven't been set before:

```go
func useMapAsDefaultMiddleware() {
    parsedLayers := layers.NewParsedLayers()

    defaultMap := map[string]map[string]interface{}{
        "config": {
            "host": "fallback.com",
            "port": 7070,
        },
    }

    err := middlewares.ExecuteMiddlewares(
        parameterLayers,
        parsedLayers,
        middlewares.UpdateFromMapAsDefault(defaultMap),
    )
    if err != nil {
        panic(err)
    }
}
```

## 5. Chaining Multiple Middlewares

You can chain middlewares to handle multiple parameter sources with precedence:

```go
func chainMiddlewares() {
    parsedLayers := layers.NewParsedLayers()

    // Define different parameter sources
    configMap := map[string]map[string]interface{}{
        "config": {
            "host": "config.com",
            "port": 5000,
        },
    }

    defaultMap := map[string]map[string]interface{}{
        "config": {
            "host": "default.com",
            "port": 8080,
        },
    }

    // Execute middlewares in order (last middleware has highest precedence)
    err := middlewares.ExecuteMiddlewares(
        parameterLayers,
        parsedLayers,
        middlewares.UpdateFromMapAsDefault(defaultMap),  // Lowest precedence
        middlewares.SetFromDefaults(parameters.WithParseStepSource(parameters.SourceDefaults)),
        middlewares.UpdateFromMap(configMap),  // Highest precedence
    )
    if err != nil {
        panic(err)
    }
}
```

## 6. Using Restricted Layers

You can restrict which layers are affected by middlewares:

```go
func useRestrictedLayers() {
    parsedLayers := layers.NewParsedLayers()

    updateMap := map[string]map[string]interface{}{
        "config": {
            "host": "restricted.com",
        },
    }

    // Only apply to whitelisted layers
    whitelistedMiddleware := middlewares.WrapWithWhitelistedLayers(
        []string{"config"},
        middlewares.UpdateFromMap(updateMap),
    )

    // Or blacklist specific layers
    blacklistedMiddleware := middlewares.WrapWithBlacklistedLayers(
        []string{"other-layer"},
        middlewares.UpdateFromMap(updateMap),
    )

    err := middlewares.ExecuteMiddlewares(
        parameterLayers,
        parsedLayers,
        whitelistedMiddleware,
        blacklistedMiddleware,
    )
    if err != nil {
        panic(err)
    }
}
```

## 7. Accessing Parsed Values

After applying middlewares, you can access the parsed values in several ways:

```go
func accessParsedValues(parsedLayers *layers.ParsedLayers) {
    // 1. Direct access through layer
    configLayer, _ := parsedLayers.Get("config")
    hostValue, _ := configLayer.GetParameter("host")

    // 2. Get all parameters as a map
    dataMap := parsedLayers.GetDataMap()
    host := dataMap["host"]

    // 3. Initialize a struct
    type Config struct {
        Host string `glazed.parameter:"host"`
        Port int    `glazed.parameter:"port"`
    }
    
    var config Config
    err := parsedLayers.InitializeStruct("config", &config)
    if err != nil {
        panic(err)
    }
}
```

## 8. Tracking Parameter History

Each parameter update is tracked in the parsing history:

```go
func checkParameterHistory(parsedLayers *layers.ParsedLayers) {
    configLayer, _ := parsedLayers.Get("config")
    hostParam, _ := configLayer.Parameters.Get("host")

    // View the parsing history
    for _, step := range hostParam.Log {
        fmt.Printf("Source: %s, Value: %v\n", step.Source, step.Value)
    }
}
```

This tutorial covers the main aspects of parameter parsing with Glazed middlewares. The middleware system is extensible, allowing you to create custom middlewares for your specific needs.

Remember that:
1. Middlewares are executed in order, with later middlewares taking precedence
2. Each parameter update is tracked in the parsing history
3. You can restrict which layers are affected by specific middlewares
4. The system supports multiple parameter sources and formats

