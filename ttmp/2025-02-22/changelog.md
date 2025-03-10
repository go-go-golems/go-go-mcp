# Improved SSE Implementation for htmx

Updated the SSE implementation to properly support the htmx SSE extension API. This ensures that UI components can be updated in real-time through server-sent events.

- Added component rendering capability to the SSE handler
- Implemented proper event formatting according to htmx SSE extension requirements
- Added support for different event types (component-update, page-reload, yaml-update)
- Added ping events to keep connections alive
- Registered component renderers in the server
- Enhanced page-reload events to also send a full page update for seamless refreshes
- Optimized page reload by extracting page content into a separate template to avoid re-rendering the base template 

# Added UI Action Resolved Status and Auto-Reset

Enhanced the UI action handling to return the resolved status in the response and automatically reset the UI when an action resolves a waiting request.

- Modified the handleUIAction function to include the resolved status in the response
- Updated the sendUIAction JavaScript function to handle resolved actions
- Added automatic form reset after a successful form submission that resolves a request
- Replaced the entire UI with a simplified waiting message after a resolved action
- This improves the user experience by providing immediate feedback and resetting the UI state

# Enhanced Page Reload Events with Page Definitions

Improved the page reload event system to include the full page definition in the event payload.

- Modified `NewPageReloadEvent` to accept and include the page definition
- Updated server code to pass the current page definition when publishing reload events
- This ensures that the SSE handler always has access to the most up-to-date page data
- Improves reliability of page content updates during reload events 

# Added Developer Documentation for SSE Page Handling

Created comprehensive documentation to help new developers understand the SSE and page handling system.

- Added detailed explanation of the SSE architecture and event flow
- Documented the file watching and page loading process
- Explained how templates are selected and rendered
- Described how page data is managed and passed through the system
- Included code examples and a complete walkthrough of the update process
- Created `ttmp/2025-02-22/03-sse-page-handling.md` as a developer guide 

# UI Action Handler with Wait-for-Response

Added a new UI handler implementation that waits for user actions before completing requests. This creates a synchronous flow between UI updates and user interactions, making it easier to build interactive applications.

- Created a new `UIHandler` struct to encapsulate all UI-related functionality
- Implemented a `WaitRegistry` to track pending requests using channels
- Added request ID generation and tracking between UI updates and actions
- Implemented timeout handling for requests without responses
- Integrated the UIHandler with the main server by removing UI-specific handlers from the Server struct
- Updated the Server initialization to create and use the UIHandler
- Added RequestID field to the UIEvent struct for better tracking of user interactions
- Implemented background cleanup for orphaned requests 

# Added Click-Submit Delay for UI Actions

Enhanced the UI handler to wait briefly after receiving a click event to check for a subsequent form submission before responding.

- Added configuration options for enabling/disabling the delay and setting its duration
- Modified the WaitRegistry to keep click events in the registry until explicitly cleaned up
- Implemented a short delay (default 200ms) after click events to wait for potential form submissions
- Added detailed logging for tracking the click-submit sequence
- This improves the user experience by ensuring that form submissions take precedence over simple button clicks 

# Enhanced UI Action Response with Related Events

Improved the UI action response to include related events, allowing clients to receive both click and submission events in a single response.

- Added `RelatedEvents` field to the `UIActionResponse` struct
- Created a new `UIActionEvent` type to represent individual events
- Modified the click-submit delay logic to store click events as related events when a submission follows
- Updated the response JSON format to include the related events array
- This enhancement provides clients with a complete history of user interactions leading to the final action 

# Clarified UI DSL Documentation on Form Submission Model

Updated the UI DSL documentation to clearly explain the single form submission model and its limitations.

- Added a new "Form Submission Model" section to the documentation
- Clarified that the UI system only supports a single response per UI update
- Explained that buttons cannot update state before form submission
- Added examples of incorrect and correct approaches to multi-step interactions
- Provided guidance on designing UIs with separate form submissions for complex interactions
- This helps developers avoid common pitfalls when designing UI components 

# Added SQLite Shell Command Tool

Created a new shell command tool for interacting with SQLite databases, making it easy to execute SQL queries and commands.

- Added `examples/sqlite/sqlite.yaml` with a complete SQLite command implementation
- Implemented support for CSV output format
- Added options for displaying column headers
- Included query timeout configuration
- Added error handling for missing SQLite installation and database files
- Provided detailed examples in the command documentation
- This tool simplifies database operations in projects that use SQLite 

# Enhanced SQLite Tool Security

Improved the SQLite shell command tool to safely handle SQL commands with special characters and prevent shell injection vulnerabilities.

- Modified the tool to write SQL commands to a temporary file instead of passing them directly to the command line
- Added proper cleanup of temporary files using trap
- Improved command execution by using input redirection
- Enhanced logging to show the temporary file path
- This change makes the tool safer for use with complex SQL queries containing quotes, semicolons, or other special characters 

# Added SQLite Session Management

Added SQLite session management to simplify working with databases across multiple commands.

- Created new `sql-open` tool that stores a database filename for subsequent operations
- Modified the `sqlite` tool to make the `db` parameter optional
- Added automatic detection of previously opened databases
- Improved error handling with helpful messages when no database is specified
- Enhanced documentation with examples of session-based workflow
- This improvement streamlines database operations by reducing repetitive parameter entry 

# Enhanced GitHub Issue Comment Tool

Improved the GitHub issue comment tool to safely handle comment bodies with special characters and prevent shell injection vulnerabilities.

- Modified the tool to write comment bodies to a temporary file instead of passing them directly to the command line
- Added proper cleanup of temporary files using trap
- Improved command execution by using the --body-file option
- Enhanced logging with descriptive messages for each operation mode
- Simplified the command logic with a clear if-elif structure
- This change makes the tool safer for use with complex comment bodies containing quotes, newlines, or other special characters 