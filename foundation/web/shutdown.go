package web

import (
	"errors"
)

// shutdownError is a type used to help with the graceful termination of the service.
type shutdownError struct {
	Message string
}

// RequestError is used to pass an error during the request through the
// application with web specific context.
type RequestError struct {
	Err    error
	Status int
}

// NewRequestError wraps a provided error with an HTTP status code. This
// function should be used when handlers encounter expected errors.
func NewRequestError(err error, status int) error {
	return &RequestError{err, status}
}

// Error implements the error interface. It uses the default message of the
// wrapped error. This is what will be shown in the services' logs.
func (re *RequestError) Error() string {
	return re.Err.Error()
}

// NewShutdownError returns an error that causes the framework to signal
// a graceful shutdown.
func NewShutdownError(message string) error {
	return &shutdownError{message}
}

// Error is the implementation of the error interface.
func (se *shutdownError) Error() string {
	return se.Message
}

// IsShutdown checks to see if the shutdown error is contained
// in the specified error value.
func IsShutdown(err error) bool {
	var se *shutdownError
	return errors.As(err, &se)
}
