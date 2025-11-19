package services

import (
	"sync"
)

// Registry holds all available service runners.
type Registry struct {
	runners map[string]ServiceRunner
	mu      sync.RWMutex
}

// NewRegistry creates a new service registry.
func NewRegistry() *Registry {
	return &Registry{
		runners: make(map[string]ServiceRunner),
	}
}

// Register registers a service runner.
// Returns an error if the service is already registered.
func (r *Registry) Register(name string, runner ServiceRunner) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.runners[name]; exists {
		return &ErrServiceAlreadyRegistered{ServiceName: name}
	}

	r.runners[name] = runner
	return nil
}

// MustRegister registers a service runner and panics if it fails.
func (r *Registry) MustRegister(name string, runner ServiceRunner) {
	if err := r.Register(name, runner); err != nil {
		panic(err)
	}
}

// Get retrieves a service runner by name.
func (r *Registry) Get(name string) (ServiceRunner, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	runner, ok := r.runners[name]
	return runner, ok
}

// List returns a list of all registered service names.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.runners))
	for name := range r.runners {
		names = append(names, name)
	}
	return names
}

// Has checks if a service is registered.
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.runners[name]
	return ok
}

// Unregister removes a service runner from the registry.
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.runners, name)
}

// DefaultRegistry is the global service registry.
var DefaultRegistry = NewRegistry()

// Register registers a service in the default registry.
func Register(name string, runner ServiceRunner) error {
	return DefaultRegistry.Register(name, runner)
}

// MustRegister registers a service in the default registry and panics if it fails.
func MustRegister(name string, runner ServiceRunner) {
	DefaultRegistry.MustRegister(name, runner)
}
