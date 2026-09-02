package embeddable

import (
	"fmt"
	"sort"
)

// Scope is one verified OAuth/MCP authority value. It is distinct from
// application-specific capability and document-access types.
type Scope string

// ScopeSet is a normalized, immutable-by-API set of verified scopes. The zero
// value is empty and grants no authority.
type ScopeSet struct {
	values []Scope
}

// NewScopeSet validates, deduplicates, and sorts scope values.
func NewScopeSet(values ...Scope) (ScopeSet, error) {
	seen := make(map[Scope]struct{}, len(values))
	for _, raw := range values {
		value := Scope(string(raw))
		if !validScopeToken(string(value)) {
			return ScopeSet{}, fmt.Errorf("invalid scope %q", raw)
		}
		seen[value] = struct{}{}
	}
	result := make([]Scope, 0, len(seen))
	for value := range seen {
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return ScopeSet{values: result}, nil
}

// ParseScopeSet converts a decoded raw scope collection at an external
// verifier boundary into typed authority.
func ParseScopeSet(values []string) (ScopeSet, error) {
	scopes := make([]Scope, len(values))
	for i, value := range values {
		scopes[i] = Scope(value)
	}
	return NewScopeSet(scopes...)
}

// validScopeToken implements RFC 6749's scope-token grammar:
// %x21 / %x23-5B / %x5D-7E. Iterating bytes also rejects non-ASCII UTF-8.
func validScopeToken(value string) bool {
	if value == "" {
		return false
	}
	for i := 0; i < len(value); i++ {
		if value[i] == 0x21 || value[i] >= 0x23 && value[i] <= 0x5b || value[i] >= 0x5d && value[i] <= 0x7e {
			continue
		}
		return false
	}
	return true
}

// Contains reports whether the scope is present.
func (s ScopeSet) Contains(scope Scope) bool {
	index := sort.Search(len(s.values), func(i int) bool { return s.values[i] >= scope })
	return index < len(s.values) && s.values[index] == scope
}

// IsSubsetOf reports whether every scope in s is present in other.
func (s ScopeSet) IsSubsetOf(other ScopeSet) bool {
	for _, scope := range s.values {
		if !other.Contains(scope) {
			return false
		}
	}
	return true
}

// Empty reports whether the set contains no authority.
func (s ScopeSet) Empty() bool { return len(s.values) == 0 }

// Values returns a defensive copy in deterministic order.
func (s ScopeSet) Values() []Scope { return append([]Scope(nil), s.values...) }

// Strings returns a defensive representation for protocol codecs.
func (s ScopeSet) Strings() []string {
	values := make([]string, len(s.values))
	for i, scope := range s.values {
		values[i] = string(scope)
	}
	return values
}
