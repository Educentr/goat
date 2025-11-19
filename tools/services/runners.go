package services

import (
	"context"

	testcontainers "github.com/testcontainers/testcontainers-go"

	"github.com/Educentr/goat/services/clickhouse"
	"github.com/Educentr/goat/services/jaeger"
	"github.com/Educentr/goat/services/minio"
	"github.com/Educentr/goat/services/psql"
	"github.com/Educentr/goat/services/redis"
	"github.com/Educentr/goat/services/s3"
	"github.com/Educentr/goat/services/victoriametrics"
	"github.com/Educentr/goat/services/xray"
)

// PostgresRunner is a ServiceRunner for PostgreSQL.
type PostgresRunner struct{}

func (r *PostgresRunner) Run(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (testcontainers.Container, error) {
	return psql.Run(ctx, opts...)
}

func (r *PostgresRunner) Name() string { return "postgres" }

// RedisRunner is a ServiceRunner for Redis.
type RedisRunner struct{}

func (r *RedisRunner) Run(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (testcontainers.Container, error) {
	return redis.Run(ctx, opts...)
}

func (r *RedisRunner) Name() string { return "redis" }

// ClickHouseRunner is a ServiceRunner for ClickHouse.
type ClickHouseRunner struct{}

func (r *ClickHouseRunner) Run(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (testcontainers.Container, error) {
	return clickhouse.Run(ctx, opts...)
}

func (r *ClickHouseRunner) Name() string { return "clickhouse" }

// S3Runner is a ServiceRunner for S3 (LocalStack).
type S3Runner struct{}

func (r *S3Runner) Run(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (testcontainers.Container, error) {
	return s3.Run(ctx, opts...)
}

func (r *S3Runner) Name() string { return "s3" }

// JaegerRunner is a ServiceRunner for Jaeger.
type JaegerRunner struct{}

func (r *JaegerRunner) Run(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (testcontainers.Container, error) {
	return jaeger.Run(ctx, opts...)
}

func (r *JaegerRunner) Name() string { return "jaeger" }

// MinioRunner is a ServiceRunner for MinIO.
type MinioRunner struct{}

func (r *MinioRunner) Run(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (testcontainers.Container, error) {
	return minio.Run(ctx, opts...)
}

func (r *MinioRunner) Name() string { return "minio" }

// VictoriaMetricsRunner is a ServiceRunner for VictoriaMetrics.
type VictoriaMetricsRunner struct{}

func (r *VictoriaMetricsRunner) Run(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (testcontainers.Container, error) {
	return victoriametrics.Run(ctx, opts...)
}

func (r *VictoriaMetricsRunner) Name() string { return "victoriametrics" }

// XRayRunner is a ServiceRunner for XRay.
type XRayRunner struct{}

func (r *XRayRunner) Run(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (testcontainers.Container, error) {
	return xray.Run(ctx, opts...)
}

func (r *XRayRunner) Name() string { return "xray" }

func init() {
	// Register all built-in service runners in the default registry
	MustRegister("postgres", &PostgresRunner{})
	MustRegister("redis", &RedisRunner{})
	MustRegister("clickhouse", &ClickHouseRunner{})
	MustRegister("s3", &S3Runner{})
	MustRegister("jaeger", &JaegerRunner{})
	MustRegister("minio", &MinioRunner{})
	MustRegister("victoriametrics", &VictoriaMetricsRunner{})
	MustRegister("xray", &XRayRunner{})
}
