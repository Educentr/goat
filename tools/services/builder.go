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

// WithPostgres enables PostgreSQL.
func (b *Builder) WithPostgres(opts ...testcontainers.ContainerCustomizer) *Builder {
	b.services.Enable("postgres", opts...)
	return b
}

// WithVault enables Vault.
func (b *Builder) WithVault(opts ...testcontainers.ContainerCustomizer) *Builder {
	b.services.Enable("vault", opts...)
	return b
}

// WithRedis enables Redis.
func (b *Builder) WithRedis(opts ...testcontainers.ContainerCustomizer) *Builder {
	b.services.Enable("redis", opts...)
	return b
}

// WithClickHouse enables ClickHouse.
func (b *Builder) WithClickHouse(opts ...testcontainers.ContainerCustomizer) *Builder {
	b.services.Enable("clickhouse", opts...)
	return b
}

// WithS3 enables S3 (LocalStack).
func (b *Builder) WithS3(opts ...testcontainers.ContainerCustomizer) *Builder {
	b.services.Enable("s3", opts...)
	return b
}

// WithJaeger enables Jaeger.
func (b *Builder) WithJaeger(opts ...testcontainers.ContainerCustomizer) *Builder {
	b.services.Enable("jaeger", opts...)
	return b
}

// WithMinio enables MinIO.
func (b *Builder) WithMinio(opts ...testcontainers.ContainerCustomizer) *Builder {
	b.services.Enable("minio", opts...)
	return b
}

// WithVictoriaMetrics enables VictoriaMetrics.
func (b *Builder) WithVictoriaMetrics(opts ...testcontainers.ContainerCustomizer) *Builder {
	b.services.Enable("victoriametrics", opts...)
	return b
}

// WithXray enables Xray.
func (b *Builder) WithXray(opts ...testcontainers.ContainerCustomizer) *Builder {
	b.services.Enable("xray", opts...)
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
