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

// NewServicesMap creates a ServicesMap with the specified services.
// All service names must be registered in DefaultRegistry, otherwise it panics.
//
// Example:
//
//	servicesMap := services.NewServicesMap("postgres", "redis")
func NewServicesMap(names ...string) ServicesMap {
	m := make(ServicesMap, len(names))
	for _, name := range names {
		if !DefaultRegistry.Has(name) {
			panic("service '" + name + "' is not registered in DefaultRegistry")
		}
		m[name] = Config{}
	}
	return m
}

// Add adds a service configuration to the map.
func (m ServicesMap) Add(name string, cfg Config) ServicesMap { //nolint:gocritic // Config is passed by value intentionally for immutability
	m[name] = cfg
	return m
}

// WithPriority sets the priority for a service.
// Lower priority services start first. Services with the same priority start in parallel.
func (m ServicesMap) WithPriority(serviceName string, priority int) ServicesMap {
	cfg := m[serviceName]
	cfg.Priority = priority
	m[serviceName] = cfg
	return m
}

// WithOptions adds container options for a service.
func (m ServicesMap) WithOptions(serviceName string, opts ...testcontainers.ContainerCustomizer) ServicesMap {
	cfg := m[serviceName]
	cfg.Opts = append(cfg.Opts, opts...)
	m[serviceName] = cfg
	return m
}

// WithDependencies sets the dependencies for a service.
// Dependencies must be started before this service.
func (m ServicesMap) WithDependencies(serviceName string, deps ...string) ServicesMap {
	cfg := m[serviceName]
	cfg.Dependencies = deps
	m[serviceName] = cfg
	return m
}

// WithHealthCheck sets a health check function for a service.
func (m ServicesMap) WithHealthCheck(serviceName string, healthCheck HealthChecker) ServicesMap {
	cfg := m[serviceName]
	cfg.HealthCheck = healthCheck
	m[serviceName] = cfg
	return m
}
