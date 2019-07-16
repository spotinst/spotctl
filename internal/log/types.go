package log

type (
	// Logger is a unified interface for logging.
	Logger interface {
		// Debug logs a Debug event.
		Debugf(format string, args ...interface{})

		// Info logs an Info event.
		Infof(format string, args ...interface{})

		// Warn logs a Warn(ing) event.
		Warnf(format string, args ...interface{})

		// Error logs an Error event.
		Errorf(format string, args ...interface{})
	}

	// Level represents the logging level.
	Level uint32

	// Format represents the logging format.
	Format string
)

const (
	// Usually only enabled when debugging. Very verbose logging.
	LevelDebug Level = iota

	// General operational entries about what's going on inside the application.
	LevelInfo

	// Non-critical events that should be looked at.
	LevelWarn

	// Critical events that require immediate attention.
	LevelError
)

const (
	// FormatJSON represents a JSON format.
	FormatJSON Format = "json"

	// FormatJSON represents a text format.
	FormatText Format = "text"
)
