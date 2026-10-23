// Package theme defines the built-in named themes and the Styler type, which
// applies a StyleIntent to a string using lipgloss.
//
// This is a direct port of src/theme.ts from Markdansi.
package theme

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/funcan/showmd/internal/types"
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

// Styler applies a StyleIntent to a string using lipgloss. It always emits
// full-fidelity (truecolor) escape sequences; the renderer downsamples the
// final output to the terminal's colour profile. When disabled it is a no-op.
type Styler struct {
	enabled bool
}

// NewStyler creates a Styler. Pass enabled=false to disable all escape output.
func NewStyler(enabled bool) *Styler {
	return &Styler{enabled: enabled}
}

// Apply styles text according to the given StyleIntent and returns the result.
func (s *Styler) Apply(text string, intent types.StyleIntent) string {
	if !s.enabled {
		return text
	}
	return style(intent).Render(text)
}

// Link styles text and makes it an OSC-8 hyperlink to url.
func (s *Styler) Link(text, url string, intent types.StyleIntent) string {
	if !s.enabled {
		return text
	}
	return style(intent).Hyperlink(url).Render(text)
}

func style(intent types.StyleIntent) lipgloss.Style {
	st := lipgloss.NewStyle().
		TabWidth(lipgloss.NoTabConversion).
		Bold(intent.Bold).
		Italic(intent.Italic).
		Underline(intent.Underline).
		Faint(intent.Dim).
		Strikethrough(intent.Strike)
	if intent.Color != "" {
		st = st.Foreground(resolveColor(intent.Color))
	}
	if intent.BgColor != "" {
		st = st.Background(resolveColor(intent.BgColor))
	}
	return st
}

// resolveColor maps a color name, hex value ("#rrggbb"), or ANSI 256 index to a
// color. Named colors map to the basic 16 ANSI colors.
func resolveColor(name string) color.Color {
	switch strings.ToLower(name) {
	case "black":
		return lipgloss.Black
	case "red":
		return lipgloss.Red
	case "green":
		return lipgloss.Green
	case "yellow":
		return lipgloss.Yellow
	case "blue":
		return lipgloss.Blue
	case "magenta":
		return lipgloss.Magenta
	case "cyan":
		return lipgloss.Cyan
	case "white":
		return lipgloss.White
	case "gray", "grey":
		return lipgloss.BrightBlack
	}
	return lipgloss.Color(name)
}
