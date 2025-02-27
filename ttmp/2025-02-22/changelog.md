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
- Implemented UI reset to show a waiting message after a resolved action
- Added YAML source display clearing when an action resolves a request
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