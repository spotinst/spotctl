package ocean

import (
	"fmt"

	olog "github.com/spotinst/ocean-operator/pkg/log"
	"github.com/theckman/yacspin"
)

type spinnerLogger struct {
	name    string
	spinner *yacspin.Spinner
}

func newSpinnerLogger(name string, spinner *yacspin.Spinner) olog.Logger {
	return &spinnerLogger{
		name:    name,
		spinner: spinner,
	}
}

func (s spinnerLogger) Enabled() bool {
	return s.spinner != nil
}

func (s spinnerLogger) Info(msg string, keysAndValues ...interface{}) {
	s.spinner.Message(fmt.Sprintf("%s   %s", s.name, msg))
}

func (s spinnerLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	s.spinner.Message(fmt.Sprintf("%s   error: %s %s, %v", err.Error(), s.name, msg, keysAndValues))
}

func (s spinnerLogger) V(level int) olog.Logger {
	return s
}

func (s spinnerLogger) WithValues(keysAndValues ...interface{}) olog.Logger {
	return s
}

func (s spinnerLogger) WithName(name string) olog.Logger {
	s.name = name
	return s
}
