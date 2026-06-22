package devmode

// LocalLoginEnabled reports whether the local password login route is active.
// Local admin bootstrap and login stay enabled in all environments as an SSO fallback.
func LocalLoginEnabled() bool {
	return true
}

// Enabled is kept for compatibility with existing call sites.
func Enabled() bool {
	return LocalLoginEnabled()
}
