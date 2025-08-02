package logger

import (
	"log/slog"
	"os"
)

var Logger *slog.Logger

// This logger specifically logs errors occurring while getting system stats
func SysDataLogger() {

	// Configure logger to write to file
	logFile, err := os.OpenFile("./logs/errors/systemstats.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		slog.Error("Failed to open log file", "error", err)
		// Return stderr of file loggin fails
		Logger = slog.New(slog.NewJSONHandler(os.Stderr, nil))
		return
	}
	// Close logfile
	defer logFile.Close()

	// Create handler with settings
	handler := slog.NewJSONHandler(
		logFile,
		&slog.HandlerOptions{
			Level: slog.LevelError, // Only log errors and fatal errors
		})

	Logger = slog.New(handler)

}
