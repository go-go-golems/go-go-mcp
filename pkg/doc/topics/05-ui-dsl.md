---
Title: UI DSL Documentation
Slug: ui-dsl
Short: Learn how to create rich, interactive web interfaces using the YAML-based UI DSL
Topics:
- ui
- dsl
- yaml
- web
Commands:
- none
Flags:
- none
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

# UI DSL Documentation

The UI DSL (Domain Specific Language) is a YAML-based language for defining user interfaces declaratively. It allows you to create rich, interactive web interfaces without writing HTML directly. The DSL is designed to be both human-readable and machine-friendly, making it ideal for both manual creation and automated generation.

## Basic Structure

Every UI definition consists of a list of components under the `components` key:

```yaml
components:
  - componentType:
      property1: value1
      property2: value2
```

## Common Properties

All components support these common properties:

- `id`: Unique identifier for the component (required)
- `disabled`: Boolean to disable the component (optional)
- `data`: Map of data attributes (optional)

## Component Types

### Button
```yaml
- button:
    text: "Click me"
    type: primary  # primary, secondary, danger, success
    id: submit-btn
    disabled: false
```

### Title (H1 Heading)
```yaml
- title:
    content: "Welcome to My App"
    id: main-title
```

### Text (Paragraph)
```yaml
- text:
    content: "This is a paragraph of text."
    id: description-text
```

### Input Field
```yaml
- input:
    type: text  # text, email, password, number, tel
    placeholder: "Enter your name"
    value: ""
    required: true
    id: name-input
```

### Textarea
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

### Checkbox
```yaml
- checkbox:
    label: "Accept terms"
    checked: false
    required: true
    name: terms
    id: terms-checkbox
```

### List

Lists are versatile components that can contain various types of nested components. They support both ordered (ol) and unordered (ul) lists, and can include an optional title. Each list item can be a simple text component or a more complex component like a checkbox, button, or even nested forms.

The list component is particularly useful for:
- Displaying menu options or navigation items
- Creating interactive checklists
- Showing grouped related actions
- Presenting structured content hierarchically

Basic Example:
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

Complex Example with Mixed Components:
```yaml
- list:
    type: ol
    title: "Conference Schedule"
    items:
      - text:
          content: "Morning Sessions:"
      - text:
          content: "9:00 - Keynote Speech"
      - form:
          id: session-1-form
          components:
            - checkbox:
                label: "Attending Keynote"
                id: keynote-attend
            - textarea:
                placeholder: "Your questions"
                id: keynote-questions
                rows: 2
      - text:
          content: "10:30 - Workshop"
      - button:
          text: "Download Schedule"
          id: schedule-btn
          type: primary
```

### Form
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

## Complete Example

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

## Best Practices

1. Always provide meaningful IDs for components that need to be referenced
2. Use semantic naming for form fields
3. Group related components inside forms
4. Use appropriate button types for different actions:
   - `primary`: Main actions
   - `secondary`: Alternative actions
   - `danger`: Destructive actions
   - `success`: Confirmation actions
5. Provide clear labels and placeholders for form inputs
