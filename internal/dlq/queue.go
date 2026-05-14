package dlq

import (
	"context"
	"fmt"
)

// Backend represents a queue backend that can receive retried jobs.
type Backend interface {
	// Name returns the identifier for this backend.
	Name() string
	// Enqueue sends a job payload back to the underlying queue.
	Enqueue(ctx context.Context, job *Job) error
}

// Router maps queue backend names to their Backend implementations.
type Router struct {
	backends map[string]Backend
}

// NewRouter creates a new Router with no registered backends.
func NewRouter() *Router {
	return &Router{
		backends: make(map[string]Backend),
	}
}

// Register adds a backend to the router. Returns an error if a backend
// with the same name is already registered.
func (r *Router) Register(b Backend) error {
	if b == nil {
		return fmt.Errorf("backend must not be nil")
	}
	name := b.Name()
	if name == "" {
		return fmt.Errorf("backend name must not be empty")
	}
	if _, exists := r.backends[name]; exists {
		return fmt.Errorf("backend %q is already registered", name)
	}
	r.backends[name] = b
	return nil
}

// Resolve returns the Backend registered under the given name.
// Returns an error if no backend is found.
func (r *Router) Resolve(name string) (Backend, error) {
	b, ok := r.backends[name]
	if !ok {
		return nil, fmt.Errorf("no backend registered for queue %q", name)
	}
	return b, nil
}

// Names returns the list of all registered backend names.
func (r *Router) Names() []string {
	names := make([]string, 0, len(r.backends))
	for name := range r.backends {
		names = append(names, name)
	}
	return names
}
