package goat

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/Educentr/goat/services"
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

// NewEnv creates a new environment with an existing services.Manager.
// The manager should be created using goat-services package.
//
// Example:
//
//	import gs "github.com/Educentr/goat-services"
//
//	// Option 1: Simple list
//	manager := gs.NewManager([]string{"postgres", "redis"})
//
//	// Option 2: Builder pattern
//	manager := gs.NewBuilder().WithPostgres().WithRedis().BuildManager()
//
//	// Create environment
//	env := NewEnv(EnvConfig{}, manager)
func NewEnv(envConf EnvConfig, manager *services.Manager) *Env {
	return &Env{
		manager: manager,
		Conf:    envConf,
	}
}

// Manager returns the services.Manager instance for direct access to service management.
// Use services.GetTyped[T](manager, name) for type-safe access to services.
//
// Example:
//
//	import "github.com/Educentr/goat-services/psql"
//	import "github.com/Educentr/goat/services"
//	pg, err := services.GetTyped[*psql.Env](env.Manager(), "postgres")
func (e *Env) Manager() *services.Manager {
	return e.manager
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
