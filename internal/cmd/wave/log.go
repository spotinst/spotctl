package wave

import (
	"github.com/go-logr/logr"
	"github.com/spotinst/spotctl/internal/log"
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
