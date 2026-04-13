package config

import "log/slog"

// logLevel is the live slog level var. main.go attaches it to the default
// handler at startup; ApplyLogLevel switches between Debug and Info based on
// the DebugLogging setting without needing a restart.
var logLevel = new(slog.LevelVar)

// LogLevel returns the shared slog LevelVar. main.go wires this into the
// default handler; callers should not mutate it directly — use ApplyLogLevel.
func LogLevel() *slog.LevelVar { return logLevel }

// ApplyLogLevel sets slog verbosity from the DebugLogging setting. Safe to
// call at any time; the handler picks up the new level on the next log call.
func ApplyLogLevel(debug bool) {
	if debug {
		logLevel.Set(slog.LevelDebug)
	} else {
		logLevel.Set(slog.LevelInfo)
	}
}
