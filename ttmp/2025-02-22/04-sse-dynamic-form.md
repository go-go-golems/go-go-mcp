---
title: "Understanding UI Action Handling in Go-Go-MCP"
date: "2025-02-22"
author: "Go-Go-MCP Team"
version: "1.0"
status: "current"

# Source code references
code_references:
  - file: "cmd/ui-server/templates.templ"
    description: "UI component templates with event binding"
  - file: "cmd/ui-server/server.go"
    description: "Server-side action handling implementation"
  - file: "cmd/ui-server/static/js/ui-actions.js"
    description: "Client-side JavaScript for action handling"

# Related documentation
related_docs:
  - "pkg/doc/topics/05-ui-dsl.md"
  - "examples/ui-dsl/halloween-party.yaml"
  - "examples/ui-dsl/dino-facts.yaml"

# RAG retrieval metadata
rag_metadata:
  topics:
    - "UI action handling"
    - "Form submission"
    - "Event logging"
    - "HTMX integration"
    - "Client-server communication"
  
  questions:
    - "How does form submission work in Go-Go-MCP?"
    - "How are UI actions logged in the system?"
    - "How do I collect form data in the UI DSL?"
    - "What events are tracked in the UI components?"
    - "How do I handle button clicks in the UI DSL?"
    - "How does the UI action endpoint work?"
    - "How are form submissions processed on the server?"
    - "How do I track which button triggered a form submission?"
    - "How do I implement smart logging for UI actions?"
    - "How do checkboxes work in the UI DSL forms?"
  
  components:
    - "sendUIAction function"
    - "handleUIAction endpoint"
    - "Form data collection"
    - "Event binding"
    - "Smart logging"
    - "Console logging"
    - "Button tracking"
    - "Component rendering"
  
  maintenance_triggers:
    - "Changes to UI component templates"
    - "Modifications to the UI action endpoint"
    - "Updates to form data collection logic"
    - "Changes to event types or logging levels"
    - "Modifications to the client-side JavaScript"
---

# Understanding UI Action Handling in Go-Go-MCP

This guide explains how UI actions and form submissions are handled in our system, including how events are captured, processed, and logged. This document is intended for new developers to understand the architecture and flow of our UI interaction system.

## Overview

Our UI server uses a combination of client-side JavaScript and server-side Go code to handle user interactions with UI components. This creates a responsive, interactive user interface that can track and respond to various user actions such as clicks, form submissions, and input changes.

The system consists of several key components:

1. **Client-Side Event Handlers**: JavaScript functions that capture user interactions
2. **UI Action Endpoint**: A REST API endpoint that receives action data
3. **Server-Side Action Processing**: Go code that processes and logs actions
4. **Form Data Collection**: Special handling for form submissions
5. **Logging System**: Differentiated logging based on action importance

## Client-Side Event Handling

When a user interacts with a UI component, client-side JavaScript captures the event and sends it to the server.

### The `sendUIAction` Function

The core of our client-side event handling is the `sendUIAction` function:

```javascript
function sendUIAction(componentId, action, data = {}) {
    logToConsole(`Component ${componentId} ${action}`);
    
    // If this is a form submission, collect all form data
    if (action === 'submitted' && document.getElementById(componentId)) {
        const form = document.getElementById(componentId);
        const formData = new FormData(form);
        const formValues = {};
        
        // Convert FormData to a plain object
        for (const [key, value] of formData.entries()) {
            formValues[key] = value;
        }
        
        // Add checkbox values (FormData doesn't include unchecked boxes)
        form.querySelectorAll('input[type="checkbox"]').forEach(checkbox => {
            formValues[checkbox.id] = checkbox.checked;
        });
        
        // Add all input values by ID (in case name attributes are missing)
        form.querySelectorAll('input:not([type="checkbox"]), textarea, select').forEach(input => {
            if (input.id) {
                formValues[input.id] = input.value;
            }
        });
        
        // Add the clicked button info if available
        if (form._lastClickedButton) {
            formValues['_clicked_button'] = form._lastClickedButton;
        }
        
        // Merge with any existing data
        data = { ...data, formData: formValues };
        
        // Log form submission for debugging
        console.log('Form submission data:', formValues);
        logToConsole(`Form ${componentId} submitted with data`);
    }
    
    // Send the action to the server
    fetch('/api/ui-action', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            componentId: componentId,
            action: action,
            data: data
        })
    })
    .then(response => response.json())
    .then(data => {
        console.log('Action response:', data);
    })
    .catch(error => {
        console.error('Error sending action:', error);
        logToConsole(`Error sending action: ${error.message}`);
    });
}
```

