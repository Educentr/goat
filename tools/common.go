package tools

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/Educentr/goat/services/clickhouse"
	"github.com/Educentr/goat/services/jaeger"
	"github.com/Educentr/goat/services/minio"
	"github.com/Educentr/goat/services/psql"
	"github.com/Educentr/goat/services/redis"
	"github.com/Educentr/goat/services/s3"
	"github.com/Educentr/goat/services/victoriametrics"
	"github.com/Educentr/goat/services/xray"
	"github.com/Educentr/goat/tools/services"
)

// EnvConfig holds configuration for the testing environment.
// This is kept for future extensibility but currently empty.
type EnvConfig struct {
	// Reserved for future configuration options
}

type Env struct {
	manager         *services.Manager
	Executor        BaseExecutor
	Conf            EnvConfig
	logFields       map[string]string
	unmarshalErrors int
}

func getFieldsCollectorFilePath() string {
	//TODO: implement something smarter
	return os.Getenv("GOAT_LOG_FIELDS_FILE")
}

func SafeCallError(fn func() error) (err error) {
	defer func() {
		if errRecover := recover(); errRecover != nil {
			err = fmt.Errorf("panic detected: %s", errRecover)
		}
	}()
	err = fn()
	return
}

func (e *Env) mergeLogFieldStats(logFields map[string]string, unmarshalErrors int) {
	if e.logFields == nil {
		e.logFields = make(map[string]string)
	}
	for k, v := range logFields {
		e.logFields[k] = v
	}
	e.unmarshalErrors += unmarshalErrors
}

// NewEnv creates a new environment with services from a simple list of names.
// This is the simplest way to create an environment with default configuration.
//
// Example:
//
//	env := NewEnv(EnvConfig{
//		StartTimeout: time.Second * 30,
//		StopTimeout:  time.Second * 10,
//	}, []string{"postgres", "redis"})
func NewEnv(envConf EnvConfig, servicesList []string) *Env {
	return &Env{
		manager: services.NewManagerFromList(servicesList),
		Conf:    envConf,
	}
}

// NewEnvWithBuilder creates a new environment using a services.Builder for advanced configuration.
//
// Example:
//
//	builder := services.NewBuilder().
//		WithPostgres(testcontainers.WithImage("postgres:15")).
//		WithRedis().
//		WithLogger(services.NewDefaultLogger())
//	env := NewEnvWithBuilder(EnvConfig{...}, builder)
func NewEnvWithBuilder(envConf EnvConfig, builder *services.Builder) *Env {
	return &Env{
		manager: builder.Build(),
		Conf:    envConf,
	}
}

// NewEnvWithManager creates a new environment with an existing services.Manager.
// This provides the most flexibility for custom configurations.
//
// Example:
//
//	manager := services.NewManager(servicesMap, managerConfig)
//	env := NewEnvWithManager(EnvConfig{...}, manager)
func NewEnvWithManager(envConf EnvConfig, manager *services.Manager) *Env {
	return &Env{
		manager: manager,
		Conf:    envConf,
	}
}

// Manager returns the services.Manager instance for direct access to service management.
func (e *Env) Manager() *services.Manager {
	return e.manager
}

// Typed getters - delegates to Manager

// GetPostgres returns the Postgres service environment.
func (e *Env) GetPostgres() (*psql.Env, error) {
	return e.manager.GetPostgres()
}

// MustGetPostgres returns the Postgres service environment or panics if not found.
func (e *Env) MustGetPostgres() *psql.Env {
	return e.manager.MustGetPostgres()
}

// GetRedis returns the Redis service environment.
func (e *Env) GetRedis() (*redis.Env, error) {
	return e.manager.GetRedis()
}

// MustGetRedis returns the Redis service environment or panics if not found.
func (e *Env) MustGetRedis() *redis.Env {
	return e.manager.MustGetRedis()
}

// GetClickHouse returns the ClickHouse service environment.
func (e *Env) GetClickHouse() (*clickhouse.Env, error) {
	return e.manager.GetClickHouse()
}

// MustGetClickHouse returns the ClickHouse service environment or panics if not found.
func (e *Env) MustGetClickHouse() *clickhouse.Env {
	return e.manager.MustGetClickHouse()
}

// GetS3 returns the S3 (LocalStack) service environment.
func (e *Env) GetS3() (*s3.Env, error) {
	return e.manager.GetS3()
}

// MustGetS3 returns the S3 (LocalStack) service environment or panics if not found.
func (e *Env) MustGetS3() *s3.Env {
	return e.manager.MustGetS3()
}

// GetJaeger returns the Jaeger service environment.
func (e *Env) GetJaeger() (*jaeger.Env, error) {
	return e.manager.GetJaeger()
}

// MustGetJaeger returns the Jaeger service environment or panics if not found.
func (e *Env) MustGetJaeger() *jaeger.Env {
	return e.manager.MustGetJaeger()
}

// GetMinio returns the MinIO service environment.
func (e *Env) GetMinio() (*minio.Env, error) {
	return e.manager.GetMinio()
}

// MustGetMinio returns the MinIO service environment or panics if not found.
func (e *Env) MustGetMinio() *minio.Env {
	return e.manager.MustGetMinio()
}

// GetVictoriaMetrics returns the VictoriaMetrics service environment.
func (e *Env) GetVictoriaMetrics() (*victoriametrics.Env, error) {
	return e.manager.GetVictoriaMetrics()
}

// MustGetVictoriaMetrics returns the VictoriaMetrics service environment or panics if not found.
func (e *Env) MustGetVictoriaMetrics() *victoriametrics.Env {
	return e.manager.MustGetVictoriaMetrics()
}

// GetXray returns the XRay service environment.
func (e *Env) GetXray() (*xray.Env, error) {
	return e.manager.GetXray()
}

// MustGetXray returns the XRay service environment or panics if not found.
func (e *Env) MustGetXray() *xray.Env {
	return e.manager.MustGetXray()
}

// Start starts all configured services with the given context.
// The context should have a timeout to prevent hanging.
func (e *Env) Start(ctx context.Context) error {
	fmt.Println("start env")

	if err := e.manager.Start(ctx); err != nil {
		return err
	}

	return nil
}

// Stop stops all running services with the given context.
// The context should have a timeout to prevent hanging.
func (e *Env) Stop(ctx context.Context) error {
	fmt.Println("stop env")
	return e.manager.Stop(ctx)
}

// CallMain is a helper function to be called from TestMain.
// It starts the environment, runs tests, stops the environment, and validates log fields.
// Uses background context with no timeout - if you need timeouts, manage them yourself.
func CallMain(env *Env, m *testing.M) {
	var exitCode = 0

	ctx := context.Background()

	if err := env.Start(ctx); err != nil {
		println("can't start environment ", err.Error())
		exitCode = 1
	} else {
		if safeErr := SafeCallError(func() error {
			exitCode = m.Run()
			return nil
		}); safeErr != nil {
			println("panic detected ", safeErr.Error())
			exitCode = 1
		}
	}

	_ = env.Stop(ctx) //nolint:errcheck // best effort cleanup, exit code already set

	logFieldsPath := getFieldsCollectorFilePath()
	if logFieldsPath != "" {
		if err := validateLogFields(logFieldsPath, env.logFields, env.unmarshalErrors); err != nil {
			println("failed validate log fields ", err.Error())
			exitCode = 1
		}
	}

	os.Exit(exitCode)
}
