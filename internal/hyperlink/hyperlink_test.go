package hyperlink_test

import (
	"strings"
	"testing"

	"github.com/funcan/showmd/internal/hyperlink"
)

// env builds a getenv func from a flat key=value map, returning "" for absent keys.
func env(pairs ...string) func(string) string {
	m := make(map[string]string, len(pairs)/2)
	for i := 0; i+1 < len(pairs); i += 2 {
		m[pairs[i]] = pairs[i+1]
	}
	return func(k string) string { return m[k] }
}

// --- OSC8 ---

func TestOSC8_Structure(t *testing.T) {
	result := hyperlink.OSC8("https://example.com", "click here")
	if !strings.HasPrefix(result, "\x1b]8;;https://example.com\x07") {
		t.Errorf("OSC8: bad opening sequence in %q", result)
	}
	if !strings.Contains(result, "click here") {
		t.Errorf("OSC8: text missing from %q", result)
	}
	if !strings.HasSuffix(result, "\x1b]8;;\x07") {
		t.Errorf("OSC8: bad closing sequence in %q", result)
	}
}

func TestOSC8_EmptyURL(t *testing.T) {
	result := hyperlink.OSC8("", "text")
	if !strings.Contains(result, "text") {
		t.Errorf("OSC8 empty url: text missing from %q", result)
	}
}

func TestOSC8_MatchesMarkdansiEscapes(t *testing.T) {
	url := "https://example.com"
	text := "link"
	got := hyperlink.OSC8(url, text)
	want := "\x1b]8;;" + url + "\x07" + text + "\x1b]8;;\x07"
	if got != want {
		t.Errorf("OSC8 exact bytes:\nwant %q\n got %q", want, got)
	}
}

// --- SupportedWith ---

func TestSupported_NoTTY_ReturnsFalse(t *testing.T) {
	if hyperlink.SupportedWith(false, env()) {
		t.Error("non-TTY stdout should return false")
	}
}

func TestSupported_ForceHyperlinkZero(t *testing.T) {
	if hyperlink.SupportedWith(true, env("FORCE_HYPERLINK", "0")) {
		t.Error("FORCE_HYPERLINK=0 should disable hyperlinks")
	}
}

func TestSupported_ForceHyperlinkOne(t *testing.T) {
	if !hyperlink.SupportedWith(false, env("FORCE_HYPERLINK", "1")) {
		t.Error("FORCE_HYPERLINK=1 should enable hyperlinks even without TTY")
	}
}

func TestSupported_CI_ReturnsFalse(t *testing.T) {
	if hyperlink.SupportedWith(true, env("CI", "true")) {
		t.Error("CI environment should return false")
	}
}

func TestSupported_ITerm_3_1_Supported(t *testing.T) {
	e := env("TERM_PROGRAM", "iTerm.app", "TERM_PROGRAM_VERSION", "3.1.0")
	if !hyperlink.SupportedWith(true, e) {
		t.Error("iTerm.app 3.1.0 should support hyperlinks")
	}
}

func TestSupported_ITerm_3_0_NotSupported(t *testing.T) {
	e := env("TERM_PROGRAM", "iTerm.app", "TERM_PROGRAM_VERSION", "3.0.0")
	if hyperlink.SupportedWith(true, e) {
		t.Error("iTerm.app 3.0.0 should NOT support hyperlinks")
	}
}

func TestSupported_ITerm_4_0_Supported(t *testing.T) {
	e := env("TERM_PROGRAM", "iTerm.app", "TERM_PROGRAM_VERSION", "4.0.0")
	if !hyperlink.SupportedWith(true, e) {
		t.Error("iTerm.app 4.0.0 should support hyperlinks")
	}
}

func TestSupported_VSCode_1_72_Supported(t *testing.T) {
	e := env("TERM_PROGRAM", "vscode", "TERM_PROGRAM_VERSION", "1.72.0")
	if !hyperlink.SupportedWith(true, e) {
		t.Error("vscode 1.72.0 should support hyperlinks")
	}
}

func TestSupported_VSCode_1_71_NotSupported(t *testing.T) {
	e := env("TERM_PROGRAM", "vscode", "TERM_PROGRAM_VERSION", "1.71.0")
	if hyperlink.SupportedWith(true, e) {
		t.Error("vscode 1.71.0 should NOT support hyperlinks")
	}
}

func TestSupported_VSCode_Cursor(t *testing.T) {
	e := env("TERM_PROGRAM", "vscode", "TERM_PROGRAM_VERSION", "0.1.0", "CURSOR_TRACE_ID", "abc")
	if !hyperlink.SupportedWith(true, e) {
		t.Error("Cursor (vscode fork with CURSOR_TRACE_ID) should support hyperlinks")
	}
}

func TestSupported_Ghostty(t *testing.T) {
	e := env("TERM_PROGRAM", "ghostty")
	if !hyperlink.SupportedWith(true, e) {
		t.Error("ghostty should support hyperlinks")
	}
}

func TestSupported_Alacritty(t *testing.T) {
	e := env("TERM", "alacritty")
	if !hyperlink.SupportedWith(true, e) {
		t.Error("alacritty should support hyperlinks")
	}
}

func TestSupported_Kitty(t *testing.T) {
	e := env("TERM", "xterm-kitty")
	if !hyperlink.SupportedWith(true, e) {
		t.Error("xterm-kitty should support hyperlinks")
	}
}

func TestSupported_VTE_0_50_0_NotSupported(t *testing.T) {
	e := env("VTE_VERSION", "0.50.0")
	if hyperlink.SupportedWith(true, e) {
		t.Error("VTE 0.50.0 should NOT support hyperlinks (known segfault)")
	}
}

func TestSupported_VTE_0_51_0_Supported(t *testing.T) {
	e := env("VTE_VERSION", "0.51.0")
	if !hyperlink.SupportedWith(true, e) {
		t.Error("VTE 0.51.0 should support hyperlinks")
	}
}

func TestSupported_UnknownTerminal_ReturnsFalse(t *testing.T) {
	e := env("TERM", "xterm-256color")
	if hyperlink.SupportedWith(true, e) {
		t.Error("generic xterm-256color should return false (unknown support)")
	}
}
