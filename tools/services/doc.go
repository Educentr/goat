// Package services provides a flexible and extensible framework for managing
// service containers in integration tests.
//
// The package is built around the concept of ServiceRunner - an interface that
// allows uniform management of different service types (PostgreSQL, Redis, Kafka, etc.).
// Services are registered in a Registry and managed by a Manager that handles their
// lifecycle (start/stop) with support for priorities, dependencies, and health checks.
//
// # Basic Usage
//
// The simplest way to use the package is through the Builder:
//
//	manager := services.NewBuilder().
//		WithPostgres().
//		WithRedis().
//		Build()
//
//	if err := manager.Start(ctx); err != nil {
//		log.Fatal(err)
//	}
//	defer manager.Stop(ctx)
//
//	// Get service container
//	pg, err := manager.GetContainer("postgres")
//	if err != nil {
//		log.Fatal(err)
//	}
//
// # Advanced Configuration
//
// For more control, you can configure services with priorities and dependencies:
//
//	services := services.NewServicesMap().
//		Add("postgres", services.Config{
//			Enabled:  true,
//			Priority: 1,
//		}).
//		Add("myapp", services.Config{
//			Enabled:      true,
//			Priority:     2,
//			Dependencies: []string{"postgres"},
//		})
//
//	manager := services.NewManager(services, services.DefaultManagerConfig())
//
// # Custom Services
//
// You can register custom service runners:
//
//	type MyServiceRunner struct{}
//
//	func (r *MyServiceRunner) Run(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (testcontainers.Container, error) {
//		// Your service startup logic
//	}
//
//	func (r *MyServiceRunner) Name() string { return "myservice" }
//
//	services.MustRegister("myservice", &MyServiceRunner{})
//
// # Logging
//
// The package uses a Logger interface for structured logging. You can provide your own
// implementation or use the default logger:
//
//	logger := services.NewDefaultLoggerWithLevel(services.DebugLevel)
//	manager := services.NewBuilder().
//		WithLogger(logger).
//		Build()
//
// # Error Handling
//
// The package defines typed errors for different failure scenarios:
//
//	if err := manager.Start(ctx); err != nil {
//		switch e := err.(type) {
//		case *services.ErrServiceStartFailed:
//			log.Printf("Failed to start %s: %v", e.ServiceName, e.Cause)
//		case *services.ErrHealthCheckFailed:
//			log.Printf("Health check failed for %s: %v", e.ServiceName, e.Cause)
//		}
//	}
package services
