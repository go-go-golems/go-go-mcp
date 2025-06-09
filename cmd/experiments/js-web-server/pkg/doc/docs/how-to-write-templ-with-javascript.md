# How to Write Templ Templates with JavaScript Integration

## Overview

This guide covers best practices for integrating JavaScript with templ templates, including how to handle events, pass data, and avoid common pitfalls. This is based on [templ's official JavaScript documentation](https://templ.guide/syntax-and-usage/script-templates/) and lessons learned from our implementation.

## Table of Contents

1. [Basic JavaScript Integration](#basic-javascript-integration)
2. [Passing Data from Go to JavaScript](#passing-data-from-go-to-javascript)
3. [Event Handling Best Practices](#event-handling-best-practices)
4. [Common Mistakes and Solutions](#common-mistakes-and-solutions)
5. [Advanced Patterns](#advanced-patterns)
6. [Security Considerations](#security-considerations)

## Basic JavaScript Integration

### Standard Script Tags

The simplest way to include JavaScript in templ templates is using standard `<script>` tags:

```go
templ myComponent() {
    <script>
        function handleClick() {
            alert('Button clicked!');
        }
    </script>
    <button onclick="handleClick()">Click me</button>
}
```

### Rendering Scripts Once

Use `templ.OnceHandle` to ensure scripts are only rendered once per HTTP response:

```go
var scriptHandle = templ.NewOnceHandle()

templ myComponent() {
    @scriptHandle.Once() {
        <script>
            function globalFunction() {
                console.log('This script runs once per page');
            }
        </script>
    }
    <button onclick="globalFunction()">Click me</button>
}
```

## Passing Data from Go to JavaScript

### Method 1: Using templ.JSFuncCall (Recommended)

Use `templ.JSFuncCall` to safely pass data to JavaScript functions:

```go
templ alertComponent(message string) {
    <button onclick={ templ.JSFuncCall("alert", message) }>Show Alert</button>
}
```

This automatically JSON-encodes the data and generates:
```html
<button onclick="alert('Hello World')">Show Alert</button>
```

### Method 2: Using Data Attributes (Recommended for Complex Data)

Pass data through HTML attributes using `templ.JSONString`:

```go
type UserData struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}

templ userCard(user UserData) {
    <div class="user-card" data-user={ templ.JSONString(user) }>
        <button onclick="showUserInfo(this.parentElement)">Show Info</button>
    </div>
    <script>
        function showUserInfo(element) {
            const userData = JSON.parse(element.getAttribute('data-user'));
            alert(`Name: ${userData.name}, Age: ${userData.age}`);
        }
    </script>
}
```

### Method 3: Using Script Elements for Data

Pass data in dedicated script elements:

```go
templ dataComponent(data interface{}) {
    @templ.JSONScript("myData", data)
    <script>
        const data = JSON.parse(document.getElementById('myData').textContent);
        console.log(data);
    </script>
}
```

### Method 4: Interpolating Data in Script Tags

You can interpolate Go data directly in script tags:

```go
templ scriptWithData(message string, count int) {
    <script>
        const message = {{ message }};  // JSON encoded
        const count = {{ count }};      // JSON encoded
        console.log(`Message: ${message}, Count: ${count}`);
    </script>
}
```

## Event Handling Best Practices

### ❌ WRONG: Using onclick with Complex Expressions

```go
// DON'T DO THIS - Will cause compilation errors
templ badExample(id string) {
    <button onclick={ "loadPreset('" + id + "')" }>Load</button>
}
```

### ✅ CORRECT: Using templ.JSFuncCall

```go
templ goodExample(id string) {
    <button onclick={ templ.JSFuncCall("loadPreset", id) }>Load</button>
}
```

### ✅ BETTER: Using Data Attributes and Event Listeners

```go
var eventHandle = templ.NewOnceHandle()

templ bestExample(id string) {
    @eventHandle.Once() {
        <script>
            document.addEventListener('DOMContentLoaded', function() {
                document.querySelectorAll('[data-preset-id]').forEach(button => {
                    button.addEventListener('click', function() {
                        const presetId = this.getAttribute('data-preset-id');
                        loadPreset(presetId);
                    });
                });
            });
        </script>
    }
    <button data-preset-id={ id } class="preset-button">Load Preset</button>
}
```

### Passing Event Objects

To pass event objects to functions, use `templ.JSExpression`:

```go
templ eventExample() {
    <script>
        function handleClick(event, message) {
            console.log(message);
            event.preventDefault();
        }
    </script>
    <button onclick={ templ.JSFuncCall("handleClick", templ.JSExpression("event"), "Hello!") }>
        Click me
    </button>
}
```

## Common Mistakes and Solutions

### Mistake 1: String Concatenation in onclick Attributes

❌ **Wrong:**
```go
<button onclick={ "myFunction('" + value + "')" }>Button</button>
```

✅ **Correct:**
```go
<button onclick={ templ.JSFuncCall("myFunction", value) }>Button</button>
```

### Mistake 2: Using templ.SafeScript for onclick

❌ **Wrong:**
```go
<button onclick={ templ.SafeScript("myFunction('" + value + "')") }>Button</button>
```

This causes compilation errors because `onclick` expects a `templ.ComponentScript`, not a string.

### Mistake 3: Not Escaping User Data

❌ **Wrong:**
```go
templ dangerousExample(userInput string) {
    <script>
        const input = "{{ userInput }}"; // Not safe if userInput contains quotes
    </script>
}
```

✅ **Correct:**
```go
templ safeExample(userInput string) {
    <script>
        const input = {{ userInput }}; // JSON encoded automatically
    </script>
}
```

### Mistake 4: Global Variable Pollution

❌ **Wrong:**
```go
templ badGlobals() {
    <script>
        var myData = 'some data'; // Pollutes global scope
        function handleClick() {
            console.log(myData);
        }
    </script>
}
```

✅ **Correct:**
```go
templ goodIIFE() {
    <script>
        (function() {
            var myData = 'some data'; // Scoped to IIFE
            function handleClick() {
                console.log(myData);
            }
            // Expose only what's needed
            window.handleClick = handleClick;
        })();
    </script>
}
```

## Advanced Patterns

### Component-Scoped JavaScript

Use IIFEs (Immediately Invoked Function Expressions) to scope JavaScript to components:

```go
templ advancedComponent(componentId string, data interface{}) {
    <div id={ componentId } class="advanced-component" data-config={ templ.JSONString(data) }>
        <button class="action-button">Action</button>
    </div>
    <script>
        (function() {
            const componentEl = document.getElementById({{ componentId }});
            const config = JSON.parse(componentEl.getAttribute('data-config'));
            
            componentEl.querySelector('.action-button').addEventListener('click', function() {
                console.log('Action triggered with config:', config);
            });
        })();
    </script>
}
```

### Using External Libraries

Include external libraries and use them safely:

```go
var chartHandle = templ.NewOnceHandle()

templ chartComponent(data []DataPoint) {
    @chartHandle.Once() {
        <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    }
    <canvas id="myChart"></canvas>
    <script>
        (function() {
            const ctx = document.getElementById('myChart').getContext('2d');
            new Chart(ctx, {
                type: 'bar',
                data: {{ data }}, // JSON encoded
                options: {
                    responsive: true
                }
            });
        })();
    </script>
}
```

### Module Pattern for Complex Components

For complex components, use the module pattern:

```go
var moduleHandle = templ.NewOnceHandle()

templ complexComponent(config ComponentConfig) {
    @moduleHandle.Once() {
        <script>
            window.MyModule = (function() {
                const modules = new Map();
                
                return {
                    init: function(elementId, config) {
                        const element = document.getElementById(elementId);
                        modules.set(elementId, {
                            element: element,
                            config: config,
                            // Component logic here
                        });
                    },
                    
                    destroy: function(elementId) {
                        modules.delete(elementId);
                    }
                };
            })();
        </script>
    }
    
    <div id={ config.ID } class="complex-component">
        <!-- Component content -->
    </div>
    
    <script>
        MyModule.init({{ config.ID }}, {{ config }});
    </script>
}
```

## Security Considerations

### 1. Always Use JSON Encoding for User Data

```go
// Safe - data is JSON encoded
templ safeComponent(userData UserData) {
    <script>
        const user = {{ userData }};
    </script>
}
```

### 2. Be Careful with templ.JSExpression

Only use `templ.JSExpression` for trusted, compile-time constants:

```go
// Safe - compile-time constant
templ safeExpression() {
    <button onclick={ templ.JSFuncCall("handleClick", templ.JSExpression("event")) }>
        Click
    </button>
}

// Dangerous - runtime user input
templ dangerousExpression(userInput string) {
    <button onclick={ templ.JSFuncCall("handleClick", templ.JSExpression(userInput)) }>
        Click
    </button>
}
```

### 3. Validate Data on Both Sides

Always validate data on both the server (Go) and client (JavaScript) sides:

```go
templ validatedComponent(input string) {
    // Validate in Go
    if len(input) > 100 {
        // Handle error
    }
    
    <script>
        function processInput(input) {
            // Validate in JavaScript too
            if (typeof input !== 'string' || input.length > 100) {
                console.error('Invalid input');
                return;
            }
            // Process input
        }
        
        processInput({{ input }});
    </script>
}
```

## Quick Reference

| Task | Recommended Method | Example |
|------|-------------------|---------|
| Simple function call | `templ.JSFuncCall` | `onclick={ templ.JSFuncCall("myFunc", arg) }` |
| Pass complex data | Data attributes + `templ.JSONString` | `data-config={ templ.JSONString(config) }` |
| Pass event objects | `templ.JSExpression("event")` | `templ.JSFuncCall("func", templ.JSExpression("event"))` |
| One-time scripts | `templ.OnceHandle` | `@handle.Once() { <script>...</script> }` |
| Component data | `templ.JSONScript` | `@templ.JSONScript("id", data)` |
| Interpolate in scripts | `{{ value }}` | `const x = {{ value }};` |

## Troubleshooting

### Error: "cannot use string as templ.ComponentScript"

This happens when you try to use string concatenation in `onclick` attributes. Use `templ.JSFuncCall` instead.

### Error: "pattern docs/*: no matching files found"

When using `go:embed`, make sure the embedded files are in a subdirectory relative to the Go file containing the embed directive.

### Scripts Not Working

1. Check browser console for JavaScript errors
2. Ensure scripts are loaded before use (consider using `DOMContentLoaded`)
3. Verify that data passed from Go is valid JSON
4. Use browser developer tools to inspect generated HTML

This guide should help you write robust, secure templ templates with JavaScript integration while avoiding common pitfalls.
