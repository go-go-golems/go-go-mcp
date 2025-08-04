# Design for Unified Editor Configuration System

**Date:** 2025-08-04  
**Author:** AI Assistant  
**Status:** Draft for Review

## Problem Statement

Currently, go-go-mcp has separate commands for each editor (Claude, Cursor) that are largely duplicated:
- `mcp claude init/edit/add-mcp-server/remove-mcp-server/list-servers/enable-server/disable-server/tail/add-go-go-server`
- `mcp cursor init/edit/add-mcp-server/add-mcp-server-sse/remove-mcp-server/list-servers/add-go-go-server/enable-server/disable-server`

This creates maintenance overhead and inconsistent UX as new editors are added.

## Current State Analysis

### Existing Commands Functionality
From `claude_config.go`, we have these operations:
1. **init** - Initialize configuration file
2. **edit** - Open in $EDITOR 
3. **add-mcp-server** - Add stdio server with env vars
4. **remove-mcp-server** - Remove server by name
5. **list-servers** - List all servers with status
6. **enable-server** - Enable disabled server
7. **disable-server** - Disable server without removing
8. **tail** - Tail Claude log files (editor-specific)
9. **add-go-go-server** - Shortcut for go-go-mcp servers

### Configstore Abstraction
The existing `pkg/core/configstore` provides unified interfaces:
- **Store interface**: Load/Save/ListServers/AddServer/RemoveServer/EnableServer/DisableServer
- **Factory pattern**: `NewStore(ConfigType)` creates appropriate store
- **Supported types**: Claude, Cursor, AmpCode, Amp, CrushLocal, CrushCwd, CrushGlobal

### Editor Configuration Locations & Targets

| Editor | Target | Config Path | Format |
|--------|--------|-------------|---------|
| claude | default | `~/.config/Claude/claude_desktop_config.json` | JSON |
| cursor | global | `~/.cursor/mcp.json` | JSON |
| cursor | cwd | `./mcp.json` | JSON |
| ampcode | default | `~/.config/Cursor/User/settings.json` | JSONC |
| amp | default | `~/.config/amp/settings.json` | JSONC |
| crush | local | `./.crush.json` | JSON |
| crush | cwd | `./crush.json` | JSON |
| crush | global | `~/.config/crush/crush.json` | JSON |

### Key Differences Between Editors
1. **Transport Support**:
   - Claude: stdio only
   - Cursor: stdio + SSE
   - Crush: stdio + HTTP + SSE (most flexible)
   - Amp/AmpCode: stdio + SSE

2. **Configuration Formats**:
   - Claude/Cursor/Crush: JSON
   - Amp/AmpCode: JSONC (with comments)

3. **Special Features**:
   - Claude: Log tailing functionality
   - Cursor: Project-specific vs global configs
   - Crush: Multiple target locations with precedence
   - Amp/AmpCode: VS Code compatible settings

## Target Design

### Unified Command Structure
Replace editor-specific commands with:
```bash
mcp config <editor> <verb> [options]
```

### Core Verbs
1. **add** - Add MCP server (replaces add-mcp-server/add-mcp-server-sse)
2. **remove** - Remove MCP server 
3. **list** - List servers with status
4. **enable** - Enable disabled server
5. **disable** - Disable server
6. **edit** - Open config in $EDITOR
7. **init** - Initialize config file

### Special Verbs
8. **tail** - Tail logs (Claude only initially)
9. **add-go-go** - Shortcut for go-go-mcp servers

### Argument and Flag Design

#### Core Arguments
- `<editor>`: Editor type (claude|cursor|ampcode|amp|crush)
- `<verb>`: Action to perform

#### Core Flags
- `--target, -t`: Target location (handled by ConfigStore interface)

#### ConfigStore Interface for Target Resolution
The ConfigStore interface is enhanced to handle target resolution internally:

```go
type Store interface {
    // Existing methods...
    Load() error
    Save() error
    ListServers() (map[string]types.CommonServer, error)
    // ... etc
    
    // New method for target resolution
    ResolveTarget(target string) error
}
```

#### Editor/Target Mapping (handled by ConfigStore)
```
claude:
  - default -> ConfigTypeClaude (~/.config/Claude/claude_desktop_config.json)

cursor: 
  - default|global -> ConfigTypeCursor (~/.cursor/mcp.json)
  - cwd -> cursor project config (./mcp.json)

ampcode:
  - default -> ConfigTypeAmpCode (~/.config/Cursor/User/settings.json)

amp:
  - default -> ConfigTypeAmp (~/.config/amp/settings.json)

crush:
  - local -> ConfigTypeCrushLocal (./.crush.json)
  - cwd -> ConfigTypeCrushCwd (./crush.json)  
  - default|global -> ConfigTypeCrushGlobal (~/.config/crush/crush.json)
```

#### Server Configuration Flags
- `--env, -e`: Environment variables (KEY=VALUE)
- `--overwrite, -w`: Overwrite existing server
- `--url`: URL for SSE/HTTP servers
- `--type`: Transport type (stdio|sse|http) - auto-detected if not specified

### File Structure
One file per command verb in `cmd/go-go-mcp/cmds/config/`:
```
cmd/go-go-mcp/cmds/config/
├── config.go          # Main config command group (mcp config <editor>)
├── add.go             # mcp config <editor> add
├── remove.go          # mcp config <editor> remove  
├── list.go            # mcp config <editor> list
├── enable.go          # mcp config <editor> enable
├── disable.go         # mcp config <editor> disable
├── edit.go            # mcp config <editor> edit
├── init.go            # mcp config <editor> init
├── tail.go            # mcp config <editor> tail (claude-specific)
└── add_go_go.go       # mcp config <editor> add-go-go
```

