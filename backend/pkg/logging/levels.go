package logging

import (
	"os"
	"strings"

	"gorm.io/gorm/logger"
)

// ResolveLevel returns the effective log level: LOG_LEVEL env, then config, then info.
func ResolveLevel(configLevel string) string {
	if v := strings.TrimSpace(os.Getenv("LOG_LEVEL")); v != "" {
		return strings.ToLower(v)
	}
	if l := strings.TrimSpace(configLevel); l != "" {
		return strings.ToLower(l)
	}
	return "info"
}

// GormLogLevel maps application log level to GORM SQL verbosity.
// info → slow queries + errors only; debug → all SQL; warn/error → errors or silent.
func GormLogLevel(appLevel string) logger.LogLevel {
	switch strings.ToLower(appLevel) {
	case "debug":
		return logger.Info
	case "info":
		return logger.Warn
	case "warn", "warning":
		return logger.Error
	default:
		return logger.Silent
	}
}

// ShouldLogRequest reports whether an HTTP access line should be emitted.
func ShouldLogRequest(appLevel string, status int) bool {
	switch strings.ToLower(appLevel) {
	case "debug":
		return true
	case "info":
		return status >= 400
	case "warn", "warning":
		return status >= 500
	default:
		return status >= 500
	}
}
