# Configuration File Format Design

Added shell command template generator:
- Created create-command.yaml template for generating shell commands
- Added comprehensive example with log processing
- Included detailed guidelines for command structure
- Added support for Go templates and Sprig functions
- Documented best practices and security considerations

Updated shell commands documentation:
- Restructured documentation to be clearer and more consistent
- Added detailed section about Go templating and Sprig functions
- Improved command structure documentation with better examples
- Added comprehensive parameter type descriptions
- Unified format between tutorial and reference documentation

Added comprehensive shell commands tutorial:
- Created detailed tutorial in pkg/doc/topics/02-shell-commands.md
- Covered basic and advanced shell command usage
- Added real-world examples for Docker and Git operations
- Included best practices and troubleshooting guides
- Documented integration with configuration system

Added documentation for bridge functionality:
- Added section in README about using go-go-mcp as a bridge between SSE and stdio servers
- Documented bridge command usage with example configuration
- Explained use case for stdio-based client integration with web-based MCP servers

Added detailed YAML configuration format design that allows:
- Multiple profiles for different environments
- Tool and prompt source configuration with parameter filtering
- Parka-style parameter management with defaults/overrides and blacklist/whitelist
- Global settings and profile inheritance
- Security considerations through parameter filtering
- Support for directory, file and external command sources 

Added implementation plan that outlines:
- Core type definitions for configuration structures
- Loading strategy with interfaces for config, profile and source management
- Integration approach with existing registry system
- Parameter management system design
- Error handling extensions
- Server integration strategy 

Redesigned implementation to be provider-centric:
- Created ConfigToolProvider and ConfigPromptProvider implementations
- Integrated with Clay's repository system for directory-based tools
- Simplified loading process by using provider constructors
- Moved parameter management into providers
- Removed separate loader interfaces and profile management 

Implemented configuration system:
- Core configuration types in pkg/config/types.go
- Parameter management with filtering and defaults/overrides
- Tool provider with Clay repository integration
- Prompt provider with Pinocchio file support
- Shell command loading from YAML files 

Updated design document to match implementation:
- Aligned YAML structure with actual types
- Added implementation status and TODO items
- Updated provider implementation details
- Added usage examples and server integration details
- Documented key features and design points 

Enhanced tool provider and server integration:
- Refactored ConfigToolProvider to use functional options pattern
- Added support for both directory-based and config-based initialization
- Integrated help system into tool provider
- Updated start command to support configuration files
- Unified shell command loading under ConfigToolProvider 

Improved error handling in tool provider:
- Modified functional options to return errors
- Added proper error handling for profile loading
- Improved error messages for file and directory loading
- Propagated errors from shell command loading
- Removed silent error skipping in favor of explicit error handling 

Added comprehensive documentation:
- Updated design document with error handling details
- Created detailed configuration file tutorial
- Added examples for all major features
- Included troubleshooting guide
- Documented best practices for configuration management 

Added example configuration:
- Created config.yaml in examples directory
- Organized tools into meaningful profiles (default, productivity, research, etc.)
- Added parameter management examples for each profile
- Demonstrated security features with blacklists/whitelists
- Included real-world use cases for different tool combinations 

## Update README with installation instructions and command examples

Updated the README.md to include comprehensive installation instructions matching our other projects' style, and updated command examples to use the installed binary path instead of relative paths.

- Added installation instructions for homebrew, apt-get, yum, and go get
- Added link to GitHub releases
- Updated all command examples to use `go-go-mcp` instead of `./go-go-mcp` 

Added help system documentation to README:
- Added section about configuration file features and help access
- Added section about shell commands features and help access
- Added overview of help system with examples
- Documented --example flag for viewing example configurations
- Added references to `help --all` for discovering all topics 