package services

import (
	"context"
	"sort"
	"sync"

	testcontainers "github.com/testcontainers/testcontainers-go"
	"golang.org/x/sync/errgroup"
)

// ServiceEnv wraps a running service container with metadata.
type ServiceEnv struct {
	Name      string
	Container testcontainers.Container
	Config    Config
}

// Manager manages the lifecycle of multiple service containers.
type Manager struct {
	running  map[string]*ServiceEnv
	config   ServicesMap
	registry *Registry
	mconfig  ManagerConfig
	mu       sync.RWMutex
}

// NewManager creates a new service manager with the given configuration.
func NewManager(services ServicesMap, config ManagerConfig) *Manager {
	if config.Logger == nil {
		config.Logger = NewDefaultLogger()
	}

	return &Manager{
		config:   services,
		mconfig:  config,
		registry: DefaultRegistry,
		running:  make(map[string]*ServiceEnv),
	}
}

// NewManagerWithRegistry creates a new service manager with a custom registry.
func NewManagerWithRegistry(services ServicesMap, config ManagerConfig, registry *Registry) *Manager {
	if config.Logger == nil {
		config.Logger = NewDefaultLogger()
	}

	return &Manager{
		config:   services,
		mconfig:  config,
		registry: registry,
		running:  make(map[string]*ServiceEnv),
	}
}

// Start starts all enabled services.
func (m *Manager) Start(ctx context.Context) error {
	m.mconfig.Logger.Info("starting services", "total", len(m.config))

	// Group services by priority
	groups := m.groupByPriority()

	// Start each priority group sequentially
	for _, priority := range m.getSortedPriorities(groups) {
		if err := m.startGroup(ctx, priority, groups[priority]); err != nil {
			if m.mconfig.StopOnError {
				m.mconfig.Logger.Error("stopping all services due to error")
				_ = m.Stop(context.Background()) //nolint:errcheck // best effort cleanup on error
			}
			return err
		}
	}

	m.mconfig.Logger.Info("all services started successfully")
	return nil
}

// Stop stops all running services.
func (m *Manager) Stop(ctx context.Context) error {
	m.mu.RLock()
	envs := make([]*ServiceEnv, 0, len(m.running))
	for _, env := range m.running {
		envs = append(envs, env)
	}
	m.mu.RUnlock()

	if len(envs) == 0 {
		m.mconfig.Logger.Info("no services to stop")
		return nil
	}

	m.mconfig.Logger.Info("stopping services", "count", len(envs))

	// Sort by priority (reverse order - highest priority stops first)
	sort.Slice(envs, func(i, j int) bool {
		return envs[i].Config.Priority > envs[j].Config.Priority
	})

	eg, egCtx := errgroup.WithContext(ctx)

	for _, env := range envs {
		env := env
		eg.Go(func() error {
			return m.stopService(egCtx, env)
		})
	}

	if err := eg.Wait(); err != nil {
		m.mconfig.Logger.Error("failed to stop some services", "error", err)
		return err
	}

	m.mconfig.Logger.Info("all services stopped successfully")
	return nil
}

// Get retrieves a running service by name.
func (m *Manager) Get(name string) (*ServiceEnv, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	env, ok := m.running[name]
	if !ok {
		return nil, &ErrServiceNotRunning{ServiceName: name}
	}

	return env, nil
}

// GetContainer retrieves the container for a running service.
func (m *Manager) GetContainer(name string) (testcontainers.Container, error) {
	env, err := m.Get(name)
	if err != nil {
		return nil, err
	}
	return env.Container, nil
}

// IsRunning checks if a service is currently running.
func (m *Manager) IsRunning(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, ok := m.running[name]
	return ok
}

