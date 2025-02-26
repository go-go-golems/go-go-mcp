# UI Update Command

This directory contains a shell command for dynamically updating the UI components in the Go-Go-MCP UI server.

## Update UI Command

The `update-ui.yaml` command allows you to update UI components dynamically by sending a JSON payload to the UI server's update endpoint. The components are specified as a YAML string and will be converted to JSON before being sent to the server.

### Usage

```bash
go-go-mcp run-command examples/ui/update-ui.yaml \
  --components 'components:
  - title:
      content: "Hello World"
      id: main-title' \
  --server_url "http://localhost:8080" \
  --verbose
```

### Parameters

- `components` (required): YAML string containing UI components definition
- `server_url` (optional): URL of the UI server (default: "http://localhost:8080")
- `verbose` (optional): Show verbose output (default: false)

### Using a YAML File

You can also use a YAML file as input:

```bash
# Read YAML from a file
YAML_CONTENT=$(cat example-form.yaml)

# Pass the YAML content to the command
go-go-mcp run-command examples/ui/update-ui.yaml \
  --components "$YAML_CONTENT" \
  --verbose
```

Or use the provided script:

```bash
./update-dino-form.sh
```

## UI DSL Reference

The UI DSL (Domain Specific Language) is a YAML-based language for defining user interfaces declaratively. It allows you to create rich, interactive web interfaces without writing HTML directly.

### Basic Structure

Every UI definition consists of a list of components under the `components` key:

```yaml
components:
  - componentType:
      property1: value1
      property2: value2
```

### Common Properties

All components support these common properties:

- `id`: Unique identifier for the component (required)
- `disabled`: Boolean to disable the component (optional)
- `data`: Map of data attributes (optional)

### Component Types

#### Button
```yaml
- button:
    text: "Click me"
    type: primary  # primary, secondary, danger, success
    id: submit-btn
    disabled: false
```

#### Title (H1 Heading)
```yaml
- title:
    content: "Welcome to My App"
    id: main-title
```

#### Text (Paragraph)
```yaml
- text:
    content: "This is a paragraph of text."
    id: description-text
```

#### Input Field
```yaml
- input:
    type: text  # text, email, password, number, tel
    placeholder: "Enter your name"
    value: ""
    required: true
    id: name-input
```

#### Textarea
```yaml
- textarea:
    placeholder: "Enter description"
    id: description-textarea
    rows: 4
    cols: 50
    value: |
      Default multiline
      text content
```

#### Checkbox
```yaml
- checkbox:
    label: "Accept terms"
    checked: false
    required: true
    name: terms
    id: terms-checkbox
```

#### List

Lists can contain various types of nested components, supporting both ordered (ol) and unordered (ul) lists.

```yaml
- list:
    type: ul  # ul or ol
    title: "Shopping List"
    items:
      - text:
          content: "Groceries to buy:"
      - checkbox:
          label: "Milk"
          id: milk-checkbox
      - checkbox:
          label: "Bread"
          id: bread-checkbox
      - button:
          text: "Add Item"
          id: add-item-btn
          type: secondary
```

#### Form
```yaml
- form:
    id: signup-form
    components:
      - title:
          content: "Sign Up"
      - input:
          type: email
          placeholder: "Email address"
          required: true
      - button:
          id: submit
          text: "Submit"
          type: primary
```

### Complete Example

Here's a complete example of a todo list interface:

```yaml
components:
  - title:
      content: What would you like to tackle next?
  
  - text:
      content: I see you have several items that need attention.
  
  - list:
      type: ul
      items:
        - Review Dependencies:
            button:
              id: review-deps-btn
              text: Review Update (#316)
              type: secondary
        - Calendar Integration:
            button:
              id: review-calendar-btn
              text: Review Calendar PR (#315)
              type: primary
  
  - form:
      id: task-input-form
      components:
        - title:
            content: Add New Task
        - input:
            id: new-task-input
            type: text
            placeholder: What needs to be done?
            required: true
        - checkbox:
            id: high-priority-check
            label: High Priority
        - button:
            id: add-task-btn
            text: Add Task
            type: success
```

### Event Handling

The UI components automatically handle various events:

1. **clicked**: When a user clicks on a button or interactive element
2. **changed**: When an input, textarea, or checkbox value changes
3. **submitted**: When a form is submitted
4. **focused**: When an input receives focus (debug level only)
5. **blurred**: When an input loses focus (debug level only)

### Form Data Collection

When a form is submitted, the system collects data from:

1. Standard form fields via FormData API
2. Checkbox states (including unchecked boxes)
3. All input values by ID (as a fallback)
4. The ID of the button that triggered the submission

### Best Practices

1. Always provide meaningful IDs for components that need to be referenced
2. Use semantic naming for form fields
3. Group related components inside forms
4. Use appropriate button types for different actions:
   - `primary`: Main actions
   - `secondary`: Alternative actions
   - `danger`: Destructive actions
   - `success`: Confirmation actions
5. Provide clear labels and placeholders for form inputs

## Example YAML for update-ui Command

Here's an example of a contact form in YAML format:

```yaml
components:
  - title:
      content: Contact Form
      id: contact-title
  
  - form:
      id: contact-form
      components:
        - input:
            type: text
            placeholder: Your Name
            required: true
            id: name-input
        
        - input:
            type: email
            placeholder: Your Email
            required: true
            id: email-input
        
        - textarea:
            placeholder: Your Message
            rows: 4
            id: message-textarea
        
        - checkbox:
            label: Subscribe to newsletter
            id: newsletter-checkbox
        
        - button:
            text: Send Message
            type: primary
            id: send-btn
```

This will create a contact form with name, email, message fields, a newsletter checkbox, and a submit button. 