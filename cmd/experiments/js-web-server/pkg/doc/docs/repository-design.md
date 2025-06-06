# Repository Pattern Design - Future Source Revision System

## Current Implementation

The repository pattern is currently implemented for script executions with the following components:

### Repository Structure

```
internal/repository/
├── interfaces.go     # Repository interfaces
├── models.go        # Data models
└── sqlite.go        # SQLite implementation
```

### Current Repositories

- **ExecutionRepository**: Manages script execution storage with full CRUD operations
- **RepositoryManager**: Coordinates access to all repositories

## Future Extensions for Source Revision System

### 1. Source Code Repository

```go
// SourceRepository manages source code storage and versioning
type SourceRepository interface {
    // Create a new source file
    CreateSource(ctx context.Context, req CreateSourceRequest) (*Source, error)
    
    // Update source file (creates new version)
    UpdateSource(ctx context.Context, id int, req UpdateSourceRequest) (*Source, error)
    
    // Get current version of source
    GetSource(ctx context.Context, id int) (*Source, error)
    
    // Get specific version of source
    GetSourceVersion(ctx context.Context, id int, version int) (*Source, error)
    
    // List all versions of a source file
    ListSourceVersions(ctx context.Context, id int) ([]Source, error)
    
    // List all source files with filtering
    ListSources(ctx context.Context, filter SourceFilter, pagination PaginationOptions) (*SourceQueryResult, error)
    
    // Branch/merge operations for future Git-like functionality
    CreateBranch(ctx context.Context, req CreateBranchRequest) (*Branch, error)
    MergeBranch(ctx context.Context, req MergeBranchRequest) (*MergeResult, error)
}
```

### 2. Source Models

```go
// Source represents a versioned source code file
type Source struct {
    ID          int       `json:"id" db:"id"`
    Name        string    `json:"name" db:"name"`           // File name/path
    Content     string    `json:"content" db:"content"`     // Source code content
    Version     int       `json:"version" db:"version"`     // Version number
    ParentID    *int      `json:"parent_id" db:"parent_id"` // Previous version ID
    BranchID    int       `json:"branch_id" db:"branch_id"` // Branch identifier
    Author      string    `json:"author" db:"author"`       // Author of changes
    Message     string    `json:"message" db:"message"`     // Commit message
    Hash        string    `json:"hash" db:"hash"`           // Content hash for integrity
    Timestamp   time.Time `json:"timestamp" db:"timestamp"`
    Tags        []string  `json:"tags"`                     // Version tags
}

// Branch represents a development branch
type Branch struct {
    ID          int       `json:"id" db:"id"`
    Name        string    `json:"name" db:"name"`
    ParentID    *int      `json:"parent_id" db:"parent_id"` // Parent branch
    HeadVersion int       `json:"head_version" db:"head_version"`
    Author      string    `json:"author" db:"author"`
    Created     time.Time `json:"created" db:"created"`
    Description string    `json:"description" db:"description"`
}
```

### 3. Execution-Source Linking

```go
// ExecutionSourceLink connects executions to their source versions
type ExecutionSourceLink struct {
    ExecutionID int `json:"execution_id" db:"execution_id"`
    SourceID    int `json:"source_id" db:"source_id"`
    Version     int `json:"version" db:"version"`
}

// Enhanced ExecutionRepository with source linking
type ExecutionRepository interface {
    // ... existing methods ...
    
    // Link execution to source version
    LinkExecutionToSource(ctx context.Context, executionID, sourceID, version int) error
    
    // Get source information for execution
    GetExecutionSource(ctx context.Context, executionID int) (*Source, error)
    
    // Get all executions for a source version
    GetSourceExecutions(ctx context.Context, sourceID, version int) ([]ScriptExecution, error)
}
```

### 4. Database Schema Extensions

