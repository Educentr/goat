# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

GOAT (Go Application Testing) is a Go framework for integration testing of Go applications. It's built on top of testcontainers and gomock, enabling developers to start their applications and all dependencies in Docker containers while mocking external services.

Module: `github.com/Educentr/goat`
Go version: 1.23.4

### Key Features

- **Context propagation**: Full context support in all Start/Stop methods for proper cancellation control
- **Typed service getters**: `env.GetPostgres()`, `env.GetRedis()`, etc. - no more type assertions!
- **ExecutorBuilder**: Fluent API for building executors with options
- **Flow with callbacks**: Direct callback parameters in Start/Stop methods with testify/require for assertions
- **Service restart**: `manager.Restart(ctx, "postgres")` and `manager.RestartAll(ctx)`
- **Better error handling**: Uses testify/require for test assertions

## Common Commands

### Testing
```bash
# Run unit tests (excludes services/*)
make test

# Run tests with coverage report
make coverage

# Run integration tests with Docker (services/* only)
make goat

# Set custom test timeout (default: 300s)
TEST_TIMEOUT=600s make test
```

### Linting
```bash
# Install golangci-lint (v1.62.0)
make install-lint

# Run linter on changed files (diff with origin/main)
make lint

# Run linter on entire codebase
make lint-full

# Auto-fix linting issues
make lint-fix
```

### Environment Variables for Testing

- `GOAT_REMOTE_DEBUG=true` - Attach delve debugger on port 2345
- `GOAT_REMOTE_DEBUG_PORT=<port>` - Custom debug port
- `GOAT_DISABLE_STDOUT=true` - Suppress stdout from test binaries
- `GOAT_OUTPUT_FILE=<path>` - Redirect stdout to file
- `GOAT_OUTPUT_ERRORS_FILE=<path>` - Redirect stderr to file
- `GOAT_LOG_FIELDS_FILE=<path>` - Enable log field validation
- `GOAT_GRPC_MOCK_ADDRESS` - Custom gRPC mock address (default: 127.0.0.1:9191)
- `GOAT_HTTP_MOCK_ADDRESS` - Custom HTTP mock address (default: 127.0.0.1:9898)

## Architecture

### Core Components

#### 1. `tools/` - Main Testing Framework
The primary interface for building integration tests:

- **Env** (`tools/common.go`): Central environment manager
  - Manages lifecycle of all service containers via `services.Manager`
  - Configured via `EnvConfig` and list of service names or `services.Builder`
  - Must be initialized in `TestMain()` and shared across all tests in a package
  - Handles log field validation if configured
  - Three constructors: `NewEnv(config, []string)`, `NewEnvWithBuilder(config, builder)`, `NewEnvWithManager(config, manager)`
  - `Start(ctx)` and `Stop(ctx)` accept context for proper lifecycle management
  - Typed getters: `GetPostgres()`, `MustGetPostgres()`, etc.

- **Flow** (`tools/flow.go`): Test orchestration
  - Coordinates the lifecycle of app + mocks for each test
  - `NewFlow(t, env, exe, httpCb, grpcCb)` - Creates flow
  - `Start(t, before, after)` - Uses require.NoError, starts mocks and app with optional callbacks
  - `Stop(t, before, after)` - Uses require.NoError, stops app and mocks with optional callbacks
  - Before/after callbacks passed directly to Start/Stop methods

- **Executor** (`tools/executor.go`): Binary execution wrapper
  - Starts Go binaries with environment variables
  - Detects data races in stdout/stderr
  - Supports remote debugging via delve
  - Handles log collection and field validation
  - `ExecutorBuilder` (`tools/executor_builder.go`) - Fluent API for building executors with options

- **MocksHandler** (`tools/mock_handler.go`): Mock service orchestration
  - Manages both HTTP (port 9898) and gRPC (port 9191) mock servers
  - Uses gomock for expectations
  - Mocks are registered via callbacks in `NewFlow()`
  - `NewMocksHandler(t, grpcCb, httpCb)` - Creates handler, uses require.NoError for errors

#### 2. `services/` - Individual Service Containers
Each subdirectory manages a specific service container:

- **Available Services**: postgres, redis, clickhouse, s3 (localstack), jaeger, minio, victoriametrics
- **Pattern**: Each service has:
  - `Run(ctx, opts...)` function that returns service-specific `*Env`
  - `Env` struct embedding `testcontainers.Container` with service-specific connection details
  - Support for custom configuration via `testcontainers.ContainerCustomizer`
- **Common utilities** in `services/common/`:
  - `DockerProxy()` - Proxies Docker images
  - `ImageSubstitutors()` - Handles image substitution for all containers including reaper

