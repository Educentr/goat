package services

import testcontainers "github.com/testcontainers/testcontainers-go"

// NewManagerFromList creates a Manager from a list of service names.
// All services will be enabled with default configuration.
//
// Example:
//
//	manager := services.NewManagerFromList([]string{"postgres", "redis"})
//	if err := manager.Start(ctx); err != nil {
//		log.Fatal(err)
//	}
func NewManagerFromList(names []string) *Manager {
	servicesMap := NewServicesMap()
	for _, name := range names {
		servicesMap.Enable(name)
	}
	return NewManager(servicesMap, DefaultManagerConfig())
}

// NewManagerFromMap creates a Manager from a map of service names to container customizers.
// This allows fine-grained control over each service's configuration.
//
// Example:
//
//	manager := services.NewManagerFromMap(map[string][]testcontainers.ContainerCustomizer{
//		"postgres": {testcontainers.WithImage("postgres:15")},
//		"redis":    {testcontainers.WithImage("redis:7")},
//	})
func NewManagerFromMap(cfg map[string][]testcontainers.ContainerCustomizer) *Manager {
	servicesMap := NewServicesMap()
	for name, opts := range cfg {
		servicesMap.Enable(name, opts...)
	}
	return NewManager(servicesMap, DefaultManagerConfig())
}

// WithMounts is a helper function to add mounts to a container request.
// This is useful for mounting configuration files or data directories into containers.
//
// Example:
//
//	mounts := testcontainers.ContainerMounts{
//		testcontainers.BindMount("/host/path", "/container/path"),
//	}
//	services.NewBuilder().WithPostgres(services.WithMounts(mounts))
func WithMounts(mounts testcontainers.ContainerMounts) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Mounts = append(req.Mounts, mounts...)
		return nil
	}
}
