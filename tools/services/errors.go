package services

import "fmt"

// ErrServiceNotFound is returned when a requested service is not found in the registry.
type ErrServiceNotFound struct {
	ServiceName string
}

func (e *ErrServiceNotFound) Error() string {
	return fmt.Sprintf("service %q not found in registry", e.ServiceName)
}

// ErrServiceNotRunning is returned when trying to access a service that is not running.
type ErrServiceNotRunning struct {
	ServiceName string
}

func (e *ErrServiceNotRunning) Error() string {
	return fmt.Sprintf("service %q is not running", e.ServiceName)
}

// ErrServiceStartFailed is returned when a service fails to start.
type ErrServiceStartFailed struct {
	Cause       error
	ServiceName string
}

func (e *ErrServiceStartFailed) Error() string {
	return fmt.Sprintf("failed to start service %q: %v", e.ServiceName, e.Cause)
}

func (e *ErrServiceStartFailed) Unwrap() error {
	return e.Cause
}

// ErrServiceStopFailed is returned when a service fails to stop.
type ErrServiceStopFailed struct {
	Cause       error
	ServiceName string
}

func (e *ErrServiceStopFailed) Error() string {
	return fmt.Sprintf("failed to stop service %q: %v", e.ServiceName, e.Cause)
}

func (e *ErrServiceStopFailed) Unwrap() error {
	return e.Cause
}

// ErrHealthCheckFailed is returned when a service health check fails.
type ErrHealthCheckFailed struct {
	Cause       error
	ServiceName string
}

func (e *ErrHealthCheckFailed) Error() string {
	return fmt.Sprintf("health check failed for service %q: %v", e.ServiceName, e.Cause)
}

func (e *ErrHealthCheckFailed) Unwrap() error {
	return e.Cause
}

// ErrDependencyNotMet is returned when a service dependency is not met.
type ErrDependencyNotMet struct {
	ServiceName    string
	DependencyName string
}

func (e *ErrDependencyNotMet) Error() string {
	return fmt.Sprintf("service %q depends on %q which is not running", e.ServiceName, e.DependencyName)
}

// ErrServiceAlreadyRegistered is returned when trying to register a service that is already registered.
type ErrServiceAlreadyRegistered struct {
	ServiceName string
}

func (e *ErrServiceAlreadyRegistered) Error() string {
	return fmt.Sprintf("service %q is already registered", e.ServiceName)
}

// ErrServiceTypeMismatch is returned when a service container cannot be cast to the expected type.
type ErrServiceTypeMismatch struct {
	ServiceName  string
	ExpectedType string
}

func (e *ErrServiceTypeMismatch) Error() string {
	return fmt.Sprintf("service %q cannot be cast to %s", e.ServiceName, e.ExpectedType)
}