#### 3. `tools/services/` - Service Management Framework
Provides flexible service orchestration:

- **Manager** (`manager.go`): Lifecycle management for all services
  - Priority-based parallel startup with `errgroup`
  - Dependency resolution
  - Health checks
  - Structured logging
  - Thread-safe container access
  - Typed getters (`typed_getters.go`) - `GetPostgres()`, `MustGetPostgres()`, etc.
  - Service restart: `Restart(ctx, serviceName)` and `RestartAll(ctx)` methods

- **Builder** (`builder.go`): Fluent API for service configuration
  - Methods like `WithPostgres()`, `WithRedis()`, etc.
  - Configure logging, parallelism, error handling
  - Example: `services.NewBuilder().WithPostgres().WithRedis().Build()`

- **Registry** (`registry.go`): Service runner registry
  - Thread-safe registration of service runners
  - Built-in runners for all available services
  - Extensible for custom services

- **Config Types** (`config.go`):
  - `ServicesMap` - map of service configs
  - `Config` - per-service configuration (enabled, priority, dependencies, options)
  - `ManagerConfig` - global manager settings

- **Helper Functions** (`helpers.go`):
  - `NewManagerFromList([]string)` - Create manager from simple list
  - `NewManagerFromMap(map[string][]ContainerCustomizer)` - From map with options
  - `WithMounts()` - Helper for adding volume mounts

#### 4. Test Flow Pattern

Typical test structure:
1. **Package-level setup** in `main_test.go`:
   - Initialize `*gtt.Env` in `init()`
   - Call `gtt.CallMain(env, m)` in `TestMain(m *testing.M)`
   - This starts all configured containers once for the entire package

2. **Per-test flow**:
   - Create app-specific executor with environment variables mapped to container addresses
   - Create `Flow` with executor + HTTP/gRPC mock callbacks using `NewFlow(t, env, exe, httpCb, grpcCb)`
   - Call `f.Start(t, beforeCallback, afterCallback)` to start app with optional hooks
   - Set mock expectations using gomock
   - Execute test logic against the running app
   - Call `f.Stop(t, beforeCallback, afterCallback)` with cleanup (typically `defer stop()`)

3. **Mock Registration**:
   - HTTP mocks: Registered via `HTTPCB` callback, receive `*http.ServeMux` and `*gomock.Controller`
   - gRPC mocks: Registered via `GrpcCB` callback, receive `*grpc.Server` and `*gomock.Controller`

### Key Design Patterns

- **Container Lifecycle**: Containers are started once per test package (in `TestMain`), not per test
- **App Lifecycle**: The application binary is started/stopped per test via `Flow`
- **Mock Lifecycle**: Mock servers run per test, expectations are set before app starts
- **Environment Mapping**: Container connection details (host, port) are passed to app via environment variables
- **Race Detection**: Executor automatically scans stdout/stderr for "WARNING: DATA RACE" pattern
- **Cleanup Pattern**: Use `defer stop()` immediately after `prepareFlow()` to ensure cleanup

### Service Configuration

Services can be configured in three ways:

**1. Simple list (most common):**
```go
env := gtt.NewEnv(gtt.EnvConfig{}, []string{"postgres", "redis"})
```

**2. Builder pattern (for customization):**
```go
import "github.com/Educentr/goat/services"

builder := services.NewBuilder().
    WithPostgres(testcontainers.WithImage("postgres:15")).
    WithRedis().
    WithLogger(services.NewDefaultLogger()).
    WithMaxParallel(5)

env := gtt.NewEnvWithBuilder(gtt.EnvConfig{}, builder)
```

**3. Direct manager (full control):**
```go
servicesMap := services.NewServicesMap("postgres").
    WithPriority("postgres", 1)

manager := services.NewManager(servicesMap, services.DefaultManagerConfig())
env := gtt.NewEnvWithManager(gtt.EnvConfig{}, manager)
```

**Accessing services:**
```go
// Typed getters - no type assertion needed!
pg := env.MustGetPostgres()  // panics on error
// Or with error handling:
pg, err := env.GetPostgres()
if err != nil {
    return err
}
// Use pg.DBHost, pg.DBPort, etc.
```

## Integration with User Projects

This is a library meant to be imported by other Go projects. Users:
1. Import `github.com/Educentr/goat` (aliased as `gtt`)
2. Create `main_test.go` with `TestMain()` that initializes env
3. Create app-specific executor function that maps container addresses to app env vars
4. Create `prepareFlow()` helper that configures mocks and returns cleanup function
5. Write test functions that use the flow pattern

See README.md for detailed integration examples.
