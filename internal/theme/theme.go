// Package theme defines the built-in named themes and the Styler type, which
// applies a StyleIntent to a string using ANSI escape codes via termenv.
//
// This is a direct port of src/theme.ts from Markdansi.
package theme

import (
	"strings"

	"github.com/funcan/showmd/internal/types"
	"github.com/muesli/termenv"
)

// Built-in themes, matching the named themes in src/theme.ts.
var (
	Default = types.Theme{
		Heading:     types.StyleIntent{Color: "yellow", Bold: true},
		Strong:      types.StyleIntent{Bold: true},
		Emph:        types.StyleIntent{Italic: true},
		InlineCode:  types.StyleIntent{Color: "cyan"},
		BlockCode:   types.StyleIntent{Color: "green"},
		Link:        types.StyleIntent{Color: "blue", Underline: true},
		Quote:       types.StyleIntent{Dim: true},
		HR:          types.StyleIntent{Dim: true},
		ListMarker:  types.StyleIntent{Color: "cyan"},
		TableHeader: types.StyleIntent{Bold: true, Color: "yellow"},
		TableCell:   types.StyleIntent{},
	}

	Dim = func() types.Theme {
		t := Default
		t.Heading = types.StyleIntent{Color: "white", Bold: true, Dim: true}
		t.Link = types.StyleIntent{Color: "blue", Underline: true, Dim: true}
		return t
	}()

	Bright = func() types.Theme {
		t := Default
		t.Heading = types.StyleIntent{Color: "magenta", Bold: true}
		t.Link = types.StyleIntent{Color: "cyan", Underline: true}
		t.InlineCode = types.StyleIntent{Color: "green"}
		t.BlockCode = types.StyleIntent{Color: "green"}
		return t
	}()

	Solarized = types.Theme{
		Heading:     types.StyleIntent{Color: "yellow", Bold: true},
		Strong:      types.StyleIntent{Bold: true},
		Emph:        types.StyleIntent{Italic: true},
		InlineCode:  types.StyleIntent{Color: "cyan"},
		BlockCode:   types.StyleIntent{Color: "#2aa198"},
		Link:        types.StyleIntent{Color: "blue", Underline: true},
		Quote:       types.StyleIntent{Color: "white", Dim: true},
		HR:          types.StyleIntent{Color: "white", Dim: true},
		ListMarker:  types.StyleIntent{Color: "cyan"},
		TableHeader: types.StyleIntent{Color: "yellow", Bold: true},
	}

	Monochrome = types.Theme{
		Heading:     types.StyleIntent{Bold: true},
		Strong:      types.StyleIntent{Bold: true},
		Emph:        types.StyleIntent{Italic: true},
		InlineCode:  types.StyleIntent{Dim: true},
		BlockCode:   types.StyleIntent{Dim: true},
		Link:        types.StyleIntent{Underline: true},
		Quote:       types.StyleIntent{Dim: true},
		HR:          types.StyleIntent{Dim: true},
		ListMarker:  types.StyleIntent{Dim: true},
		TableHeader: types.StyleIntent{Bold: true},
	}

	Contrast = types.Theme{
		Heading:     types.StyleIntent{Color: "magenta", Bold: true},
		Strong:      types.StyleIntent{Color: "white", Bold: true},
		Emph:        types.StyleIntent{Color: "white", Italic: true},
		InlineCode:  types.StyleIntent{Color: "cyan", Bold: true},
		BlockCode:   types.StyleIntent{Color: "green", Bold: true},
		Link:        types.StyleIntent{Color: "blue", Underline: true},
		Quote:       types.StyleIntent{Color: "white", Dim: true},
		HR:          types.StyleIntent{Color: "white", Dim: true},
		ListMarker:  types.StyleIntent{Color: "yellow", Bold: true},
		TableHeader: types.StyleIntent{Color: "yellow", Bold: true},
		TableCell:   types.StyleIntent{Color: "white"},
	}
)

// ByName returns the built-in theme for the given name, falling back to Default
// if the name is not recognised.
func ByName(name types.ThemeName) types.Theme {
	switch name {
	case types.ThemeDim:
		return Dim
	case types.ThemeBright:
		return Bright
	case types.ThemeSolarized:
		return Solarized
	case types.ThemeMonochrome:
		return Monochrome
	case types.ThemeContrast:
		return Contrast
	default:
		return Default
	}
}

