package errors

import (
	"fmt"
	"strings"
	"sync"

	"github.com/hashicorp/go-multierror"
)

// ErrorGroup is an error type to track multiple errors. This is used to
// accumulate errors in cases and return them as a single error.
type ErrorGroup struct {
	errMu sync.Mutex
	err   multierror.Error
}

// NewErrorGroup returns a new error group.
func NewErrorGroup() *ErrorGroup {
	return &ErrorGroup{
		err: multierror.Error{
			ErrorFormat: listFormatFunc,
		},
	}
}

// Add adds a new error to the group.
func (e *ErrorGroup) Add(err error) {
	e.errMu.Lock()
	defer e.errMu.Unlock()

	e.err.Errors = append(e.err.Errors, err)
}

// Len implements sort.Interface function for length.
func (e *ErrorGroup) Len() int {
	e.errMu.Lock()
	defer e.errMu.Unlock()

	return e.err.Len()
}

// Swap implements sort.Interface function for swapping elements.
func (e *ErrorGroup) Swap(i, j int) {
	e.errMu.Lock()
	defer e.errMu.Unlock()

	e.err.Swap(i, j)
}

// Less implements sort.Interface function for determining order.
func (e *ErrorGroup) Less(i, j int) bool {
	e.errMu.Lock()
	defer e.errMu.Unlock()

	return e.err.Less(i, j)
}

// Error implements the internal error interface.
func (e *ErrorGroup) Error() string {
	e.errMu.Lock()
	defer e.errMu.Unlock()

	return e.err.Error()
}

// Errors returns the list of errors that this group is wrapping.
func (e *ErrorGroup) Errors() []error {
	e.errMu.Lock()
	defer e.errMu.Unlock()

	errors := make([]error, e.err.Len())
	copy(errors, e.err.Errors)

	return errors
}

// listFormatFunc is a custom error formatter used by multierror.Error.
func listFormatFunc(errs []error) string {
	if len(errs) == 0 {
		return ""
	}

	if len(errs) == 1 {
		return fmt.Sprintf("1 error occurred:\n * %s\n", errs[0])
	}

	points := make([]string, len(errs))
	for i, err := range errs {
		points[i] = fmt.Sprintf("* %s", err)
	}

	return fmt.Sprintf("%d errors occurred:\n %s\n",
		len(errs), strings.Join(points, "\n "))
}