This function:
1. Logs the action to the console for user feedback
2. For form submissions, collects all form data including:
   - Standard form fields via FormData API
   - Checkbox states (including unchecked boxes)
   - All input values by ID (as a fallback)
   - The ID of the button that triggered the submission
3. Sends the action data to the `/api/ui-action` endpoint
4. Handles the response or any errors

### Component Event Binding

Each UI component is bound to the appropriate event handlers using the `data-hx-on` attribute. For example:

```html
<!-- Button click event -->
<button 
    id="subscribe-btn"
    data-hx-on:click="sendUIAction('subscribe-btn', 'clicked')"
    class="btn btn-success">
    Subscribe
</button>

<!-- Input change event -->
<input
    id="email-input"
    data-hx-on:change="sendUIAction('email-input', 'changed', {value: this.value})"
    type="email"
    class="form-control"
/>

<!-- Form submission event -->
<form
    id="fact-subscription"
    data-hx-on:submit="event.preventDefault(); sendUIAction('fact-subscription', 'submitted')"
    class="needs-validation"
    novalidate
>
    <!-- Form contents -->
</form>
```

### Tracked Event Types

The system tracks several types of events:

1. **clicked**: When a user clicks on a button or interactive element
2. **changed**: When an input, textarea, or checkbox value changes
3. **submitted**: When a form is submitted
4. **focused**: When an input receives focus (debug level only)
5. **blurred**: When an input loses focus (debug level only)

## Server-Side Action Processing

When the server receives an action, it processes it through the `handleUIAction` function:

```go
func (s *Server) handleUIAction() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Only accept POST requests
        if r.Method != http.MethodPost {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }

        // Parse JSON body
        var action struct {
            ComponentID string                 `json:"componentId"`
            Action      string                 `json:"action"`
            Data        map[string]interface{} `json:"data"`
        }

        if err := json.NewDecoder(r.Body).Decode(&action); err != nil {
            http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
            return
        }

        // Determine if this is an important event to log at INFO level
        isImportantEvent := false
        switch action.Action {
        case "clicked", "changed", "submitted":
            isImportantEvent = true
        }

        // Log the action at appropriate level
        logger := log.Debug()
        if isImportantEvent {
            logger = log.Info()
        }

        // Create log entry with component and action info
        logger = logger.
            Str("componentId", action.ComponentID).
            Str("action", action.Action)

        // Add data to log if it exists and is relevant
        if len(action.Data) > 0 {
            // For form submissions, log the form data in detail
            if action.Action == "submitted" && action.Data["formData"] != nil {
                logger = logger.Interface("formData", action.Data["formData"])
            } else if action.Action == "changed" {
                // For changed events, log the new value
                logger = logger.Interface("data", action.Data)
            } else {
                // For other events, just log that data exists
                logger = logger.Bool("hasData", true)
            }
        }

        // Output the log message
        logger.Msg("UI action received")

        // Return success response
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        err := json.NewEncoder(w).Encode(map[string]string{"status": "success"})
        if err != nil {
            http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
        }
    }
}
```

This function:
1. Validates that the request is a POST
2. Parses the JSON body into an action struct
3. Determines the importance of the event for logging purposes
4. Creates a structured log entry with component ID and action type
5. Adds relevant data to the log based on the action type
6. Returns a success response to the client

### Smart Logging

The system uses smart logging to prioritize important events:

