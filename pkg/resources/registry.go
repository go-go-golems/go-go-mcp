package resources

import (
	"context"
	"sort"
	"sync"

	"github.com/go-go-golems/go-go-mcp/pkg"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
)

// Registry provides a simple way to register individual resources
type Registry struct {
	mu        sync.RWMutex
	resources map[string]protocol.Resource
	handlers  map[string]Handler
	// subscribers maps resource URIs to channels that receive update notifications
	subscribers map[string][]chan struct{}
}

var _ pkg.ResourceProvider = &Registry{}

// Handler is a function that provides the content for a resource
type Handler func(resource protocol.Resource) (*protocol.ResourceContent, error)

// NewRegistry creates a new resource registry
func NewRegistry() *Registry {
	return &Registry{
		resources:   make(map[string]protocol.Resource),
		handlers:    make(map[string]Handler),
		subscribers: make(map[string][]chan struct{}),
	}
}

// RegisterResource adds a resource to the registry
func (r *Registry) RegisterResource(resource protocol.Resource) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.resources[resource.URI] = resource
	r.notifySubscribers(resource.URI)
}

// RegisterResourceWithHandler adds a resource with a custom handler
func (r *Registry) RegisterResourceWithHandler(resource protocol.Resource, handler Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.resources[resource.URI] = resource
	r.handlers[resource.URI] = handler
	r.notifySubscribers(resource.URI)
}

// UnregisterResource removes a resource from the registry
func (r *Registry) UnregisterResource(uri string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.resources, uri)
	delete(r.handlers, uri)
	r.notifySubscribers(uri)
}

// ListResources implements ResourceProvider interface
func (r *Registry) ListResources(_ context.Context, cursor string) ([]protocol.Resource, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	resources := make([]protocol.Resource, 0, len(r.resources))
	for _, res := range r.resources {
		resources = append(resources, res)
	}

	sort.Slice(resources, func(i, j int) bool {
		return resources[i].URI < resources[j].URI
	})

	if cursor == "" {
		return resources, "", nil
	}

	pos := -1
	for i, res := range resources {
		if res.URI == cursor {
			pos = i
			break
		}
	}

	if pos == -1 {
		return resources, "", nil
	}

	return resources[pos+1:], "", nil
}

// ReadResource implements ResourceProvider interface
func (r *Registry) ReadResource(_ context.Context, uri string) ([]protocol.ResourceContent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	resource, ok := r.resources[uri]
	if !ok {
		return nil, pkg.ErrResourceNotFound
	}

	if handler, ok := r.handlers[uri]; ok {
		content, err := handler(resource)
		if err != nil {
			return nil, err
		}
		return []protocol.ResourceContent{*content}, nil
	}

	// Return empty content if no handler is registered
	return []protocol.ResourceContent{{
		URI: uri,
	}}, nil
}

// ListResourceTemplates implements ResourceProvider interface
func (r *Registry) ListResourceTemplates(_ context.Context) ([]protocol.ResourceTemplate, error) {
	// This is a basic implementation that returns no templates
	return []protocol.ResourceTemplate{}, nil
}

// SubscribeToResource implements ResourceProvider interface
func (r *Registry) SubscribeToResource(_ context.Context, uri string) (chan struct{}, func(), error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.resources[uri]; !ok {
		return nil, nil, pkg.ErrResourceNotFound
	}

	ch := make(chan struct{}, 1)
	if r.subscribers[uri] == nil {
		r.subscribers[uri] = make([]chan struct{}, 0)
	}
	r.subscribers[uri] = append(r.subscribers[uri], ch)

	// Return cleanup function
	cleanup := func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		for i, sub := range r.subscribers[uri] {
			if sub == ch {
				r.subscribers[uri] = append(r.subscribers[uri][:i], r.subscribers[uri][i+1:]...)
				close(ch)
				break
			}
		}
	}

	return ch, cleanup, nil
}

// notifySubscribers sends notifications to all subscribers of a resource
func (r *Registry) notifySubscribers(uri string) {
	if subs, ok := r.subscribers[uri]; ok {
		for _, ch := range subs {
			select {
			case ch <- struct{}{}:
			default:
				// Channel is full, skip notification
			}
		}
	}
}
