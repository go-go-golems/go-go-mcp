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

HTML Simplification Tool Optimization
Enhanced the HTML simplification tool to produce more compact YAML output while preserving essential information.

- Optimized YAML output format for better readability and reduced size
- Added compact SVG handling to simplify vector graphics output
- Combined attributes into space-separated strings
- Improved text node handling for cleaner output

HTML Simplification SVG Handling
Added option to completely remove SVG elements from the output for simpler text-focused analysis.

- Added --strip-svg flag to remove all SVG elements
- SVG stripping is disabled by default
- Works in conjunction with --compact-svg flag

HTML List and Select Box Compaction
Added option to limit the number of items shown in lists and select boxes for more concise output.

- Added --max-list-items flag to limit items in lists and select boxes (default: 4)
- Automatically adds ellipsis (...) to indicate truncated items
- Applies to ul, ol, and select elements
- Set to 0 for unlimited items

HTML Table Row Limiting and Selector Filtering
Added table row limiting and selector-based filtering capabilities to the HTML simplification tool.

- Added --max-table-rows flag to limit table rows (default: 4)
- Added --config flag to specify a YAML configuration file
- Implemented selector-based filtering using CSS selectors
- Created example configuration file with common use cases
- Table rows and filtered elements are handled consistently with list items

HTML Selector Enhancement
Added XPath selector support to complement existing CSS selector filtering capabilities.

- Added support for XPath selectors in configuration file
- Updated configuration format to specify selector type
- Added validation for selector types
- Updated example configuration with both CSS and XPath examples
- Improved error handling for XPath selector execution

Text Node Simplification
Added option to simplify nodes that contain only text and line breaks into a single text field.

- Added --simplify-text flag (enabled by default)
- Collapses nodes with only text and <br> children into a text field
- Preserves line breaks as newlines in the text
- Works in conjunction with --shorten-text for length limiting
- Reduces output complexity for text-heavy content

Table Row Check Fix
Fixed table row limiting to correctly check child nodes instead of parent nodes.

- Fixed bug where table row check was incorrectly applied to parent nodes
- Improved accuracy of table row counting and limiting
- Ensures correct truncation of table rows 