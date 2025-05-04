package session

import (
	"sync"

	"github.com/google/uuid"
)

// SessionStore defines the interface for managing sessions.
type SessionStore interface {
	Get(sessionID SessionID) (*Session, bool)
	Create() *Session
	Update(session *Session) // Note: For in-memory, Get returns a pointer, so Update might be implicit.
	Delete(sessionID SessionID)
	// TODO: Add cleanup mechanism for stale sessions
}

// InMemorySessionStore provides a thread-safe in-memory implementation of SessionStore.
type InMemorySessionStore struct {
	mu       sync.RWMutex
	sessions map[SessionID]*Session
}

// NewInMemorySessionStore creates a new in-memory session store.
func NewInMemorySessionStore() *InMemorySessionStore {
	return &InMemorySessionStore{
		sessions: make(map[SessionID]*Session),
	}
}

// Get retrieves a session by its ID.
func (s *InMemorySessionStore) Get(sessionID SessionID) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[sessionID]
	return session, ok
}

// Create generates a new session with a unique ID and adds it to the store.
func (s *InMemorySessionStore) Create() *Session {
	s.mu.Lock()
	defer s.mu.Unlock()

	newSession := &Session{
		ID:    SessionID(uuid.NewString()),
		State: make(SessionState),
	}
	s.sessions[newSession.ID] = newSession
	return newSession
}

// Update stores the potentially modified session state.
// For this in-memory store using pointers, changes to the retrieved session
// are inherently reflected. This method ensures the session exists.
func (s *InMemorySessionStore) Update(session *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Just ensure it's in the map, assuming modifications happened on the pointer
	if _, exists := s.sessions[session.ID]; exists {
		s.sessions[session.ID] = session
	}
	// Optionally, add logic here if sessions need explicit saving or validation
}

// Delete removes a session from the store.
func (s *InMemorySessionStore) Delete(sessionID SessionID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, sessionID)
}

// Ensure InMemorySessionStore implements SessionStore
var _ SessionStore = (*InMemorySessionStore)(nil)
