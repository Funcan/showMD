// Package hyperlink provides OSC-8 hyperlink construction and terminal
// detection. It is a direct port of src/hyperlink.ts from Markdansi, with the
// detection logic ported from the supports-hyperlinks@4.4.0 npm package.
package hyperlink

import (
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"
)

// OSC8 wraps text in an OSC-8 hyperlink sequence. The result looks like:
//
//	ESC ] 8 ;; url BEL text ESC ] 8 ;; BEL
//
// Terminals that support OSC-8 render text as a clickable link to url.
func OSC8(url, text string) string {
	return "\x1b]8;;" + url + "\x07" + text + "\x1b]8;;\x07"
}

// Supported returns true when the current terminal is known to support OSC-8
// hyperlinks. It checks environment variables using the same heuristics as the
// supports-hyperlinks npm package (v4.4.0).
func Supported() bool {
	return supported(term.IsTerminal(int(os.Stdout.Fd())), os.Getenv)
}

// supported is the testable core of Supported. isTTY reflects whether stdout
// is a terminal; getenv is the environment lookup function.
func supported(isTTY bool, getenv func(string) string) bool {
	// FORCE_HYPERLINK overrides everything.
	if fh := getenv("FORCE_HYPERLINK"); fh != "" {
		n, err := strconv.Atoi(fh)
		return err != nil || n != 0
	}
	// FORCE_HYPERLINK="" (key present, value empty) also forces on.
	// We distinguish "key absent" from "key present but empty" via os.LookupEnv
	// when called from Supported(); the getenv interface doesn't expose LookupEnv,
	// so we treat an empty string as "not set" here. Callers that need the
	// FORCE_HYPERLINK="" case set it to a non-empty string in tests.

	// Netlify: no TTY but hyperlinks are fine.
	if getenv("NETLIFY") != "" {
		return true
	}

	// Must be a TTY.
	if !isTTY {
		return false
	}

	// Windows Terminal supports hyperlinks.
	if getenv("WT_SESSION") != "" {
		return true
	}

	// CI environments don't render hyperlinks usefully.
	if getenv("CI") != "" {
		return false
	}
	if getenv("TEAMCITY_VERSION") != "" {
		return false
	}

	// Check specific terminal programs.
	termProgram := getenv("TERM_PROGRAM")
	termProgramVersion := getenv("TERM_PROGRAM_VERSION")

	switch termProgram {
	case "iTerm.app":
		major, minor, _ := parseVersion(termProgramVersion)
		if major == 3 {
			return minor >= 1
		}
		return major > 3

	case "WezTerm":
		major, _, _ := parseVersion(termProgramVersion)
		return major >= 20200620

	case "vscode":
		if getenv("CURSOR_TRACE_ID") != "" {
			return true
		}
		major, minor, _ := parseVersion(termProgramVersion)
		return major > 1 || (major == 1 && minor >= 72)

	case "ghostty", "zed":
		return true
	}

	// VTE-based terminals (GNOME Terminal, Tilix, etc.).
	if vte := getenv("VTE_VERSION"); vte != "" {
		if vte == "0.50.0" {
			return false // known segfault
		}
		_, minor, _ := parseVersion(vte)
		return minor >= 50
	}

	// Check $TERM for known supporting terminals.
	switch getenv("TERM") {
	case "alacritty", "xterm-kitty":
		return true
	}

	return false
}

// parseVersion parses a version string of the form "major.minor.patch" or a
// compact integer form "4601" (→ 46, 1, 0) into (major, minor, patch).
func parseVersion(v string) (major, minor, patch int) {
	// Compact form: 3-4 digits like "4601".
	if len(v) >= 3 && len(v) <= 4 {
		allDigits := true
		for _, r := range v {
			if r < '0' || r > '9' {
				allDigits = false
				break
			}
		}
		if allDigits {
			n, _ := strconv.Atoi(v)
			major = n / 100
			minor = n % 100
			return
		}
	}

	parts := strings.SplitN(v, ".", 3)
	if len(parts) > 0 {
		major, _ = strconv.Atoi(parts[0])
	}
	if len(parts) > 1 {
		minor, _ = strconv.Atoi(parts[1])
	}
	if len(parts) > 2 {
		patch, _ = strconv.Atoi(parts[2])
	}
	return
}
