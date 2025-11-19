package services

import (
	testcontainers "github.com/testcontainers/testcontainers-go"
)

// Config holds the configuration for a single service.
type Config struct {
	// HealthCheck is an optional health check function to verify service readiness
	HealthCheck HealthChecker

	// Opts are testcontainers options passed to the service runner
	Opts []testcontainers.ContainerCustomizer

	// Dependencies is a list of service names that must be started before this service.
	// Example: []string{"postgres", "redis"}
	Dependencies []string

	// Priority controls the order of service startup (lower starts first).
	// Services with the same priority start in parallel.
	// Default: 0
	Priority int

	// Enabled determines whether the service should be started
	Enabled bool
}

// ManagerConfig holds the configuration for the service manager.
type ManagerConfig struct {
	// Logger is the logger to use. If nil, a default logger will be used.
	Logger Logger

	// MaxParallel is the maximum number of services to start in parallel.
	// Default: 10
	MaxParallel int

	// StopOnError determines whether to stop all services if one fails to start.
	// Default: true
	StopOnError bool
}

// DefaultManagerConfig returns a ManagerConfig with sensible defaults.
func DefaultManagerConfig() ManagerConfig {
	return ManagerConfig{
		MaxParallel: 10,
		Logger:      NewDefaultLogger(),
		StopOnError: true,
	}
}

// ServicesMap is a map of service names to their configurations.
type ServicesMap map[string]Config //nolint:revive // ServicesMap is intentionally named to be descriptive

// NewServicesMap creates an empty services map.
func NewServicesMap() ServicesMap {
	return make(ServicesMap)
}

// Add adds a service configuration to the map.
func (m ServicesMap) Add(name string, cfg Config) ServicesMap { //nolint:gocritic // Config is passed by value intentionally for immutability
	m[name] = cfg
	return m
}

// Enable enables a service with default configuration.
func (m ServicesMap) Enable(name string, opts ...testcontainers.ContainerCustomizer) ServicesMap {
	m[name] = Config{
		Enabled: true,
		Opts:    opts,
	}
	return m
}

// EnableWithPriority enables a service with a specific priority.
func (m ServicesMap) EnableWithPriority(name string, priority int, opts ...testcontainers.ContainerCustomizer) ServicesMap {
	m[name] = Config{
		Enabled:  true,
		Priority: priority,
		Opts:     opts,
	}
	return m
}
