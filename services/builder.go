package services

import (
	"context"

	testcontainers "github.com/testcontainers/testcontainers-go"
)

// Builder provides a fluent API for configuring services.
type Builder struct {
	services ServicesMap
	config   ManagerConfig
}

// NewBuilder creates a new services builder with default configuration.
func NewBuilder() *Builder {
	return &Builder{
		services: NewServicesMap(),
		config:   DefaultManagerConfig(),
	}
}

// WithLogger sets a custom logger.
func (b *Builder) WithLogger(logger Logger) *Builder {
	b.config.Logger = logger
	return b
}

// WithMaxParallel sets the maximum number of parallel service starts.
func (b *Builder) WithMaxParallel(maxParallel int) *Builder {
	b.config.MaxParallel = maxParallel
	return b
}

// WithStopOnError sets whether to stop all services on error.
func (b *Builder) WithStopOnError(stop bool) *Builder {
	b.config.StopOnError = stop
	return b
}

// WithService adds a service to the builder.
// The service must be registered in DefaultRegistry, otherwise it panics.
//
// Example:
//
//	builder.WithService("postgres", testcontainers.WithImage("postgres:15"))
func (b *Builder) WithService(name string, opts ...testcontainers.ContainerCustomizer) *Builder {
	if !DefaultRegistry.Has(name) {
		panic("service '" + name + "' is not registered in DefaultRegistry")
	}
	b.services[name] = Config{
		Opts: opts,
	}
	return b
}

// WithServices adds multiple services to the builder without options.
// All services must be registered in DefaultRegistry, otherwise it panics.
//
// Example:
//
//	builder.WithServices("postgres", "redis", "clickhouse")
func (b *Builder) WithServices(names ...string) *Builder {
	for _, name := range names {
		b.WithService(name)
	}
	return b
}

// Build creates and returns a new Manager.
func (b *Builder) Build() *Manager {
	return NewManager(b.services, b.config)
}

// BuildAndStart creates a Manager and starts all services.
// This is a convenience method for simple use cases.
func (b *Builder) BuildAndStart(ctx context.Context) (*Manager, error) {
	manager := b.Build()
	if err := manager.Start(ctx); err != nil {
		return nil, err
	}
	return manager, nil
}
