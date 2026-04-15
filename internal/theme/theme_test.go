package theme_test

import (
	"strings"
	"testing"

	"github.com/funcan/showmd/internal/theme"
	"github.com/funcan/showmd/internal/types"
)

// --- ByName ---

func TestByName_Default(t *testing.T) {
	th := theme.ByName(types.ThemeDefault)
	if th.Heading.Color != "yellow" {
		t.Errorf("default heading color: want yellow, got %q", th.Heading.Color)
	}
	if !th.Heading.Bold {
		t.Error("default heading should be bold")
	}
}

func TestByName_Dim(t *testing.T) {
	th := theme.ByName(types.ThemeDim)
	if th.Heading.Color != "white" {
		t.Errorf("dim heading color: want white, got %q", th.Heading.Color)
	}
	if !th.Heading.Dim {
		t.Error("dim heading should be dim")
	}
}

func TestByName_Bright(t *testing.T) {
	th := theme.ByName(types.ThemeBright)
	if th.Heading.Color != "magenta" {
		t.Errorf("bright heading color: want magenta, got %q", th.Heading.Color)
	}
	if th.Link.Color != "cyan" {
		t.Errorf("bright link color: want cyan, got %q", th.Link.Color)
	}
}

func TestByName_Solarized(t *testing.T) {
	th := theme.ByName(types.ThemeSolarized)
	if th.BlockCode.Color != "#2aa198" {
		t.Errorf("solarized blockCode color: want #2aa198, got %q", th.BlockCode.Color)
	}
}

func TestByName_Monochrome(t *testing.T) {
	th := theme.ByName(types.ThemeMonochrome)
	if th.Heading.Color != "" {
		t.Errorf("monochrome heading should have no color, got %q", th.Heading.Color)
	}
	if !th.Heading.Bold {
		t.Error("monochrome heading should be bold")
	}
}

func TestByName_Contrast(t *testing.T) {
	th := theme.ByName(types.ThemeContrast)
	if th.Heading.Color != "magenta" {
		t.Errorf("contrast heading color: want magenta, got %q", th.Heading.Color)
	}
	if th.TableCell.Color != "white" {
		t.Errorf("contrast tableCell color: want white, got %q", th.TableCell.Color)
	}
}

func TestByName_Unknown_FallsBackToDefault(t *testing.T) {
	th := theme.ByName("nonexistent")
	if th.Heading.Color != "yellow" {
		t.Errorf("unknown theme should fall back to default (yellow heading), got %q", th.Heading.Color)
	}
}

// --- Resolve: inlineCode/blockCode fallback ---

func TestResolve_CodeFallback(t *testing.T) {
	customCode := types.StyleIntent{Color: "red"}
	opts := types.RenderOptions{
		Theme: &types.Theme{
			Code: customCode,
			// InlineCode and BlockCode intentionally zero
		},
	}
	th := theme.Resolve(opts)
	if th.InlineCode.Color != "red" {
		t.Errorf("InlineCode should fall back to Code color (red), got %q", th.InlineCode.Color)
	}
	if th.BlockCode.Color != "red" {
		t.Errorf("BlockCode should fall back to Code color (red), got %q", th.BlockCode.Color)
	}
}

func TestResolve_ExplicitInlineCodeNotOverridden(t *testing.T) {
	opts := types.RenderOptions{
		Theme: &types.Theme{
			Code:       types.StyleIntent{Color: "red"},
			InlineCode: types.StyleIntent{Color: "cyan"},
		},
	}
	th := theme.Resolve(opts)
	if th.InlineCode.Color != "cyan" {
		t.Errorf("explicit InlineCode should not be overridden by Code fallback, got %q", th.InlineCode.Color)
	}
}

// --- Merge ---

func TestMerge_OverrideFields(t *testing.T) {
	base := theme.Default
	override := types.Theme{
		Heading: types.StyleIntent{Color: "red", Bold: true},
	}
	merged := theme.Merge(base, override)
	if merged.Heading.Color != "red" {
		t.Errorf("merged heading color: want red, got %q", merged.Heading.Color)
	}
	// Other fields should be unchanged from base.
	if merged.Link.Color != base.Link.Color {
		t.Errorf("merge should not change link color, want %q got %q", base.Link.Color, merged.Link.Color)
	}
}

// --- Styler ---

func TestStyler_NoColor_IsNoop(t *testing.T) {
	s := theme.NewStyler(false)
	result := s.Apply("hello", types.StyleIntent{Color: "red", Bold: true})
	if result != "hello" {
		t.Errorf("no-color styler should return text unchanged, got %q", result)
	}
}

func TestStyler_Color_ContainsEscapes(t *testing.T) {
	s := theme.NewStyler(true)
	result := s.Apply("hello", types.StyleIntent{Color: "red"})
	if !strings.Contains(result, "hello") {
		t.Error("styled output should contain original text")
	}
	if !strings.Contains(result, "\x1b[") {
		t.Errorf("color styler should emit ANSI escapes, got %q", result)
	}
}

func TestStyler_Bold(t *testing.T) {
	s := theme.NewStyler(true)
	result := s.Apply("x", types.StyleIntent{Bold: true})
	if !strings.Contains(result, "\x1b[") {
		t.Errorf("bold styler should emit ANSI escapes, got %q", result)
	}
}

func TestStyler_HexColor(t *testing.T) {
	s := theme.NewStyler(true)
	// Should not panic and should produce output containing the text.
	result := s.Apply("x", types.StyleIntent{Color: "#2aa198"})
	if !strings.Contains(result, "x") {
		t.Errorf("hex color output should contain original text, got %q", result)
	}
}

func TestStyler_EmptyIntent_NoChange(t *testing.T) {
	s := theme.NewStyler(true)
	result := s.Apply("plain", types.StyleIntent{})
	if !strings.Contains(result, "plain") {
		t.Errorf("empty intent should still contain text, got %q", result)
	}
}
