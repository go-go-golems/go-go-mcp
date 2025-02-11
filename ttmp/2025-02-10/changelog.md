# Configuration File Format Design

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