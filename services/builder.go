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

// WithService enables a custom service with configuration.
func (b *Builder) WithService(name string, cfg *Config) *Builder {
	b.services.Add(name, *cfg)
	return b
}

// WithServiceSimple enables a custom service with just options.
func (b *Builder) WithServiceSimple(name string, opts ...testcontainers.ContainerCustomizer) *Builder {
	b.services.Enable(name, opts...)
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