// ListRunning returns a list of all running service names.
func (m *Manager) ListRunning() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.running))
	for name := range m.running {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Restart stops and then starts a specific service.
// Returns an error if the service is not currently running or fails to restart.
func (m *Manager) Restart(ctx context.Context, serviceName string) error {
	m.mconfig.Logger.Info("restarting service", "name", serviceName)

	// Get the service env
	env, err := m.Get(serviceName)
	if err != nil {
		return err
	}

	// Stop the service
	if stopErr := m.stopService(ctx, env); stopErr != nil {
		return stopErr
	}

	// Start it again
	if startErr := m.startService(ctx, serviceName, &env.Config); startErr != nil {
		return startErr
	}

	m.mconfig.Logger.Info("service restarted", "name", serviceName)
	return nil
}

// RestartAll stops and restarts all running services.
// Services are restarted in priority order.
func (m *Manager) RestartAll(ctx context.Context) error {
	m.mconfig.Logger.Info("restarting all services")

	m.mu.RLock()
	envs := make([]*ServiceEnv, 0, len(m.running))
	for _, env := range m.running {
		envs = append(envs, env)
	}
	m.mu.RUnlock()

	if len(envs) == 0 {
		m.mconfig.Logger.Info("no services to restart")
		return nil
	}

	// Stop all services
	if err := m.Stop(ctx); err != nil {
		return err
	}

	// Rebuild config map from stopped services
	configs := make(ServicesMap)
	for _, env := range envs {
		configs[env.Name] = env.Config
	}

	// Create temporary manager with same config
	tempManager := NewManager(configs, m.mconfig)
	tempManager.registry = m.registry

	// Start all services
	if err := tempManager.Start(ctx); err != nil {
		return err
	}

	// Copy running services back to this manager
	m.mu.Lock()
	m.running = tempManager.running
	m.mu.Unlock()

	m.mconfig.Logger.Info("all services restarted")
	return nil
}

func (m *Manager) groupByPriority() map[int]map[string]Config {
	groups := make(map[int]map[string]Config)

	for name, cfg := range m.config {
		if !cfg.Enabled {
			continue
		}

		if groups[cfg.Priority] == nil {
			groups[cfg.Priority] = make(map[string]Config)
		}
		groups[cfg.Priority][name] = cfg
	}

	return groups
}

func (m *Manager) getSortedPriorities(groups map[int]map[string]Config) []int {
	priorities := make([]int, 0, len(groups))
	for p := range groups {
		priorities = append(priorities, p)
	}
	sort.Ints(priorities)
	return priorities
}

func (m *Manager) startGroup(ctx context.Context, priority int, configs map[string]Config) error {
	m.mconfig.Logger.Debug("starting service group", "priority", priority, "count", len(configs))

	eg, egCtx := errgroup.WithContext(ctx)
	eg.SetLimit(m.mconfig.MaxParallel)

	for name, cfg := range configs {
		name, cfg := name, cfg
		eg.Go(func() error {
			return m.startService(egCtx, name, &cfg)
		})
	}

	return eg.Wait()
}

func (m *Manager) startService(ctx context.Context, name string, cfg *Config) error {
	m.mconfig.Logger.Debug("starting service", "name", name)

	// Check dependencies
	for _, dep := range cfg.Dependencies {
		if !m.IsRunning(dep) {
			return &ErrDependencyNotMet{ServiceName: name, DependencyName: dep}
		}
	}

	// Get runner
	runner, ok := m.registry.Get(name)
	if !ok {
		return &ErrServiceNotFound{ServiceName: name}
	}

	// Run container
	container, err := runner.Run(ctx, cfg.Opts...)
	if err != nil {
		return &ErrServiceStartFailed{ServiceName: name, Cause: err}
	}

	// Health check
	if cfg.HealthCheck != nil {
		if healthErr := cfg.HealthCheck.Check(ctx, container); healthErr != nil {
			_ = container.Terminate(ctx) //nolint:errcheck // best effort cleanup on health check failure
			return &ErrHealthCheckFailed{ServiceName: name, Cause: healthErr}
		}
	}

	// Store running service
	m.mu.Lock()
	m.running[name] = &ServiceEnv{
		Container: container,
		Name:      name,
		Config:    *cfg,
	}
	m.mu.Unlock()

	m.mconfig.Logger.Info("service started", "name", name)
	return nil
}

func (m *Manager) stopService(ctx context.Context, env *ServiceEnv) error {
	m.mconfig.Logger.Debug("stopping service", "name", env.Name)

	if err := env.Container.Terminate(ctx); err != nil {
		return &ErrServiceStopFailed{ServiceName: env.Name, Cause: err}
	}

	m.mu.Lock()
	delete(m.running, env.Name)
	m.mu.Unlock()

	m.mconfig.Logger.Info("service stopped", "name", env.Name)
	return nil
}
