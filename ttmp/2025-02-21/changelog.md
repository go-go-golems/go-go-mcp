UI Action Handling Documentation Metadata

Enhanced the UI action handling documentation with structured metadata:
- Added YAML preamble with document provenance information
- Included code references to relevant source files
- Added related documentation links
- Created comprehensive RAG metadata for document retrieval
- Added maintenance triggers to identify when document updates are needed
- Structured questions the document answers for better discoverability

UI Action Request ID Tracking

Implemented request ID tracking in UI actions to enable synchronous wait-for-response behavior:
- Updated sendUIAction JavaScript function to include request ID in action data
- Added request ID extraction from UI definition in page templates
- Enhanced UI update template to maintain request ID during SSE updates
- Improved logging to show request ID in console messages
- Added data attributes to store request ID in the DOM
- Ensured proper propagation of request ID between server and client

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

UI Component Action Endpoint

Added a generic event handler system for UI components that sends events to a REST endpoint:
- Created new `/api/ui-action` endpoint to receive component actions
- Added JavaScript function to send component actions to the server
- Updated all UI components to use the new action system
- Actions include component ID, action type, and optional data
- Server logs all actions for monitoring and debugging
- Maintained backward compatibility with existing console logging 

Enhanced UI Action Handling

Improved the UI action system to focus on data-relevant events and provide better form submission data:
- Enhanced form submission to include complete form data in the action payload
- Implemented smart logging that prioritizes data-relevant events (clicked, changed, submitted)
- Added detailed form data logging for form submissions
- Used INFO level for important events and DEBUG level for less important ones
- Improved checkbox handling in form data collection
- Maintained backward compatibility with existing event system

Fixed Form Submission Data Collection

Fixed an issue where form input values weren't being properly collected during form submission:
- Added explicit collection of all input fields by ID during form submission
- Ensured input elements have name attributes matching their IDs
- Simplified form submission handling by consolidating data collection logic
- Added additional logging for form submission data
- Fixed email input value collection in subscription forms

UI Action Handling Documentation

Added comprehensive documentation for the UI action handling system in ttmp/2025-02-22/04-sse-dynamic-form.md:
- Detailed explanation of client-side event handling
- Server-side action processing and smart logging
- Form data collection from multiple sources
- Component rendering and event binding
- Complete flow walkthrough for form submissions
- Component-specific event handling reference
- Console logging and debugging features 