```sql
-- Source code storage
CREATE TABLE sources (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    content TEXT NOT NULL,
    version INTEGER NOT NULL,
    parent_id INTEGER REFERENCES sources(id),
    branch_id INTEGER NOT NULL DEFAULT 1,
    author TEXT NOT NULL,
    message TEXT,
    hash TEXT NOT NULL,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(name, version, branch_id)
);

-- Branch management
CREATE TABLE branches (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    parent_id INTEGER REFERENCES branches(id),
    head_version INTEGER,
    author TEXT NOT NULL,
    created DATETIME DEFAULT CURRENT_TIMESTAMP,
    description TEXT
);

-- Link executions to source versions
CREATE TABLE execution_source_links (
    execution_id INTEGER REFERENCES script_executions(id),
    source_id INTEGER REFERENCES sources(id),
    version INTEGER NOT NULL,
    
    PRIMARY KEY (execution_id, source_id)
);

-- Version tags
CREATE TABLE source_tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source_id INTEGER REFERENCES sources(id),
    version INTEGER NOT NULL,
    tag_name TEXT NOT NULL,
    created DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(tag_name)
);

-- Indexes for performance
CREATE INDEX idx_sources_name_version ON sources(name, version);
CREATE INDEX idx_sources_branch_version ON sources(branch_id, version);
CREATE INDEX idx_sources_hash ON sources(hash);
CREATE INDEX idx_execution_source_links_execution ON execution_source_links(execution_id);
CREATE INDEX idx_execution_source_links_source ON execution_source_links(source_id, version);
```

### 5. API Extensions

```go
// Enhanced API endpoints for source management
func SourceHandler(repos repository.RepositoryManager) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case "POST":   // Create or update source
        case "GET":    // Get source or list sources
        case "DELETE": // Delete source version
        }
    }
}

// Version management endpoints
func VersionHandler(repos repository.RepositoryManager) http.HandlerFunc {
    // GET /sources/{id}/versions - List all versions
    // GET /sources/{id}/versions/{version} - Get specific version
    // POST /sources/{id}/versions/{version}/tag - Tag a version
}

// Branch management endpoints  
func BranchHandler(repos repository.RepositoryManager) http.HandlerFunc {
    // GET /branches - List branches
    // POST /branches - Create branch
    // POST /branches/{id}/merge - Merge branch
}
```

### 6. Usage Examples

#### Save Executed Code as Source

```javascript
// JavaScript API extension for saving executions as source
app.post("/api/save-as-source", (req, res) => {
    const { executionId, name, message } = req.body;
    
    // This would be implemented in Go backend
    const source = saveExecutionAsSource(executionId, name, message);
    res.json({ success: true, source });
});
```

#### Version Comparison

```javascript
// Compare two versions of source code
app.get("/api/sources/:id/compare/:v1/:v2", (req, res) => {
    const { id, v1, v2 } = req.params;
    const diff = compareSourceVersions(id, v1, v2);
    res.json({ diff });
});
```

#### Execution History for Source

```javascript
// Get all executions for a source file
app.get("/api/sources/:id/executions", (req, res) => {
    const { id } = req.params;
    const executions = getSourceExecutions(id);
    res.json({ executions });
});
```

## Migration Path

### Phase 1 (Current) ✅
- [x] Repository pattern for executions
- [x] Clean separation of data access
- [x] Extensible interface design

### Phase 2 (Next)
- [ ] Source repository implementation
- [ ] Basic versioning (linear history)
- [ ] Source-execution linking
- [ ] Save execution as source feature

### Phase 3 (Future)
- [ ] Branch management
- [ ] Merge operations
- [ ] Conflict resolution
- [ ] Git-like operations (diff, blame, etc.)

### Phase 4 (Advanced)
- [ ] Distributed repositories
- [ ] Remote synchronization
- [ ] Advanced merge strategies
- [ ] Integration with external VCS

## Benefits of Repository Pattern

1. **Modularity**: Each repository handles one domain (executions, sources, etc.)
2. **Testability**: Easy to mock repositories for unit testing
3. **Flexibility**: Can swap implementations (SQLite → PostgreSQL → MongoDB)
4. **Extensibility**: New features don't break existing code
5. **Clean Architecture**: Business logic separated from data access
6. **Future-Proof**: Ready for complex version control features

## Current Repository Manager Extension

```go
type RepositoryManager interface {
    Executions() ExecutionRepository
    
    // Future repositories
    Sources() SourceRepository
    Branches() BranchRepository
    Tags() TagRepository
    
    Close() error
}
```

This design provides a solid foundation for evolving the JavaScript playground into a full-featured source code management system while maintaining backward compatibility and clean architecture principles.
