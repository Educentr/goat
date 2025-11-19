package services

import (
	"fmt"
	"io"
	"os"
	"time"
)

// LogLevel represents the severity of a log message.
type LogLevel int

const (
	// DebugLevel is for debug messages
	DebugLevel LogLevel = iota
	// InfoLevel is for informational messages
	InfoLevel
	// WarnLevel is for warning messages
	WarnLevel
	// ErrorLevel is for error messages
	ErrorLevel
)

func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// DefaultLogger is a simple logger that writes to stdout/stderr.
type DefaultLogger struct {
	out      io.Writer
	minLevel LogLevel
}

// NewDefaultLogger creates a new default logger with INFO level.
func NewDefaultLogger() *DefaultLogger {
	return &DefaultLogger{
		minLevel: InfoLevel,
		out:      os.Stdout,
	}
}

// NewDefaultLoggerWithLevel creates a new default logger with the specified level.
func NewDefaultLoggerWithLevel(level LogLevel) *DefaultLogger {
	return &DefaultLogger{
		minLevel: level,
		out:      os.Stdout,
	}
}

// SetLevel sets the minimum log level.
func (l *DefaultLogger) SetLevel(level LogLevel) {
	l.minLevel = level
}

// SetOutput sets the output writer.
func (l *DefaultLogger) SetOutput(w io.Writer) {
	l.out = w
}

func (l *DefaultLogger) log(level LogLevel, msg string, keysAndValues ...interface{}) {
	if level < l.minLevel {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	fmt.Fprintf(l.out, "[%s] %s: %s", timestamp, level.String(), msg)

	if len(keysAndValues) > 0 {
		fmt.Fprint(l.out, " |")
		for i := 0; i < len(keysAndValues); i += 2 {
			if i+1 < len(keysAndValues) {
				fmt.Fprintf(l.out, " %v=%v", keysAndValues[i], keysAndValues[i+1])
			} else {
				fmt.Fprintf(l.out, " %v=<missing>", keysAndValues[i])
			}
		}
	}

	fmt.Fprintln(l.out)
}

// Debug logs a debug message.
func (l *DefaultLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.log(DebugLevel, msg, keysAndValues...)
}

// Info logs an info message.
func (l *DefaultLogger) Info(msg string, keysAndValues ...interface{}) {
	l.log(InfoLevel, msg, keysAndValues...)
}

// Warn logs a warning message.
func (l *DefaultLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.log(WarnLevel, msg, keysAndValues...)
}

// Error logs an error message.
func (l *DefaultLogger) Error(msg string, keysAndValues ...interface{}) {
	l.log(ErrorLevel, msg, keysAndValues...)
}

// NoopLogger is a logger that does nothing.
type NoopLogger struct{}

// NewNoopLogger creates a new noop logger.
func NewNoopLogger() *NoopLogger {
	return &NoopLogger{}
}

// Debug does nothing.
func (l *NoopLogger) Debug(_ string, _ ...interface{}) {}

// Info does nothing.
func (l *NoopLogger) Info(_ string, _ ...interface{}) {}

// Warn does nothing.
func (l *NoopLogger) Warn(_ string, _ ...interface{}) {}

// Error does nothing.
func (l *NoopLogger) Error(_ string, _ ...interface{}) {}
