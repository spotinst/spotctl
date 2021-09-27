package log

import (
	"github.com/go-logr/logr"
)

// logrWrapper wraps the default log into a logr.Logger interface
type logrWrapper struct {
	name string
}

func (w logrWrapper) Info(msg string, keysAndValues ...interface{}) {
	Infof("%s - %s, %v", w.name, msg, keysAndValues)
}

func (w logrWrapper) Enabled() bool {
	return true
}

func (w logrWrapper) Error(err error, msg string, keysAndValues ...interface{}) {
	Errorf("%s - ERROR %s, %w, %v", w.name, msg, err, keysAndValues)
}

func (w logrWrapper) V(level int) logr.Logger {
	// ignore
	return w
}

func (w logrWrapper) WithValues(keysAndValues ...interface{}) logr.Logger {
	// ignore
	return w
}

func (w logrWrapper) WithName(name string) logr.Logger {
	// ignore
	w.name = name
	return w
}

func GetLogrLogger() logr.Logger {
	return &logrWrapper{}
}
