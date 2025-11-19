# GOAT = Go Application Testing

## What is it?
This is framework for integration testing of golang applications. It is based on testcontainers (https://golang.testcontainers.org/) and gomock.
It allows to start your app and all dependencies in docker containers and mock all external services.
It also provides some useful tools to simplify testing.
Mocks should be generated over protobuf files and swagger files (using ogen). But for some services there are already prepared mocks.
Some of complex cases are crafted manually because of tech debt.

## ✨ Key Features

- **Context support**: Full context propagation for proper cancellation control
- **Typed getters**: No type assertions! Use `env.GetPostgres()`, `env.GetRedis()`, etc.
- **ExecutorBuilder**: Fluent API for building executors with `NewExecutorBuilder(binary).WithEnv(...).Build()`
- **Flow with callbacks**: Direct callback parameters in Start/Stop methods with testify/require for assertions
- **Service restart**: `manager.Restart(ctx, "postgres")` and `manager.RestartAll(ctx)`
- **Better error handling**: Uses testify/require for test assertions

## Requirements:
 - docker
 - golang

## Steps to attach framework to your project

### 1. Create main_test.go with such content:

```golang
import(
    "testing"
    "time"

    gtt "github.com/Educentr/goat/tools"
)

var env *gtt.Env

func init() {
	// Simple way: just list the services you need
	env = gtt.NewEnv(gtt.EnvConfig{}, []string{"postgres", "redis" })
}

func TestMain(m *testing.M) { // This function is called by go test, once when tests started in package.
	gtt.CallMain(env, m)
}
```

**Available services**: postgres, redis, clickhouse, s3, jaeger, minio, victoriametrics

**Advanced configuration** (if you need custom settings):

```golang
import (
    "github.com/Educentr/goat/tools"
    "github.com/Educentr/goat/tools/services"
    testcontainers "github.com/testcontainers/testcontainers-go"
)

func init() {
    // Use Builder for advanced configuration
    builder := services.NewBuilder().
        WithPostgres(testcontainers.WithImage("postgres:15")). // Custom postgres version
        WithRedis().
        WithLogger(services.NewDefaultLogger()). // Enable logging
        WithMaxParallel(5) // Start up to 5 services in parallel

    env = gtt.NewEnvWithBuilder(gtt.EnvConfig{}, builder)
}
```

### 2. Prepare app specific settings

```golang
func NewApp(env *gtt.Env) *gtt.Executor {
	// Use typed getters - no type assertions needed!
	pg := env.MustGetPostgres()
	rd := env.MustGetRedis()

	m := map[string]string{ // example of environment variables that will be passed to your app
		"YOUR_PROJECT_PREFIX_DB_HOST":     pg.DBHost,
		"YOUR_PROJECT_PREFIX_DB_NAME":     pg.DBName,
		"YOUR_PROJECT_PREFIX_DB_PASSWORD": pg.DBPass,
		"YOUR_PROJECT_PREFIX_DB_PORT":     pg.DBPort,
		"YOUR_PROJECT_PREFIX_DB_USER":     pg.DBUser,
		"YOUR_PROJECT_PREFIX_REDIS_ADDR":  rd.Address,
	}

	p := os.Getenv("YOUR_PROJECT_PREFIX_BINARY") // path to your app binary
	if p == "" {
		p = "/tmp/your_app_binary"
	}

	// Option 1: Simple way
	return gtt.NewExecutor(p, m)

	// Option 2: Builder pattern (more flexible)
	return gtt.NewExecutorBuilder(p).
		WithEnv(m).
		WithOutputFile("/tmp/app.log").
		Build()
}
```

### 3. Configure your mocks and app flow starter

```golang
type Mocks struct {
    SomeApiMock *someapimock.MockHandler
}

func prepareFlow(t *testing.T, env *gtt.Env) (Mocks, func()) {
	var m Mocks

	// Create flow
	f := gtt.NewFlow(
		t,
		env,
		NewApp(env),
		func(server *http.ServeMux, ctl *gomock.Controller) {
			m.SomeApiMock = someapimock.SomeApiV4ManualMockHandler(server, ctl)
		},
		nil, // gRPC callback (nil if not needed)
	)

	// Start with before/after callbacks
	f.Start(t,
		func(env *gtt.Env) error {
			// Callback before app start
			return YourFunctionToPrepareApp(env)
		},
		func(env *gtt.Env) error {
			// Callback after app start
			return prepareApp(env)
		},
	)

	stop := func() {
		f.Stop(t,
			nil, // before stop callback
			func(env *gtt.Env) error {
				// Cleanup after test - use typed getter!
				pg := env.MustGetPostgres()
				s, err := pg.SQL()
				if err != nil {
					return err
				}
				_, err = s.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;")
				return err
			},
		)
	}

	return m, stop
}
```

### 4. Create new file and write your tests

```golang
func TestSome(t *testing.T) { // Example of test
    m, stop := prepareFlow(t, env) // Prepare test flow
    defer stop()         // Stop flow after test is finished

	m.SomeApiMock.EXPECT().GetSome(gomock.Any(), uint64(99)).Return(someApi, nil).MinTimes(1) // Set expectations for mocks
	m.IeMock.EXPECT().GetSome(gomock.Any(), gomock.Any()).Return(&ieSome, nil).MinTimes(1) // Set expectations for mocks
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10) // Context with timeout is needed to stop test if it is stuck
	defer cancel()

	cl := yourAppGrpcClient(ctx, t, appAddress) // Create grpc client to your app
	rsp, err := cl.GetSome(ctx, &grpcApi.GetCurrenciesRequest{ // Call your app grpc endpoint with request
		SomeId:          99,
	})
	assert.NoError(t, err) // Check if there is no error
	assert.Len(t, rsp.GetSome(), 3) // Check if response is as expected
}

```

### 5. Add steps in your makefile to build app

```makefile
#	...
.PHONY: build_for_test
build_for_test:
	go build -cover -race -tags '${TAGS}' ${LDFLAGS} -o /tmp/${BINARY_NAME} ${PACKAGE}/cmd/${BINARY_NAME}
#	...
```

### 6. Run your tests

You can start it directly from your IDE. Or you can run it from command line
```bash
go test .
```