### Implementation Strategy

#### Phase 1: ConfigStore Interface Enhancement
1. Add `ResolveTarget(target string) error` method to Store interface
2. Update all store implementations to handle target resolution internally
3. Create editor argument parsing in main config command

#### Phase 2: Command Implementation  
1. Implement each verb as separate file
2. Leverage existing configstore abstractions with new target resolution
3. Handle editor-specific features gracefully

#### Phase 3: Clean Implementation (No Backwards Compatibility)
1. Remove existing `claude` and `cursor` commands completely
2. Single unified interface from the start
3. Clean command structure without legacy baggage

### Error Handling Strategy

#### Invalid Editor/Target Combinations
- Validate editor/target combinations upfront
- Provide helpful error messages suggesting valid combinations
- Auto-suggest closest valid combination

#### Transport Type Validation
- Validate transport types against editor capabilities
- Auto-detect transport type from URL/command presence
- Provide clear errors for unsupported combinations

### User Experience Considerations

#### Smart Defaults
- `<editor>`: Required argument (no auto-detection to avoid ambiguity)
- `--target`: Default to most common target per editor (handled by ConfigStore)
- Transport type: Auto-detect from presence of URL vs command

#### Tab Completion
- Editor names as arguments
- Target names per editor as flags
- Server names for enable/disable/remove

#### Help Integration
- Editor-specific help showing supported targets
- Transport capability matrix in help
- Examples for each editor type

## Technical Implementation Details

### ConfigStore Factory with Target Resolution
```go
func NewStoreWithTarget(editor, target string) (configstore.Store, error) {
    // Determine base config type from editor
    var baseConfigType configstore.ConfigType
    switch editor {
    case "claude":
        baseConfigType = configstore.ConfigTypeClaude
    case "cursor":
        baseConfigType = configstore.ConfigTypeCursor
    case "ampcode":
        baseConfigType = configstore.ConfigTypeAmpCode
    case "amp":
        baseConfigType = configstore.ConfigTypeAmp
    case "crush":
        baseConfigType = configstore.ConfigTypeCrushGlobal // default
    default:
        return nil, fmt.Errorf("unsupported editor: %s", editor)
    }
    
    store, err := configstore.NewStore(baseConfigType)
    if err != nil {
        return nil, err
    }
    
    // Let the store handle target resolution internally
    if target != "" {
        if err := store.ResolveTarget(target); err != nil {
            return nil, err
        }
    }
    
    return store, nil
}
```

### Enhanced ConfigStore Interface
```go
type Store interface {
    // Existing methods
    Load() error
    Save() error
    ListServers() (map[string]types.CommonServer, error)
    GetServer(name string) (types.CommonServer, bool, error)
    AddServer(server types.CommonServer, overwrite bool) error
    RemoveServer(name string) error
    IsServerDisabled(name string) (bool, error)
    EnableServer(name string) error
    DisableServer(name string) error
    
    // New method for target resolution
    ResolveTarget(target string) error
    GetSupportedTargets() []string
}
```

### Command Factory Pattern
```go
type ConfigCommandOptions struct {
    Editor     string
    Target     string
    ConfigPath string // explicit override
}

func ExecuteWithStore(opts ConfigCommandOptions, fn func(store configstore.Store) error) error {
    store, err := NewStoreWithTarget(opts.Editor, opts.Target)
    if err != nil {
        return err
    }
    
    return fn(store)
}
```

### Transport Type Auto-Detection
```go
func DetectTransportType(server types.CommonServer) string {
    if server.URL != "" {
        if server.IsSSE {
            return "sse"
        }
        return "http"
    }
    return "stdio"
}
```

## Simplified Design (No Backwards Compatibility)

### Key Design Decisions

#### 1. **Command Structure: `mcp config <editor> <verb>` - Clean and Simple**
- Editor as positional argument eliminates ambiguity
- Follows patterns like `git remote add` or `docker context use`
- No auto-detection needed - explicit and clear

#### 2. **Target Resolution in ConfigStore Interface**
- Each store implementation handles its own target resolution logic
- Cleaner separation of concerns
- Eliminates complex validation logic in command layer

#### 3. **No Migration Burden**
- Clean slate implementation without legacy support
- Simpler codebase and maintenance
- Users get immediate benefit of unified interface

### Enhanced Error Handling
```go
func ValidateEditor(editor string) error {
    validEditors := []string{"claude", "cursor", "ampcode", "amp", "crush"}
    
    for _, valid := range validEditors {
        if editor == valid {
            return nil
        }
    }
    
    return fmt.Errorf("editor '%s' not recognized. Available editors: %s", 
        editor, strings.Join(validEditors, ", "))
}
```

## Next Steps

1. ✅ Design completed - simplified approach with no backwards compatibility
2. Enhance ConfigStore interface with `ResolveTarget()` method
3. Implement unified command structure `mcp config <editor> <verb>`
4. Implement each verb as separate file
5. Remove existing editor-specific commands
6. Add comprehensive tests including error cases
7. Update documentation with new unified interface

## Success Criteria

- [ ] Unified command interface for all editors: `mcp config <editor> <verb>`
- [ ] Consistent UX across editor types  
- [ ] Enhanced ConfigStore interface with target resolution
- [ ] Graceful error handling for edge cases
- [ ] Clean implementation without legacy baggage
- [ ] Extensible design for future editors
- [ ] Comprehensive test coverage
- [ ] Updated documentation and examples
