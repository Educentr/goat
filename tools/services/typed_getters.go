package services

import (
	"github.com/Educentr/goat/services/clickhouse"
	"github.com/Educentr/goat/services/jaeger"
	"github.com/Educentr/goat/services/minio"
	"github.com/Educentr/goat/services/psql"
	"github.com/Educentr/goat/services/redis"
	"github.com/Educentr/goat/services/s3"
	"github.com/Educentr/goat/services/victoriametrics"
	"github.com/Educentr/goat/services/xray"
)

// Typed getters for common services.
// These methods eliminate the need for type assertions in user code.

// GetPostgres returns the Postgres service environment.
func (m *Manager) GetPostgres() (*psql.Env, error) {
	container, err := m.GetContainer("postgres")
	if err != nil {
		return nil, err
	}
	pg, ok := container.(*psql.Env)
	if !ok {
		return nil, &ErrServiceTypeMismatch{ServiceName: "postgres", ExpectedType: "*psql.Env"}
	}
	return pg, nil
}

// MustGetPostgres returns the Postgres service environment or panics if not found.
func (m *Manager) MustGetPostgres() *psql.Env {
	pg, err := m.GetPostgres()
	if err != nil {
		panic(err)
	}
	return pg
}

// GetRedis returns the Redis service environment.
func (m *Manager) GetRedis() (*redis.Env, error) {
	container, err := m.GetContainer("redis")
	if err != nil {
		return nil, err
	}
	rd, ok := container.(*redis.Env)
	if !ok {
		return nil, &ErrServiceTypeMismatch{ServiceName: "redis", ExpectedType: "*redis.Env"}
	}
	return rd, nil
}

// MustGetRedis returns the Redis service environment or panics if not found.
func (m *Manager) MustGetRedis() *redis.Env {
	rd, err := m.GetRedis()
	if err != nil {
		panic(err)
	}
	return rd
}

// GetClickHouse returns the ClickHouse service environment.
func (m *Manager) GetClickHouse() (*clickhouse.Env, error) {
	container, err := m.GetContainer("clickhouse")
	if err != nil {
		return nil, err
	}
	ch, ok := container.(*clickhouse.Env)
	if !ok {
		return nil, &ErrServiceTypeMismatch{ServiceName: "clickhouse", ExpectedType: "*clickhouse.Env"}
	}
	return ch, nil
}

// MustGetClickHouse returns the ClickHouse service environment or panics if not found.
func (m *Manager) MustGetClickHouse() *clickhouse.Env {
	ch, err := m.GetClickHouse()
	if err != nil {
		panic(err)
	}
	return ch
}

// GetS3 returns the S3 (LocalStack) service environment.
func (m *Manager) GetS3() (*s3.Env, error) {
	container, err := m.GetContainer("s3")
	if err != nil {
		return nil, err
	}
	s3Env, ok := container.(*s3.Env)
	if !ok {
		return nil, &ErrServiceTypeMismatch{ServiceName: "s3", ExpectedType: "*s3.Env"}
	}
	return s3Env, nil
}

// MustGetS3 returns the S3 (LocalStack) service environment or panics if not found.
func (m *Manager) MustGetS3() *s3.Env {
	s3Env, err := m.GetS3()
	if err != nil {
		panic(err)
	}
	return s3Env
}

// GetJaeger returns the Jaeger service environment.
func (m *Manager) GetJaeger() (*jaeger.Env, error) {
	container, err := m.GetContainer("jaeger")
	if err != nil {
		return nil, err
	}
	j, ok := container.(*jaeger.Env)
	if !ok {
		return nil, &ErrServiceTypeMismatch{ServiceName: "jaeger", ExpectedType: "*jaeger.Env"}
	}
	return j, nil
}

// MustGetJaeger returns the Jaeger service environment or panics if not found.
func (m *Manager) MustGetJaeger() *jaeger.Env {
	j, err := m.GetJaeger()
	if err != nil {
		panic(err)
	}
	return j
}

// GetMinio returns the MinIO service environment.
func (m *Manager) GetMinio() (*minio.Env, error) {
	container, err := m.GetContainer("minio")
	if err != nil {
		return nil, err
	}
	mn, ok := container.(*minio.Env)
	if !ok {
		return nil, &ErrServiceTypeMismatch{ServiceName: "minio", ExpectedType: "*minio.Env"}
	}
	return mn, nil
}

// MustGetMinio returns the MinIO service environment or panics if not found.
func (m *Manager) MustGetMinio() *minio.Env {
	mn, err := m.GetMinio()
	if err != nil {
		panic(err)
	}
	return mn
}

// GetVictoriaMetrics returns the VictoriaMetrics service environment.
func (m *Manager) GetVictoriaMetrics() (*victoriametrics.Env, error) {
	container, err := m.GetContainer("victoriametrics")
	if err != nil {
		return nil, err
	}
	vm, ok := container.(*victoriametrics.Env)
	if !ok {
		return nil, &ErrServiceTypeMismatch{ServiceName: "victoriametrics", ExpectedType: "*victoriametrics.Env"}
	}
	return vm, nil
}

// MustGetVictoriaMetrics returns the VictoriaMetrics service environment or panics if not found.
func (m *Manager) MustGetVictoriaMetrics() *victoriametrics.Env {
	vm, err := m.GetVictoriaMetrics()
	if err != nil {
		panic(err)
	}
	return vm
}

// GetXray returns the XRay service environment.
func (m *Manager) GetXray() (*xray.Env, error) {
	container, err := m.GetContainer("xray")
	if err != nil {
		return nil, err
	}
	xr, ok := container.(*xray.Env)
	if !ok {
		return nil, &ErrServiceTypeMismatch{ServiceName: "xray", ExpectedType: "*xray.Env"}
	}
	return xr, nil
}

// MustGetXray returns the XRay service environment or panics if not found.
func (m *Manager) MustGetXray() *xray.Env {
	xr, err := m.GetXray()
	if err != nil {
		panic(err)
	}
	return xr
}
