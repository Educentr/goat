package testutil

import (
	"context"
	"database/sql"

	gtt "github.com/Educentr/goat"
)

// ServiceConfig describes the configuration of the service under test
type ServiceConfig interface {
	// ServiceName returns the service name (e.g., "xvpnback")
	ServiceName() string

	// APIPort returns the API service port
	APIPort() string

	// SysPort returns the system endpoint port
	SysPort() string
}

// ExecutorBuilder creates an Executor for the test application
type ExecutorBuilder interface {
	// NewExecutor creates an Executor with full configuration
	// env - goat environment with services (postgres, xray, etc.)
	// mockAddress - HTTP mock server address
	NewExecutor(env *gtt.Env, mockAddress string) *gtt.Executor
}

// MigrationRunner applies migrations to the database
type MigrationRunner interface {
	// ApplyMigrations applies all migrations to the database
	// Implementation determines the order and method of applying migrations
	ApplyMigrations(ctx context.Context, db *sql.DB) error
}

// TableCleaner cleans up test data
type TableCleaner interface {
	// CleanupTables cleans tables between tests
	// Implementation determines the cleanup order considering FK constraints
	CleanupTables(ctx context.Context, db *sql.DB) error
}

// ActiveRecordConfig configures ActiveRecord for tests
type ActiveRecordConfig interface {
	// ConfigMap returns a configuration map for ActiveRecord
	ConfigMap(dbHost, dbPort, dbUser, dbPass, dbName string) map[string]interface{}
}

// TestAppConfig combines all configuration interfaces
type TestAppConfig interface {
	ServiceConfig
	ExecutorBuilder
	MigrationRunner
	TableCleaner
	ActiveRecordConfig
}
