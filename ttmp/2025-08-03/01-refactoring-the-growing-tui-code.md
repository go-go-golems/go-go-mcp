# Refactoring Plan for go-go-mcp/pkg/ui/tui

**Date:** 2025-08-03  
**Status:** Draft  
**Scope:** Comprehensive refactoring of the TUI codebase to address maintainability, testability, and performance concerns

## Problem Statement

The current `model.go` file has grown to over 1,000 lines and exhibits several code smells:

- **Mega-file smell**: Single file handles key mappings, messages, model state, update logic, domain commands, and parsing helpers
- **Monolithic Update method**: ~450 lines in one switch statement, difficult to reason about and extend
- **Mixed concerns**: Domain logic (file I/O, config parsing) mixed with UI logic (rendering, key handling)
- **Duplicate code paths**: Repeated `if m.configType == ConfigTypeProfile` blocks throughout
- **State management issues**: Value receivers cause unnecessary copying, mutable pointers in commands
- **Type safety concerns**: Repeated type assertions that can panic, inconsistent error handling

## Goal

Transform the current monolithic codebase into a maintainable, testable, and performant TUI layer while keeping day-to-day development unblocked. The plan is split into 7 focused phases that can be shipped incrementally.

---

## PHASE 0 – Baseline & Safety Nets (½ day)

**Objective:** Establish safety measures before refactoring

### Tasks
- [ ] Create a `refactor` branch
- [ ] Add CI workflow that runs:
  - `go vet`
  - `go test ./...`
  - `go test ./... -run=^$ -bench=.`
- [ ] Freeze generated binaries to enable "before vs after" testing
- [ ] Document current behavior with manual test cases

**Outcome:** Safety net in place to catch regressions during refactoring

---

## PHASE 1 – File Structure & Naming (1 day)

**Objective:** Break the monolithic file into focused modules

### Target Structure
```
pkg/ui/tui/
├── keymap.go          # key bindings and help
├── messages.go        # BubbleTea Msg types
├── model/
│   ├── root.go        # RootModel: high-level mode switching
│   ├── list.go        # ListModel: server/profile lists
│   ├── form.go        # ServerFormModel
│   ├── profile.go     # ProfileFormModel
│   └── confirm.go     # ConfirmModel
├── cmds.go            # BubbleTea Cmd helpers (IO, persistence)
├── view/              # pure rendering helpers
└── state/             # domain state structs
```

### Implementation Steps
1. `git mv model.go → tmp/model_legacy.go` to keep compilation passing
2. Add empty files with package declarations; commit
3. Gradually move type declarations to new files (no logic changes):
   - KeyMap + constructor → `keymap.go`
   - All Msg structs → `messages.go`
   - ConfirmModel, FormModel, etc. → `model/*.go`
4. Run `goimports` and `go vet`

**Outcome:** Identical behavior with dramatically better code navigation

---

## PHASE 2 – Separate UI vs Domain Concerns (2 days)

**Objective:** Extract domain logic from UI concerns

### Problem
Current `model.go` mixes:
- Pure data manipulation (load/save configs via Viper, file I/O)
- UI state (cursor position, selected item)

### Solution
Introduce domain service layer under `pkg/core`:

```
core/configstore/
├── claude.go
├── cursor.go
├── amp.go
└── profiles.go
```

Each file exposes a common interface:

```go
type Store interface {
    Load() (State, error)
    Save(State) error
    Delete(name string) error
    List() ([]Item, error)
    Get(id string) (Item, error)
}
```

### Implementation
- `State` is domain-only struct (no lipgloss imports, no key bindings)
- Persistence (Viper/JSON) lives **only** in configstore
- TUI layer talks to `configstore.Store` through BubbleTea Cmds in `cmds.go`:

```go
func LoadServersCmd(s configstore.Store) tea.Cmd
func SaveServerCmd(s configstore.Store, srv types.CommonServer) tea.Cmd
```

**Benefits:**
- IO & business rules are unit-testable without BubbleTea
- TUI becomes a thin orchestrator
- Clear separation of concerns

---

## PHASE 3 – State Management Refinement (1 day)

**Objective:** Simplify state management and mode switching

### New Architecture
```go
type RootModel struct {
    active Child    // interface { Init() tea.Cmd; Update(tea.Msg) (Child, tea.Cmd); View() string }
}
```

### Changes
- Each mode (menu, list, form, confirm) implements `Child` interface
- Mode switching becomes: `m.active = newListModel(...)`
- Remove `mode int` enum and scattered switches
- Each sub-model handles its own keys
- RootModel forwards events and renders global help footer

### Model Ownership
- `ServerListModel` owns: `items []ServerItem`, `selected int`, `clipboard *types.CommonServer`
- `ServerFormModel` owns its `focusIndex`, validation state
- No shared mutable state between models

**Outcome:** Cleaner state transitions, easier to reason about

---

## PHASE 4 – Error Handling Consistency (½ day)

**Objective:** Standardize error handling patterns

### Conventions
- Every IO-related Cmd returns either success-specific Msg or `errorMsg`
- TUI displays errors in fixed status bar with consistent styling:
  ```go
  lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(err)
  ```
