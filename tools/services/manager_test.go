package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	testcontainers "github.com/testcontainers/testcontainers-go"
)

// MockRunner is a mock implementation of ServiceRunner for testing
type MockRunner struct {
	runFunc func(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (testcontainers.Container, error)
	name    string
}

func (m *MockRunner) Run(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (testcontainers.Container, error) {
	if m.runFunc != nil {
		return m.runFunc(ctx, opts...)
	}
	return nil, nil
}

func (m *MockRunner) Name() string {
	return m.name
}

func TestRegistry(t *testing.T) {
	t.Run("Register and Get", func(t *testing.T) {
		registry := NewRegistry()
		runner := &MockRunner{name: "test"}

		err := registry.Register("test", runner)
		require.NoError(t, err)

		got, ok := registry.Get("test")
		assert.True(t, ok)
		assert.Equal(t, runner, got)
	})

	t.Run("Register duplicate", func(t *testing.T) {
		registry := NewRegistry()
		runner := &MockRunner{name: "test"}

		err := registry.Register("test", runner)
		require.NoError(t, err)

		err = registry.Register("test", runner)
		assert.Error(t, err)
		assert.IsType(t, &ErrServiceAlreadyRegistered{}, err)
	})

	t.Run("Get nonexistent", func(t *testing.T) {
		registry := NewRegistry()

		got, ok := registry.Get("nonexistent")
		assert.False(t, ok)
		assert.Nil(t, got)
	})

	t.Run("List services", func(t *testing.T) {
		registry := NewRegistry()
		registry.MustRegister("test1", &MockRunner{name: "test1"})
		registry.MustRegister("test2", &MockRunner{name: "test2"})

		list := registry.List()
		assert.Len(t, list, 2)
		assert.Contains(t, list, "test1")
		assert.Contains(t, list, "test2")
	})

	t.Run("Unregister", func(t *testing.T) {
		registry := NewRegistry()
		registry.MustRegister("test", &MockRunner{name: "test"})

		assert.True(t, registry.Has("test"))

		registry.Unregister("test")
		assert.False(t, registry.Has("test"))
	})
}

func TestServicesMap(t *testing.T) {
	t.Run("Add service", func(t *testing.T) {
		m := NewServicesMap()
		cfg := Config{Enabled: true, Priority: 1}

		m.Add("test", cfg)

		assert.Len(t, m, 1)
		assert.Equal(t, cfg, m["test"])
	})

	t.Run("Enable service", func(t *testing.T) {
		m := NewServicesMap()
		m.Enable("test")

		assert.Len(t, m, 1)
		assert.True(t, m["test"].Enabled)
	})

	t.Run("EnableWithPriority", func(t *testing.T) {
		m := NewServicesMap()
		m.EnableWithPriority("test", 5)

		assert.Len(t, m, 1)
		assert.True(t, m["test"].Enabled)
		assert.Equal(t, 5, m["test"].Priority)
	})
}

func TestBuilder(t *testing.T) {
	t.Run("Build with services", func(t *testing.T) {
		builder := NewBuilder()
		builder.WithPostgres().WithRedis()

		manager := builder.Build()

		assert.NotNil(t, manager)
		assert.Len(t, manager.config, 2)
		assert.True(t, manager.config["postgres"].Enabled)
		assert.True(t, manager.config["redis"].Enabled)
	})

	t.Run("Build with custom logger", func(t *testing.T) {
		logger := NewNoopLogger()
		builder := NewBuilder().WithLogger(logger)

		manager := builder.Build()

		assert.NotNil(t, manager)
		assert.Equal(t, logger, manager.mconfig.Logger)
	})

	t.Run("Build with max parallel", func(t *testing.T) {
		builder := NewBuilder().WithMaxParallel(5)

		manager := builder.Build()

		assert.NotNil(t, manager)
		assert.Equal(t, 5, manager.mconfig.MaxParallel)
	})
}

func TestLogger(t *testing.T) {
	t.Run("DefaultLogger", func(_ *testing.T) {
		logger := NewDefaultLogger()

		// Should not panic
		logger.Debug("test", "key", "value")
		logger.Info("test", "key", "value")
		logger.Warn("test", "key", "value")
		logger.Error("test", "key", "value")
	})

	t.Run("NoopLogger", func(_ *testing.T) {
		logger := NewNoopLogger()

		// Should not panic
		logger.Debug("test", "key", "value")
		logger.Info("test", "key", "value")
		logger.Warn("test", "key", "value")
		logger.Error("test", "key", "value")
	})

	t.Run("DefaultLogger with level", func(_ *testing.T) {
		logger := NewDefaultLoggerWithLevel(ErrorLevel)
		logger.SetLevel(DebugLevel)

		// Should not panic
		logger.Debug("test")
	})
}

func TestErrors(t *testing.T) {
	t.Run("ErrServiceNotFound", func(t *testing.T) {
		err := &ErrServiceNotFound{ServiceName: "test"}
		assert.Contains(t, err.Error(), "test")
	})

	t.Run("ErrServiceNotRunning", func(t *testing.T) {
		err := &ErrServiceNotRunning{ServiceName: "test"}
		assert.Contains(t, err.Error(), "test")
	})

	t.Run("ErrServiceStartFailed", func(t *testing.T) {
		cause := assert.AnError
		err := &ErrServiceStartFailed{ServiceName: "test", Cause: cause}
		assert.Contains(t, err.Error(), "test")
		assert.Equal(t, cause, err.Unwrap())
	})

	t.Run("ErrServiceStopFailed", func(t *testing.T) {
		cause := assert.AnError
		err := &ErrServiceStopFailed{ServiceName: "test", Cause: cause}
		assert.Contains(t, err.Error(), "test")
		assert.Equal(t, cause, err.Unwrap())
	})

	t.Run("ErrHealthCheckFailed", func(t *testing.T) {
		cause := assert.AnError
		err := &ErrHealthCheckFailed{ServiceName: "test", Cause: cause}
		assert.Contains(t, err.Error(), "test")
		assert.Equal(t, cause, err.Unwrap())
	})

	t.Run("ErrDependencyNotMet", func(t *testing.T) {
		err := &ErrDependencyNotMet{ServiceName: "test", DependencyName: "dep"}
		assert.Contains(t, err.Error(), "test")
		assert.Contains(t, err.Error(), "dep")
	})
}
