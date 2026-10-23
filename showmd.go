// Package showmd renders Markdown for display in a terminal.
//
// The renderer never probes the real terminal or environment: callers describe
// the output device with a Terminal value. The zero Terminal renders plain text
// wrapped at 80 columns.
//
//	out := showmd.Render(md, showmd.Terminal{Width: 100, Color: showmd.ColorTrue}, showmd.Options{})
package showmd

import (
	"github.com/funcan/showmd/internal/render"
	"github.com/funcan/showmd/internal/types"
)

type (
	// Terminal describes the capabilities of the output device.
	Terminal = types.Terminal
	// ColorProfile is the colour capability of the output device.
	ColorProfile = types.ColorProfile
	// Options controls rendering style. All fields are optional.
	Options = types.RenderOptions
	// Theme maps semantic roles to styles.
	Theme = types.Theme
	// ThemeName is the name of a built-in theme.
	ThemeName = types.ThemeName
	// StyleIntent describes how a piece of text should be styled.
	StyleIntent = types.StyleIntent
	// TableBorder controls the border style used when rendering tables.
	TableBorder = types.TableBorder
	// Highlighter is a syntax-highlighting hook for fenced code blocks.
	Highlighter = types.Highlighter
)

const (
	ColorNone = types.ColorNone
	Color16   = types.Color16
	Color256  = types.Color256
	ColorTrue = types.ColorTrue

	ThemeDefault    = types.ThemeDefault
	ThemeDim        = types.ThemeDim
	ThemeBright     = types.ThemeBright
	ThemeSolarized  = types.ThemeSolarized
	ThemeMonochrome = types.ThemeMonochrome
	ThemeContrast   = types.ThemeContrast

	TableBorderUnicode = types.TableBorderUnicode
	TableBorderASCII   = types.TableBorderASCII
	TableBorderNone    = types.TableBorderNone
)

// Render renders markdown for display on a terminal with the given
// capabilities.
func Render(markdown string, term Terminal, opts Options) string {
	return render.Render(markdown, term, opts)
}
