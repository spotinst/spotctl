package log

import (
	"io"
	"os"
)

type LoggerOptions struct {
	// Out specifies the output data stream.
	Out io.Writer

	// Level specifies the logging level the logger should log at. This is
	// typically (and defaults to) `Info`, which allows Info(), Warn() and
	// Error() to be logged.
	Level Level

	// Format specifies the logging format the logger should log at.
	Format Format
}

// LoggerOption sets an optional parameter for loggers.
type LoggerOption func(*LoggerOptions)

// WithOutput sets the logger output.
func WithOutput(out io.Writer) LoggerOption {
	return func(opts *LoggerOptions) {
		opts.Out = out
	}
}

// WithLevel sets the logger verbosity level.
func WithLevel(level Level) LoggerOption {
	return func(opts *LoggerOptions) {
		opts.Level = level
	}
}

// WithFormat sets the logger format.
func WithFormat(format Format) LoggerOption {
	return func(opts *LoggerOptions) {
		opts.Format = format
	}
}

func initDefaultOptions() *LoggerOptions {
	return &LoggerOptions{
		Out:    os.Stdout,
		Level:  LevelInfo,
		Format: FormatText,
	}
}
