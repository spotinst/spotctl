package errors

import (
	"fmt"
)

// Error provides a way to return detailed information for an error.
type Error struct {
	// Type defines an error type, expressed as a string value.
	Type ErrorType `json:"type"`

	// Detail defines a human-readable explanation specific to the error.
	Detail string `json:"detail"`
}

// ErrorType represents the type of an error.
type ErrorType string

const (
	// ErrorTypeInternal is used to represent an error that occurs when an internal
	// error is thrown during the command execution.
	ErrorTypeInternal ErrorType = "InternalError"

	// ErrorTypeRequired is used to represent an error that occurs when a required
	// parameter is missing.
	ErrorTypeRequired ErrorType = "RequiredError"

	// ErrorTypeInvalid is used to represent an error that occurs when a string
	// parameter is set a malformed value (e.g. failed regex match, too long).
	ErrorTypeInvalid ErrorType = "InvalidError"

	// ErrorTypeRange is used to represent an error that occurs when a numeric
	// parameter is outside of its valid range.
	ErrorTypeRange ErrorType = "RangeError"

	// ErrorTypeNotImplemented is used to represent an error that occurs when
	// an unimplemented method is called.
	ErrorTypeNotImplemented ErrorType = "NotImplementedError"
)

// Error implements the `error` interface.
func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Type, e.Detail)
}

// New generates a custom error.
func New(typ ErrorType, detail string) error {
	return &Error{Type: typ, Detail: detail}
}

// Internal generates an instance representing an error that occurs when
// an internal error is thrown during the command execution.
func Internal(err error) error {
	return New(ErrorTypeInternal, fmt.Sprintf("internal error occurred: %v", err))
}

// Required generates an instance representing an error that occurs when a required
// parameter is missing.
func Required(name string) error {
	return New(ErrorTypeRequired, fmt.Sprintf("missing value for argument: %s", name))
}

// Invalid generates an instance representing an error that occurs when a string
// parameter is set a malformed value (e.g. failed regex match, too long).
func Invalid(name string, value interface{}) error {
	return New(ErrorTypeInvalid, fmt.Sprintf("invalid value for argument: %s=%v", name, value))
}

// Range generates an instance representing an error that occurs when a numeric
// parameter is outside of its valid range.
func Range(name string, value interface{}) error {
	return New(ErrorTypeRange, fmt.Sprintf("invalid range for argument: %s=%v", name, value))
}

// Noninteractive generates an instance representing an error that occurs when
// a command cannot run in non-interactive mode.
func Noninteractive(op string) error {
	return New(ErrorTypeInternal, fmt.Sprintf("cannot run %q in non-interactive mode", op))
}

// NotImplemented generates an instance representing an error that occurs when
// an unimplemented method is called.
func NotImplemented() error {
	return New(ErrorTypeNotImplemented, "not implemented")
}
