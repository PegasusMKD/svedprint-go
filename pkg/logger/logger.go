package logger

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Setup initializes the global logger
func Setup(level, serviceName string) {
	// Set the log level
	logLevel := parseLevel(level)
	zerolog.SetGlobalLevel(logLevel)

	// Configure pretty logging for development
	if strings.ToLower(os.Getenv("GIN_MODE")) == "debug" {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		})
	} else {
		// JSON logging for production
		log.Logger = zerolog.New(os.Stdout).With().
			Timestamp().
			Str("service", serviceName).
			Logger()
	}
}

// SetupWithWriter initializes a logger with a custom writer
func SetupWithWriter(level, serviceName string, writer io.Writer) zerolog.Logger {
	logLevel := parseLevel(level)

	return zerolog.New(writer).With().
		Timestamp().
		Str("service", serviceName).
		Logger().Level(logLevel)
}

// parseLevel converts a string level to zerolog.Level
func parseLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}

// Get returns the global logger
func Get() *zerolog.Logger {
	return &log.Logger
}
