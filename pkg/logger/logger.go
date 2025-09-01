package logger

import (
	"os"

	"github.com/rs/zerolog"
)

// Logger is the global logger instance
var Logger zerolog.Logger

func init() {
	Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
}

// Info logs an info message
func Info() *zerolog.Event {
	return Logger.Info()
}

// Error logs an error message
func Error() *zerolog.Event {
	return Logger.Error()
}

// Warn logs a warning message
func Warn() *zerolog.Event {
	return Logger.Warn()
}

// Debug logs a debug message
func Debug() *zerolog.Event {
	return Logger.Debug()
}
