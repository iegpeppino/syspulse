package logger

import (
	"log/slog"
	"os"
	"path/filepath"
)

var Logger *slog.Logger

// This logger specifically logs errors occurring while getting system stats
func SysDataLogger() {

	// Create directory for logs if it doesn't exist
	logPath := "./logs/errors/systemstats.log"
	logDir := filepath.Dir(logPath)

	err := os.MkdirAll(logDir, 0755)
	if err != nil {
		slog.Error("Failed to create log directory", "error", err)
		Logger = slog.New(slog.NewJSONHandler(os.Stderr, nil))
		return
	}

	// Configure logger to write to file
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		slog.Error("Failed to open log file", "error", err)
		// Return stderr of file loggin fails
		Logger = slog.New(slog.NewJSONHandler(os.Stderr, nil))
		return
	}

	// Create handler with settings
	handler := slog.NewJSONHandler(
		logFile,
		&slog.HandlerOptions{
			Level: slog.LevelError, // Only log errors and fatal errors
		})

	Logger = slog.New(handler)

}
