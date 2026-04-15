// export_test.go exposes internal symbols for testing only.
package hyperlink

// SupportedWith calls the internal supported function, allowing tests to
// control both the TTY flag and the environment without forking a real process.
func SupportedWith(isTTY bool, getenv func(string) string) bool {
	return supported(isTTY, getenv)
}
