package log

import (
	"time"

	"github.com/sirupsen/logrus"
)

type logrusLogger struct {
	logger *logrus.Logger
}

func NewLogrusLogger(options ...LoggerOption) Logger {
	opts := initDefaultOptions()
	for _, opt := range options {
		opt(opts)
	}

	// Initialize a new logger.
	logger := logrus.New()

	// Set the output data stream.
	logger.SetOutput(opts.Out)

	// Set the logger verbosity level.
	var level logrus.Level
	switch opts.Level {
	case LevelDebug:
		level = logrus.DebugLevel
	case LevelInfo:
		level = logrus.InfoLevel
	case LevelWarn:
		level = logrus.WarnLevel
	case LevelError:
		level = logrus.ErrorLevel
	}
	logger.SetLevel(level)

	// Set the logger format.
	var formatter logrus.Formatter
	switch opts.Format {
	case FormatText:
		formatter = &logrus.TextFormatter{
			TimestampFormat: time.RFC3339,
			FullTimestamp:   true,
		}
	case FormatJSON:
		formatter = &logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "@timestamp",
				logrus.FieldKeyLevel: "@level",
				logrus.FieldKeyMsg:   "@message",
			},
		}
	}
	if formatter != nil {
		logger.SetFormatter(formatter)
	}

	return &logrusLogger{
		logger: logger,
	}
}

func (x *logrusLogger) Debugf(format string, args ...interface{}) {
	x.logger.Debugf(format, args...)
}

func (x *logrusLogger) Infof(format string, args ...interface{}) {
	x.logger.Infof(format, args...)
}

func (x *logrusLogger) Warnf(format string, args ...interface{}) {
	x.logger.Warnf(format, args...)
}

func (x *logrusLogger) Errorf(format string, args ...interface{}) {
	x.logger.Errorf(format, args...)
}
