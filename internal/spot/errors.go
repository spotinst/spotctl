package spot

import (
	"errors"

	"github.com/spotinst/spotinst-sdk-go/spotinst/client"
)

const (
	codeResourceDoesNotExist = "RESOURCE_DOES_NOT_EXIST"
)

func parseError(err error) error {

	sdkError := client.Error{}
	if errors.As(err, &sdkError) {
		ok, wrappedErr := wrapError(sdkError)
		if ok {
			return wrappedErr
		}
	}

	// TODO(thorsteinn) Handle parsing multiple errors from error list
	sdkErrors := client.Errors{}
	if errors.As(err, &sdkErrors) {
		for _, e := range sdkErrors {
			ok, wrappedErr := wrapError(e)
			if ok {
				return wrappedErr
			}
		}
	}

	return err
}

func wrapError(sdkError client.Error) (bool, error) {
	switch sdkError.Code {
	case codeResourceDoesNotExist:
		return true, newResourceDoesNotExistError(sdkError)
	}
	return false, sdkError
}

type ResourceDoesNotExistError struct {
	err error
}

func newResourceDoesNotExistError(err error) ResourceDoesNotExistError {
	return ResourceDoesNotExistError{err: err}
}

func (e ResourceDoesNotExistError) Error() string {
	return e.err.Error()
}

func (e ResourceDoesNotExistError) Unwrap() error {
	return e.err
}
