package wave

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/spotinst/spotctl/internal/log"
	"github.com/theckman/yacspin"
)

// wrap the spotctl log into a logr.Logger

type waveLogger struct {
	name string
}

func (w waveLogger) Info(msg string, keysAndValues ...interface{}) {
	log.Infof("%s   %s, %v", w.name, msg, keysAndValues)
}

func (w waveLogger) Enabled() bool {
	return true
}

func (w waveLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	log.Errorf("%s   ERROR %s, %w, %v", w.name, msg, err, keysAndValues)
}

func (w waveLogger) V(level int) logr.InfoLogger {
	// ignore
	return w
}

func (w waveLogger) WithValues(keysAndValues ...interface{}) logr.Logger {
	// ignore
	return w
}

func (w waveLogger) WithName(name string) logr.Logger {
	// ignore
	w.name = name
	return w
}

func getWaveLogger() logr.Logger {
	return &waveLogger{}
}

type spinnerLogger struct {
	name    string
	spinner *yacspin.Spinner
}

func (s spinnerLogger) Enabled() bool {
	return s.spinner != nil
}

func (s spinnerLogger) Info(msg string, keysAndValues ...interface{}) {
	s.spinner.Message(fmt.Sprintf("%s   %s, %v", s.name, msg, keysAndValues))
}

func (s spinnerLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	s.spinner.Message(fmt.Sprintf("%s   ERROR:%s %s, %v", err.Error(), s.name, msg, keysAndValues))
}

func (s spinnerLogger) V(level int) logr.Logger {
	return s
}

func (s spinnerLogger) WithValues(keysAndValues ...interface{}) logr.Logger {
	return s
}

func (s spinnerLogger) WithName(name string) logr.Logger {
	s.name = name
	return s
}

func getSpinnerLogger(name string, spinner *yacspin.Spinner) logr.Logger {
	return &spinnerLogger{name, spinner}
}
