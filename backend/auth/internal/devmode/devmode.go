package devmode

import (
	"os"
	"strings"
)

// Enabled reports whether local bootstrap login is allowed.
// Requires DEV_MODE=true and ENV must not be prod.
func Enabled() bool {
	if isProdEnv() {
		return false
	}
	v := strings.TrimSpace(strings.ToLower(os.Getenv("DEV_MODE")))
	return v == "true" || v == "1" || v == "yes" || v == "on"
}

func isProdEnv() bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv("ENV")))
	return v == "prod" || v == "production"
}
