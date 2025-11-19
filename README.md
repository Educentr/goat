# GOAT = Go Application Testing

A powerful framework for integration testing of Go applications with Docker containers and service mocking.

## What is it?

GOAT is a comprehensive testing framework built on [testcontainers-go](https://golang.testcontainers.org/) and [gomock](https://github.com/uber-go/mock). It enables you to:
- Start your application and all dependencies in Docker containers
- Mock external services with type-safe interfaces
- Write integration tests with real infrastructure
- Manage complex test environments with ease

## ✨ Key Features

- **Clean Architecture**: No `/tools` suffix - simple `github.com/Educentr/goat` import
- **Generic Type-Safe Getters**: `services.GetTyped[T]()` for compile-time type safety
- **External Service Support**: Integrate with [goat-services](https://github.com/Educentr/goat-services) for 9+ ready-to-use services
- **Full Context Support**: Proper context propagation for cancellation control
- **ExecutorBuilder**: Fluent API with `NewExecutorBuilder(binary).WithEnv(...).Build()`
- **Flow with Callbacks**: Direct callback parameters in Start/Stop with testify/require
- **Service Restart**: `manager.Restart(ctx, "postgres")` and `manager.RestartAll(ctx)`
- **No Circular Dependencies**: Clean separation between framework and service packages

## Requirements

- Docker
- Go 1.23+

## Quick Start

### 1. Install

```bash
go get github.com/Educentr/goat@latest
go get github.com/Educentr/goat-services@latest
```

### 2. Create main_test.go

```go
package myapp_test

import (
    "context"
    "testing"

    gtt "github.com/Educentr/goat"
    "github.com/Educentr/goat/services"
    testcontainers "github.com/testcontainers/testcontainers-go"

    // Import services from goat-services
    "github.com/Educentr/goat-services/psql"
    "github.com/Educentr/goat-services/redis"
)

var env *gtt.Env

// wrapServiceRunner wraps typed service runners for registration
func wrapServiceRunner[T testcontainers.Container](fn func(context.Context, ...testcontainers.ContainerCustomizer) (T, error)) func(context.Context, ...testcontainers.ContainerCustomizer) (testcontainers.Container, error) {
    return func(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (testcontainers.Container, error) {
        return fn(ctx, opts...)
    }
}

func init() {
    // Register services from goat-services
    services.MustRegisterServiceFunc("postgres", wrapServiceRunner(psql.Run))
    services.MustRegisterServiceFunc("redis", wrapServiceRunner(redis.Run))

    // Create manager and enable services
    servicesMap := services.NewServicesMap()
    servicesMap.Enable("postgres")
    servicesMap.Enable("redis")

    manager := services.NewManager(servicesMap, services.DefaultManagerConfig())
    env = gtt.NewEnv(gtt.EnvConfig{}, manager)
}

func TestMain(m *testing.M) {
    gtt.CallMain(env, m)
}
```

### 3. Available Services

Via [goat-services](https://github.com/Educentr/goat-services) package:
- **postgres** - PostgreSQL database
- **redis** - Redis cache
- **clickhouse** - ClickHouse analytics database
- **s3** - S3-compatible storage (LocalStack)
- **minio** - MinIO object storage
- **jaeger** - Jaeger distributed tracing
- **victoriametrics** - VictoriaMetrics time-series database
- **xray** - Xray proxy server
- **singbox** - Singbox VPN proxy

### 4. Access Services with Type-Safe Getters

```go
import (
    "github.com/Educentr/goat/services"
    "github.com/Educentr/goat-services/psql"
    "github.com/Educentr/goat-services/redis"
)

func NewApp(env *gtt.Env) *gtt.Executor {
    // Use generic type-safe getters - compile-time type checking!
    pg := services.MustGetTyped[*psql.Env](env.Manager(), "postgres")
    rd := services.MustGetTyped[*redis.Env](env.Manager(), "redis")

    // Or with error handling
    pg, err := services.GetTyped[*psql.Env](env.Manager(), "postgres")
    if err != nil {
        panic(err)
    }

    // Build environment variables for your app
    envVars := map[string]string{
        "DB_HOST":     pg.DBHost,
        "DB_NAME":     pg.DBName,
        "DB_PASSWORD": pg.DBPass,
        "DB_PORT":     pg.DBPort,
        "DB_USER":     pg.DBUser,
        "REDIS_ADDR":  rd.Address,
    }

    binaryPath := os.Getenv("APP_BINARY")
    if binaryPath == "" {
        binaryPath = "/tmp/myapp"
    }

    // Create executor with fluent builder
    return gtt.NewExecutorBuilder(binaryPath).
        WithEnv(envVars).
        WithOutputFile("/tmp/myapp-test.log").
        Build()
}
```

### 5. Advanced Configuration

**Custom service options:**

```go
import (
    "github.com/Educentr/goat/services"
    testcontainers "github.com/testcontainers/testcontainers-go"
    "github.com/Educentr/goat-services/psql"
)

func init() {
    services.MustRegisterServiceFunc("postgres", wrapServiceRunner(psql.Run))

    // Create manager with custom configuration
    servicesMap := services.NewServicesMap()
    servicesMap.Enable("postgres",
        testcontainers.WithImage("postgres:15"),
        testcontainers.WithEnv(map[string]string{
            "POSTGRES_MAX_CONNECTIONS": "200",
        }),
    )

    // Configure manager settings
    config := services.DefaultManagerConfig()
    config.Logger = services.NewDefaultLogger()  // Enable logging
    config.MaxParallel = 5                       // Parallel startup limit

    manager := services.NewManager(servicesMap, config)
    env = gtt.NewEnv(gtt.EnvConfig{}, manager)
}
```

**Using Builder pattern:**

```go
func init() {
    // Register services
    services.MustRegisterServiceFunc("postgres", wrapServiceRunner(psql.Run))
    services.MustRegisterServiceFunc("redis", wrapServiceRunner(redis.Run))

    // Build manager with fluent API
    manager := services.NewBuilder().
        WithServiceSimple("postgres", testcontainers.WithImage("postgres:15")).
        WithServiceSimple("redis").
        WithLogger(services.NewDefaultLogger()).
        WithMaxParallel(3).
        Build()

    env = gtt.NewEnv(gtt.EnvConfig{}, manager)
}
```

### 6. Configure Mocks and Flow

```go
import (
    "net/http"
    "testing"

    "go.uber.org/mock/gomock"
    gtt "github.com/Educentr/goat"
    "github.com/Educentr/goat/services"
    "github.com/Educentr/goat-services/psql"
)

type Mocks struct {
    PaymentAPI *paymentmock.MockClient
}

func prepareFlow(t *testing.T, env *gtt.Env) (Mocks, func()) {
    var m Mocks

    // Create flow with HTTP and gRPC mock callbacks
    flow := gtt.NewFlow(
        t,
        env,
        NewApp(env),
        func(server *http.ServeMux, ctl *gomock.Controller) {
            // Register HTTP mocks
            m.PaymentAPI = paymentmock.NewMockClient(ctl)
            paymentmock.RegisterHandlers(server, m.PaymentAPI)
        },
        nil, // gRPC callback (nil if not needed)
    )

    // Start flow with before/after callbacks
    flow.Start(t,
        func(env *gtt.Env) error {
            // Before app start: initialize database, etc.
            pg := services.MustGetTyped[*psql.Env](env.Manager(), "postgres")
            return initDatabase(pg)
        },
        func(env *gtt.Env) error {
            // After app start: wait for readiness, etc.
            return waitForApp()
        },
    )

    // Return cleanup function
    stop := func() {
        flow.Stop(t,
            nil, // before stop callback
            func(env *gtt.Env) error {
                // Cleanup: drop database schema, etc.
                pg := services.MustGetTyped[*psql.Env](env.Manager(), "postgres")
                sql, err := pg.SQL()
                if err != nil {
                    return err
                }
                _, err = sql.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;")
                return err
            },
        )
    }

    return m, stop
}
```

### 7. Write Tests

```go
func TestPaymentFlow(t *testing.T) {
    mocks, stop := prepareFlow(t, env)
    defer stop()

    // Set mock expectations
    mocks.PaymentAPI.EXPECT().
        CreatePayment(gomock.Any(), gomock.Any()).
        Return(&payment.Response{ID: "pay_123"}, nil).
        Times(1)

    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Test your application
    client := newAppClient(ctx, t)
    resp, err := client.CreateOrder(ctx, &api.CreateOrderRequest{
        Amount: 100,
    })

    require.NoError(t, err)
    assert.Equal(t, "pay_123", resp.PaymentID)
}
```

## Service Management

**Restart services during tests:**

```go
// Restart specific service
err := env.Manager().Restart(ctx, "postgres")

// Restart all services
err := env.Manager().RestartAll(ctx)
```

**Check service status:**

```go
// Check if running
isRunning := env.Manager().IsRunning("postgres")

// List all running services
services := env.Manager().ListRunning()
```

## Environment Variables

**Docker proxy support:**

```bash
export DOCKER_PROXY=your-registry.example.com
# Images will be pulled from your-registry.example.com/postgres:latest
```

**Debug mode:**

```bash
export GOAT_REMOTE_DEBUG=true
export GOAT_REMOTE_DEBUG_PORT=2345
```

## Migration from v0.0.x

### Import Path Changes

```go
// Old
import gtt "github.com/Educentr/goat/tools"
import "github.com/Educentr/goat/tools/services"

// New
import gtt "github.com/Educentr/goat"
import "github.com/Educentr/goat/services"
```

### Typed Getters

```go
// Old
pg := env.MustGetPostgres()

// New
import "github.com/Educentr/goat-services/psql"
pg := services.MustGetTyped[*psql.Env](env.Manager(), "postgres")
```

### Environment Creation

```go
// Old
env = gtt.NewEnv(gtt.EnvConfig{}, []string{"postgres", "redis"})

// New - register services first
services.MustRegisterServiceFunc("postgres", wrapServiceRunner(psql.Run))
services.MustRegisterServiceFunc("redis", wrapServiceRunner(redis.Run))

servicesMap := services.NewServicesMap()
servicesMap.Enable("postgres")
servicesMap.Enable("redis")

manager := services.NewManager(servicesMap, services.DefaultManagerConfig())
env = gtt.NewEnv(gtt.EnvConfig{}, manager)
```

### Builder Pattern

```go
// Old
builder := services.NewBuilder().WithPostgres().WithRedis()

// New
builder := services.NewBuilder().
    WithServiceSimple("postgres").
    WithServiceSimple("redis")
```

## Architecture

GOAT follows a clean architecture with clear separation:

```
Your Test Code
    ↓
    ├─→ github.com/Educentr/goat (framework)
    └─→ github.com/Educentr/goat-services (service containers)

No circular dependencies! ✅
```

## Documentation

- **CLAUDE.md** - Detailed architecture and patterns
- **examples/** - Sample projects (coming soon)
- [goat-services README](https://github.com/Educentr/goat-services) - Service documentation

## Contributing

Contributions are welcome! Please read our contributing guidelines.

## License

MIT License - see LICENSE file for details

## Credits

Built with:
- [testcontainers-go](https://github.com/testcontainers/testcontainers-go) - Docker container management
- [gomock](https://github.com/uber-go/mock) - Mock generation and testing

---

**Version:** v0.1.0
**Go Version:** 1.23+
