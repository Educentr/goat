package goat

import (
	"os"
)

// ExecutorBuilder provides a fluent API for building Executor instances.
//
//nolint:govet // fieldalignment: struct optimization not worth the readability cost
type ExecutorBuilder struct {
	env           map[string]string
	args          []string
	binary        string
	debugPort     string
	outputFile    string
	errorsFile    string
	fieldsFile    string
	debug         bool
	disableStdout bool
}

// NewExecutorBuilder creates a new ExecutorBuilder with the given binary path.
func NewExecutorBuilder(binary string) *ExecutorBuilder {
	return &ExecutorBuilder{
		binary: binary,
		env:    make(map[string]string),
	}
}

// WithEnv sets the environment variables for the executor.
// This replaces any previously set environment variables.
func (b *ExecutorBuilder) WithEnv(env map[string]string) *ExecutorBuilder {
	b.env = env
	return b
}

// WithEnvVar adds a single environment variable.
func (b *ExecutorBuilder) WithEnvVar(key, value string) *ExecutorBuilder {
	if b.env == nil {
		b.env = make(map[string]string)
	}
	b.env[key] = value
	return b
}

// WithArgs sets the command-line arguments for the binary.
func (b *ExecutorBuilder) WithArgs(args ...string) *ExecutorBuilder {
	b.args = args
	return b
}

// WithDebug enables debug mode with delve debugger.
// Default debug port is 2345.
func (b *ExecutorBuilder) WithDebug() *ExecutorBuilder {
	b.debug = true
	return b
}

// WithDebugPort sets a custom debug port (implies WithDebug).
func (b *ExecutorBuilder) WithDebugPort(port string) *ExecutorBuilder {
	b.debug = true
	b.debugPort = port
	return b
}

// WithOutputFile redirects stdout to the specified file.
func (b *ExecutorBuilder) WithOutputFile(path string) *ExecutorBuilder {
	b.outputFile = path
	return b
}

// WithErrorsFile redirects stderr to the specified file.
func (b *ExecutorBuilder) WithErrorsFile(path string) *ExecutorBuilder {
	b.errorsFile = path
	return b
}

// WithFieldsFile enables log field validation using the specified CSV file.
func (b *ExecutorBuilder) WithFieldsFile(path string) *ExecutorBuilder {
	b.fieldsFile = path
	return b
}

// WithDisableStdout disables stdout output.
func (b *ExecutorBuilder) WithDisableStdout(disable bool) *ExecutorBuilder {
	b.disableStdout = disable
	return b
}

// Build creates the Executor with the configured options.
func (b *ExecutorBuilder) Build() *Executor {
	// Set environment variables for configuration
	if b.debug {
		if b.debugPort != "" {
			os.Setenv("GOAT_REMOTE_DEBUG_PORT", b.debugPort)
		}
		os.Setenv("GOAT_REMOTE_DEBUG", TrueValue)
	}

	if b.disableStdout {
		os.Setenv("GOAT_DISABLE_STDOUT", TrueValue)
	}

	if b.outputFile != "" {
		os.Setenv("GOAT_OUTPUT_FILE", b.outputFile)
	}

	if b.errorsFile != "" {
		os.Setenv("GOAT_OUTPUT_ERRORS_FILE", b.errorsFile)
	}

	if b.fieldsFile != "" {
		os.Setenv("GOAT_LOG_FIELDS_FILE", b.fieldsFile)
	}

	// Build the executor using the existing constructor
	return NewExecutor(b.binary, b.env, b.args...)
}