// Resolve returns the effective theme from a RenderOptions: a custom Theme
// object if set, otherwise the named theme (defaulting to Default).
// It also applies the inlineCode/blockCode fallback to the Code field.
func Resolve(opts types.RenderOptions) types.Theme {
	var base types.Theme
	if opts.Theme != nil {
		base = *opts.Theme
	} else {
		base = ByName(opts.ThemeName)
	}
	return applyCodeFallback(base)
}

// applyCodeFallback ensures InlineCode and BlockCode fall back to Code when
// they are zero-valued, mirroring the mergedTheme logic in src/render.ts.
func applyCodeFallback(t types.Theme) types.Theme {
	if (t.InlineCode == types.StyleIntent{}) {
		t.InlineCode = t.Code
	}
	if (t.BlockCode == types.StyleIntent{}) {
		t.BlockCode = t.Code
	}
	return t
}

// Merge returns a new Theme with non-zero fields from override applied on top
// of base. A field is considered "set" if it is non-zero.
func Merge(base, override types.Theme) types.Theme {
	result := base
	if (override.Heading != types.StyleIntent{}) {
		result.Heading = override.Heading
	}
	if (override.Strong != types.StyleIntent{}) {
		result.Strong = override.Strong
	}
	if (override.Emph != types.StyleIntent{}) {
		result.Emph = override.Emph
	}
	if (override.InlineCode != types.StyleIntent{}) {
		result.InlineCode = override.InlineCode
	}
	if (override.BlockCode != types.StyleIntent{}) {
		result.BlockCode = override.BlockCode
	}
	if (override.Code != types.StyleIntent{}) {
		result.Code = override.Code
	}
	if (override.Link != types.StyleIntent{}) {
		result.Link = override.Link
	}
	if (override.Quote != types.StyleIntent{}) {
		result.Quote = override.Quote
	}
	if (override.HR != types.StyleIntent{}) {
		result.HR = override.HR
	}
	if (override.ListMarker != types.StyleIntent{}) {
		result.ListMarker = override.ListMarker
	}
	if (override.TableHeader != types.StyleIntent{}) {
		result.TableHeader = override.TableHeader
	}
	if (override.TableCell != types.StyleIntent{}) {
		result.TableCell = override.TableCell
	}
	return result
}

// Styler applies a StyleIntent to a string, emitting ANSI escape codes when
// color output is enabled. When color is false it is a no-op.
type Styler struct {
	color   bool
	profile termenv.Profile
}

// NewStyler creates a Styler. Pass color=false to disable all ANSI output.
func NewStyler(color bool) *Styler {
	profile := termenv.TrueColor
	if !color {
		profile = termenv.Ascii
	}
	return &Styler{color: color, profile: profile}
}

// Apply styles text according to the given StyleIntent and returns the result.
func (s *Styler) Apply(text string, intent types.StyleIntent) string {
	if !s.color {
		return text
	}

	style := s.profile.String(text)

	if intent.Color != "" {
		style = style.Foreground(s.resolveColor(intent.Color))
	}
	if intent.BgColor != "" {
		style = style.Background(s.resolveColor(intent.BgColor))
	}
	if intent.Bold {
		style = style.Bold()
	}
	if intent.Italic {
		style = style.Italic()
	}
	if intent.Underline {
		style = style.Underline()
	}
	if intent.Dim {
		style = style.Faint()
	}
	if intent.Strike {
		style = style.CrossOut()
	}

	return style.String()
}

// resolveColor maps a color name, hex value, or ANSI index string to a
// termenv.Color. Named colors map to ANSI 16-color values; hex strings (#rrggbb)
// use TrueColor; bare integers use ANSI 256-color.
func (s *Styler) resolveColor(name string) termenv.Color {
	switch strings.ToLower(name) {
	case "black":
		return s.profile.Color("0")
	case "red":
		return s.profile.Color("1")
	case "green":
		return s.profile.Color("2")
	case "yellow":
		return s.profile.Color("3")
	case "blue":
		return s.profile.Color("4")
	case "magenta":
		return s.profile.Color("5")
	case "cyan":
		return s.profile.Color("6")
	case "white":
		return s.profile.Color("7")
	case "gray", "grey":
		return s.profile.Color("8")
	}
	// Hex color (#rrggbb) or ANSI 256 index — pass directly to termenv.
	if strings.HasPrefix(name, "#") || isDigits(name) {
		return s.profile.Color(name)
	}
	// Unknown: return a no-op color.
	return s.profile.Color(name)
}

func isDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return len(s) > 0
}
