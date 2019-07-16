package survey

import (
	"errors"
)

// ErrNotImplemented is the error returned if a method is not implemented.
var ErrNotImplemented = errors.New("survey: not implemented")

type (
	// Interface defines the interface of a survey prompter.
	Interface interface {
		// InputString gets an answer to a prompt from a user's free-form input.
		InputString(input *Input) (string, error)

		// InputInt gets an answer to a prompt from a user's free-form input.
		InputInt64(input *Input) (int64, error)

		// InputFloat64 gets an answer to a prompt from a user's free-form input.
		InputFloat64(input *Input) (float64, error)

		// Password gets a password (via hidden input) from a user's free-form input.
		Password(input *Input) (string, error)

		// Confirm prompts the user to confirm something.
		Confirm(input *Input) (bool, error)

		// Select gets the user to select an option from a list of options.
		Select(input *Select) (string, error)

		// SelectMulti gets the user to select multiple selections from a list of options.
		SelectMulti(input *Select) ([]string, error)
	}

	// Input holds the configuration for input prompt.
	Input struct {
		Message   string
		Help      string
		Default   interface{}
		Required  bool
		Validate  Validator
		Transform Transformer
	}

	// Select holds the configuration for both select and multi-select prompts.
	Select struct {
		Message   string
		Help      string
		Defaults  []interface{}
		Options   []interface{}
		Validate  Validator
		Transform Transformer
	}

	// Validator is a function used to validate a response. If the function
	// returns an error, then the user will be prompted again for another response.
	Validator func(answer interface{}) error

	// Transformer is a function used to implement a custom logic that will
	// result to return a different representation of the given answer.
	Transformer func(answer interface{}) (newAnswer interface{})
)
