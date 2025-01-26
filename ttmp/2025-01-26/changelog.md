# HTML Selector Testing Tool Documentation

Added detailed technical documentation for implementing the HTML selector testing tool in Go, including core components, data structures, and best practices.

- Added implementation guide for HTML selector testing tool
- Included code examples and component breakdowns
- Added error handling and performance considerations
- Included testing strategy recommendations

# HTML Selector Testing Tool Implementation

Implemented the HTML selector testing tool as a command-line application that helps developers test and verify CSS and XPath selectors against HTML documents.

- Created main CLI application using Cobra
- Implemented CSS and XPath selector support using goquery and htmlquery
- Added YAML configuration and output formatting
- Created example configuration and HTML files for testing
- Added comprehensive README with usage instructions

# HTML Selector Testing Tool Enhancement

Added command-line flag to override input HTML file path, improving flexibility and ease of use.

- Added --input/-i flag to specify HTML file path
- Updated documentation to reflect new usage options
- Made file path in config optional when using command-line flag

# Path Output Enhancement

Enhanced the path output to include element IDs and classes for better specificity and context.

- Updated path generation functions to append IDs and classes
- Improved path readability and accuracy

# Extract Flag Addition

Added an --extract flag to run selectors and return all matches, enhancing the tool's functionality.

- Implemented --extract/-e flag to print all matches for each selector
- Updated README to document the new feature

HTML Simplification Tool Implementation
Added a new command-line tool to simplify and minimize HTML documents, with options to strip scripts, CSS, and shorten text content. The tool outputs a structured YAML representation of the simplified HTML.

- Added `cmd/tools/simplify-html` command with Cobra implementation
- Implemented HTML parsing using goquery
- Added options for script stripping, CSS removal, and text shortening
- Structured YAML output format for processed HTML 