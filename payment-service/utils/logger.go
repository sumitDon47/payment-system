package utils

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Logger is a global logger instance
var Logger = initLogger()

func initLogger() *logrus.Logger {
	logger := logrus.New()

	// Set output to stdout
	logger.SetOutput(os.Stdout)

	// Set JSON format for structured logging
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		PrettyPrint:     false,
	})

	// Set log level
	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		level = "info"
	}
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logger.SetLevel(logrus.InfoLevel)
	} else {
		logger.SetLevel(logLevel)
	}

	return logger
}

// Error logs an error with context
func Error(msg string, err error, context map[string]interface{}) {
	if context == nil {
		context = make(map[string]interface{})
	}
	context["error"] = err.Error()
	Logger.WithFields(logrus.Fields(context)).Error(msg)
}

// Info logs an info message with context
func Info(msg string, context map[string]interface{}) {
	Logger.WithFields(logrus.Fields(context)).Info(msg)
}

// Warn logs a warning with context
func Warn(msg string, context map[string]interface{}) {
	Logger.WithFields(logrus.Fields(context)).Warn(msg)
}

// Debug logs a debug message with context
func Debug(msg string, context map[string]interface{}) {
	Logger.WithFields(logrus.Fields(context)).Debug(msg)
}
