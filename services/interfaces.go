package services

import (
	"context"

	testcontainers "github.com/testcontainers/testcontainers-go"
)

// ServiceRunner defines the interface for running a service container.
// Each service (Postgres, Redis, etc.) should implement this interface.
type ServiceRunner interface {
	// Run starts the service container with the given options
	Run(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (testcontainers.Container, error)

	// Name returns the service name (e.g., "postgres", "redis")
	Name() string
}

// Logger defines the interface for structured logging.
// Users can provide their own implementation or use the default logger.
type Logger interface {
	// Debug logs a debug message with key-value pairs
	Debug(msg string, keysAndValues ...interface{})

	// Info logs an info message with key-value pairs
	Info(msg string, keysAndValues ...interface{})

	// Warn logs a warning message with key-value pairs
	Warn(msg string, keysAndValues ...interface{})

	// Error logs an error message with key-value pairs
	Error(msg string, keysAndValues ...interface{})
}

// HealthChecker defines the interface for service health checks.
type HealthChecker interface {
	// Check performs a health check on the container
	Check(ctx context.Context, container testcontainers.Container) error
}

// HealthCheckFunc is a function type that implements HealthChecker.
type HealthCheckFunc func(ctx context.Context, container testcontainers.Container) error

// Check implements the HealthChecker interface.
func (f HealthCheckFunc) Check(ctx context.Context, container testcontainers.Container) error {
	return f(ctx, container)
}