- Domain layer never logs; it returns errors with context
- UI layer logs with zerolog only when debugging (`if debugEnabled`)

### Implementation
- Enforce with `go vet` + staticcheck rules
- Add error message timeout/dismissal logic
- Consistent error wrapping with `fmt.Errorf("operation failed: %w", err)`

**Outcome:** Predictable error handling, better user experience

---

## PHASE 5 – Type Safety Improvements (1 day)

**Objective:** Eliminate runtime errors and improve type safety

### Changes

1. **Environment Variables**
   ```go
   // Before: map[string]string
   type EnvVar struct{ Key, Value string }
   type Env []EnvVar
   ```
   Benefits: deterministic ordering, easier diff, validation

2. **Strong Types**
   ```go
   type ServerID string
   type ServerKind int
   
   const (
       KindCmd ServerKind = iota
       KindSSE
   )
   ```

3. **Safe Type Assertions**
   ```go
   // Before: selectedItem := m.activeList.SelectedItem().(serverItem)
   selectedItem, ok := m.activeList.SelectedItem().(serverItem)
   if !ok {
       return m, errorCmd("invalid selection")
   }
   ```

4. **Wrapped Viper Access**
   - Never use `viper.GetString("...")` outside configstore
   - All config access goes through typed interfaces

### Enforcement
- Add `go vet -tags=enforcealias` to prevent accidental string/alias assignments
- Use `staticcheck` to catch unsafe type assertions

**Outcome:** Fewer runtime panics, clearer intent

---

## PHASE 6 – Performance & UX Polish (optional, ½ day)

**Objective:** Optimize performance and user experience

### Optimizations

1. **Cached Collections**
   - Replace repeated `sort` + `map→slice` conversions
   - Cache sorted slice, invalidate only on mutation

2. **Lazy Loading**
   - Load big config files only on first expansion of that list
   - Show loading indicators for slow operations

3. **List Rendering**
   - Use `list.NewDelegate` with `HeightDelegate` optimization
   - Avoid rendering off-screen rows

4. **Memory Efficiency**
   - Pass `*types.CommonServer` where safe to avoid copying
   - Use pointer receivers for large models

**Outcome:** Faster startup, smoother interactions

---

## PHASE 7 – Testing Strategy (1-2 days)

**Objective:** Establish comprehensive testing

### Unit Tests (Domain)
```
core/configstore/*_test.go
```
- Test CRUD operations against temp directory
- Test migration logic for version changes
- Test error conditions and edge cases

### Unit Tests (TUI Sub-models)
Use Charm's BubbleTea programmatic driver:

```go
func TestListNavigation(t *testing.T) {
    p := tea.NewProgram(NewListModel(...), tea.WithoutRenderer())
    p.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
    // assert state changes
}
```

### Integration Tests
- Use BubbleTea testing harness
- Snapshot tests with `termdash` screenshots
- End-to-end scenarios (add, edit, delete, yank/paste)

### CI Setup
- Run `go test ./...` in GitHub Actions
- Add `golangci-lint` with strict settings
- Benchmark tests to catch performance regressions

**Outcome:** Confidence for future changes, regression prevention

---

## Roll-out Sequence & Risk Mitigation

### Incremental Approach
1. After each phase, merge back to main behind feature flags
2. Keep CLI functional throughout refactoring
3. Only delete legacy monolith after Phase 4 passes all tests
4. Maintain public API stability until Phase 5

### Risk Mitigation
- Comprehensive manual testing after each phase
- Feature flags to switch between old/new implementations
- Rollback plan if critical issues discovered
- Documentation of breaking changes for Phase 5

---

## Expected Outcomes

### Code Quality
- Model files <200 LOC each, single-responsibility
- Domain logic decoupled from UI → easier to port to Web or future GUI
- Consistent error & state handling
- Stronger type guarantees

### Development Experience
- Easier to add new features (search, bulk-edit, etc.)
- Better IDE navigation and code completion
- Faster compilation due to smaller files
- Clear separation makes onboarding easier

### Testing & Maintenance
- CI + tests give confidence for changes
- Unit tests enable TDD for new features
- Performance benchmarks prevent regressions
- Clear ownership of code sections

---

## Implementation Timeline

| Phase | Duration | Dependencies | Risk Level |
|-------|----------|--------------|------------|
| Phase 0 | 0.5 days | None | Low |
| Phase 1 | 1 day | Phase 0 | Low |
| Phase 2 | 2 days | Phase 1 | Medium |
| Phase 3 | 1 day | Phase 2 | Medium |
| Phase 4 | 0.5 days | Phase 3 | Low |
| Phase 5 | 1 day | Phase 4 | High |
| Phase 6 | 0.5 days | Phase 5 | Low |
| Phase 7 | 1-2 days | All phases | Medium |

**Total Estimated Time:** 7-8 days

## Success Criteria

- [ ] All existing functionality preserved
- [ ] Build time improved by >20%
- [ ] Test coverage >80% for core logic
- [ ] No files >300 lines
- [ ] Zero `go vet` warnings
- [ ] All staticcheck issues resolved
- [ ] Performance benchmarks stable or improved

---

*This refactoring plan ensures the project remains in a compilable and usable state throughout the process, allowing for incremental progress and early feedback.*