```go
// Determine if this is an important event to log at INFO level
isImportantEvent := false
switch action.Action {
case "clicked", "changed", "submitted":
    isImportantEvent = true
}

// Log the action at appropriate level
logger := log.Debug()
if isImportantEvent {
    logger = log.Info()
}
```

This ensures that:
- Important events (clicks, changes, submissions) are logged at INFO level
- Less important events (focus, blur) are logged at DEBUG level
- Logs are kept clean and focused on meaningful interactions

### Form Data Handling

Form submissions receive special handling:

```go
// For form submissions, log the form data in detail
if action.Action == "submitted" && action.Data["formData"] != nil {
    logger = logger.Interface("formData", action.Data["formData"])
}
```

This ensures that all form data is captured and logged, making it easy to debug form submissions and process the data as needed.

## Form Data Collection

Form data collection is a critical part of the system, especially for complex forms.

### How Form Data is Collected

When a form is submitted, the system collects data from multiple sources:

1. **FormData API**: The primary source of form data
   ```javascript
   const formData = new FormData(form);
   const formValues = {};
   
   // Convert FormData to a plain object
   for (const [key, value] of formData.entries()) {
       formValues[key] = value;
   }
   ```

2. **Checkbox Values**: Special handling for checkboxes (which aren't included in FormData when unchecked)
   ```javascript
   form.querySelectorAll('input[type="checkbox"]').forEach(checkbox => {
       formValues[checkbox.id] = checkbox.checked;
   });
   ```

3. **Input Values by ID**: Backup collection for inputs without name attributes
   ```javascript
   form.querySelectorAll('input:not([type="checkbox"]), textarea, select').forEach(input => {
       if (input.id) {
           formValues[input.id] = input.value;
       }
   });
   ```

4. **Clicked Button**: Tracking which button triggered the submission
   ```javascript
   if (form._lastClickedButton) {
       formValues['_clicked_button'] = form._lastClickedButton;
   }
   ```

### Form Button Tracking

To track which button triggered a form submission, we use a click handler that's added during page initialization:

```javascript
document.addEventListener('DOMContentLoaded', (event) => {
    // Add click handlers to all form buttons
    document.querySelectorAll('form button').forEach(button => {
        button.addEventListener('click', function(e) {
            this.form._lastClickedButton = this.id;
        });
    });
});
```

This allows forms with multiple submit buttons to track which one was clicked, providing valuable context for form processing.

## Component Rendering and Event Binding

Each component type has its own rendering logic in the `renderComponent` function, which includes binding the appropriate events.

### Button Component

```go
case "button":
    if id, ok := props["id"].(string); ok {
        <div id={ fmt.Sprintf("component-%s", id) }>
            <button
                id={ id }
                data-hx-on:click={ fmt.Sprintf("sendUIAction('%s', 'clicked')", id) }
                if disabled, ok := props["disabled"].(bool); ok && disabled {
                    disabled="disabled"
                }
                class={
                    "btn",
                    templ.KV("btn-primary", props["type"] == "primary"),
                    templ.KV("btn-secondary", props["type"] == "secondary"),
                    templ.KV("btn-danger", props["type"] == "danger"),
                    templ.KV("btn-success", props["type"] == "success"),
                }
            >
                if text, ok := props["text"].(string); ok {
                    { text }
                }
            </button>
        </div>
    }
```

### Input Component

```go
case "input":
    if id, ok := props["id"].(string); ok {
        <div id={ fmt.Sprintf("component-%s", id) }>
            <input
                id={ id }
                data-hx-on:change={ fmt.Sprintf("sendUIAction('%s', 'changed', {value: this.value})", id) }
                data-hx-on:focus={ fmt.Sprintf("sendUIAction('%s', 'focused')", id) }
                data-hx-on:blur={ fmt.Sprintf("sendUIAction('%s', 'blurred')", id) }
                if typ, ok := props["type"].(string); ok {
                    type={ typ }
                }
                if placeholder, ok := props["placeholder"].(string); ok {
                    placeholder={ placeholder }
                }
                if value, ok := props["value"].(string); ok {
                    value={ value }
                }
                if required, ok := props["required"].(bool); ok && required {
                    required="required"
                }
                if name, ok := props["id"].(string); ok {
                    name={ name }
                }
                class="form-control"
            />
        </div>
    }
```

### Form Component

```go
case "form":
    if id, ok := props["id"].(string); ok {
        <div id={ fmt.Sprintf("component-%s", id) }>
            <form
                id={ id }
                data-hx-on:submit={ fmt.Sprintf("event.preventDefault(); sendUIAction('%s', 'submitted')", id) }
                class="needs-validation"
                novalidate
            >
                if components, ok := props["components"].([]interface{}); ok {
                    for _, comp := range components {
                        if c, ok := comp.(map[string]interface{}); ok {
                            for typ, props := range c {
                                @renderComponent(typ, props.(map[string]interface{}))
                            }
                        }
                    }
                }
            </form>
        </div>
    }
```

## UI Action Endpoint

The UI action endpoint is registered in the server's initialization:

```go
// Set up UI action endpoint
s.mux.Handle("/api/ui-action", s.handleUIAction())
```

This endpoint accepts POST requests with JSON payloads containing:
- `componentId`: The ID of the component that triggered the action
- `action`: The type of action (clicked, changed, submitted, etc.)
- `data`: Optional data associated with the action

### Example Request

```json
{
  "componentId": "fact-subscription",
  "action": "submitted",
  "data": {
    "formData": {
      "email-input": "user@example.com",
      "newsletter-check": true,
      "_clicked_button": "subscribe-btn"
    }
  }
}
```

### Response Format

The endpoint returns a simple JSON response:

```json
{
  "status": "success"
}
```

## Console Logging

In addition to server-side logging, the system provides client-side logging through an interaction console:

```javascript
function logToConsole(message) {
    const console = document.getElementById('interaction-console');
    const entry = document.createElement('div');
    entry.className = 'console-entry';
    entry.textContent = message;
    console.appendChild(entry);
    console.scrollTop = console.scrollHeight;
    if (console.children.length > 50) {
        console.removeChild(console.firstChild);
    }
}
```

This function:
1. Adds a new entry to the interaction console
2. Scrolls to show the latest entry
3. Limits the console to 50 entries to prevent memory issues

The console is displayed at the bottom of every page:

```html
<div class="console-spacer"></div>
<div id="interaction-console"></div>
```

## The Complete Flow

Let's walk through the complete flow of a form submission:

1. A user fills out a form and clicks the submit button
2. The button click sets the `_lastClickedButton` property on the form
3. The form's submit event is captured, and `event.preventDefault()` is called to prevent the default form submission
4. The `sendUIAction` function is called with the form's ID and 'submitted' action
5. The function collects all form data from multiple sources
6. The action data is sent to the `/api/ui-action` endpoint
7. The server receives the request and parses the JSON payload
8. The server determines this is an important event (submitted) and uses INFO level logging
9. The server logs the action with the form data
10. The server sends a success response back to the client
11. The client logs the response to the console

This architecture provides several benefits:
- **Comprehensive Data Collection**: All form data is collected, even from complex forms
- **Smart Logging**: Important events are highlighted in logs
- **Debugging Support**: The interaction console provides immediate feedback
- **Extensibility**: The system can be easily extended to handle new event types or components

## Handling Different Component Types

Different component types have different event handling needs:

| Component | Primary Event | Data Collected |
|-----------|---------------|----------------|
| Button    | clicked       | None           |
| Input     | changed       | Current value  |
| Checkbox  | changed       | Checked state  |
| Textarea  | changed       | Current value  |
| Form      | submitted     | All form data  |
| Text/Title| clicked       | None           |

Each component's events are bound during rendering, ensuring consistent behavior across the application.

## Conclusion

The UI action handling system provides a powerful way to track and respond to user interactions with the UI. By understanding how events are captured, processed, and logged, you can effectively use the system to create interactive, data-driven interfaces.

For further exploration, look at:
- The `renderComponent` function to see how different components are rendered and bound to events
- The `handleUIAction` function to understand how actions are processed on the server
- The JavaScript console to see real-time action data during development 