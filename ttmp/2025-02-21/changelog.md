UI DSL Documentation

Added comprehensive documentation for the UI DSL system in pkg/doc/topics/05-ui-dsl.md. The documentation includes:
- Complete schema description
- Component reference
- Examples and best practices
- Styling and JavaScript integration details
- Validation rules and limitations 

UI DSL YAML Display

Enhanced the UI server to display the YAML source of each page alongside its rendered output:
- Added syntax highlighting using highlight.js
- Split view with rendered UI and YAML source side by side
- Improved visual presentation with Bootstrap cards 

UI DSL Interaction Logging

Added an interaction console to display user interactions with the UI:
- Fixed-position console at the bottom of the page
- Logs button clicks, checkbox changes, and form submissions
- Form submissions display data in YAML format
- Limited console history to 50 entries with auto-scroll 

Enhanced List Component Documentation
Added title field to list component and updated documentation to match actual implementation in examples.
- Added title field to list component in UI DSL
- Updated documentation to reflect real-world usage patterns
- Updated halloween-party example to use the new title field
- Added detailed explanatory paragraphs about list component usage
- Added two comprehensive examples showing basic and complex list implementations
- Updated templ templates to render list titles and improve list layout
- Simplified list item rendering logic in templates
- Updated all example files to use consistent list title format
- Improved structure of nested lists and items in examples
- Updated UI DSL schema documentation with new list format
- Converted all remaining examples to use standardized list structure 