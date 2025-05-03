package session

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type SessionID string

// SessionState holds arbitrary data for a session.
type SessionState map[string]interface{}

// Session represents a client session.
type Session struct {
	ID    SessionID
	State SessionState
	mu    sync.RWMutex
	// TODO: Add metadata like CreationTime, LastAccessTime
}

// contextKey is an unexported type for context keys defined in this package.
// This prevents collisions with keys defined in other packages.
type contextKey string

// sessionContextKey is the key used to store the Session in the context.
const sessionContextKey = contextKey("session")

// WithSession returns a new context derived from ctx that carries the provided session.
func WithSession(ctx context.Context, session *Session) context.Context {
	return context.WithValue(ctx, sessionContextKey, session)
}

// GetSessionFromContext retrieves the Session stored in the context, if any.
func GetSessionFromContext(ctx context.Context) (*Session, bool) {
	session, ok := ctx.Value(sessionContextKey).(*Session)
	return session, ok
}

func NewSession() *Session {
	return &Session{
		ID:    SessionID(uuid.New().String()),
		State: make(SessionState),
	}
}

// SetData stores a key-value pair in the session's state.
// It is thread-safe.
func (s *Session) SetData(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.State == nil {
		s.State = make(SessionState)
	}
	s.State[key] = value
	log.Debug().Str("sessionID", string(s.ID)).Str("key", key).Msg("Set data in session")
}

// GetData retrieves a value from the session's state by key.
// The boolean return value indicates whether the key was found.
// It is thread-safe.
func (s *Session) GetData(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.State == nil {
		return nil, false
	}
	value, ok := s.State[key]
	if ok {
		log.Debug().Str("sessionID", string(s.ID)).Str("key", key).Msg("Get data from session")
	} else {
		log.Debug().Str("sessionID", string(s.ID)).Str("key", key).Msg("Data not found in session")
	}
	return value, ok
}

// DeleteData removes a key-value pair from the session's state.
// It is thread-safe.
func (s *Session) DeleteData(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.State == nil {
		return
	}
	delete(s.State, key)
	log.Debug().Str("sessionID", string(s.ID)).Str("key", key).Msg("Deleted data from session")
}
