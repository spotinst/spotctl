package log

import (
	"sync"
)

var (
	// defaultLogger is the default logger that only logs at the default level.
	// It may be overwritten by calling to `InitDefaultLogger` with options.
	defaultLogger         Logger
	defaultLoggerInitOnce sync.Once
)

func InitDefaultLogger(opts ...LoggerOption) {
	if defaultLogger == nil {
		defaultLogger = NewLogrusLogger(opts...)
	}
}

func DefaultLogger() Logger { return defaultLogger }

func initDefaultLogger() { InitDefaultLogger() }

func Debugf(format string, args ...interface{}) {
	defaultLoggerInitOnce.Do(initDefaultLogger)
	defaultLogger.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	defaultLoggerInitOnce.Do(initDefaultLogger)
	defaultLogger.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	defaultLoggerInitOnce.Do(initDefaultLogger)
	defaultLogger.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	defaultLoggerInitOnce.Do(initDefaultLogger)
	defaultLogger.Errorf(format, args...)
}
