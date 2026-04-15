// Package types defines the public data types shared across all showmd packages.
// It is a direct port of src/types.ts from Markdansi.
package types

// ThemeName is the name of a built-in theme.
type ThemeName string

const (
	ThemeDefault    ThemeName = "default"
	ThemeDim        ThemeName = "dim"
	ThemeBright     ThemeName = "bright"
	ThemeSolarized  ThemeName = "solarized"
	ThemeMonochrome ThemeName = "monochrome"
	ThemeContrast   ThemeName = "contrast"
)

// Highlighter is a hook called for fenced code blocks. It receives the raw code
// and an optional language name, and returns ANSI-coloured text. It must not
// add or remove newlines — showmd owns all indentation and box-drawing.
type Highlighter func(code, lang string) string

// TableBorder controls the border style used when rendering tables.
type TableBorder string

const (
	TableBorderUnicode TableBorder = "unicode"
	TableBorderASCII   TableBorder = "ascii"
	TableBorderNone    TableBorder = "none"
)

// StyleIntent describes how a piece of text should be styled. The zero value
// means "no styling". Color is a named terminal color ("red", "cyan", …),
// a hex value ("#rrggbb"), or an ANSI 256-color index ("42").
type StyleIntent struct {
	Color     string // named color, "#rrggbb", ANSI index, or ""
	BgColor   string
	Bold      bool
	Italic    bool
	Underline bool
	Dim       bool
	Strike    bool
}

// Theme maps semantic roles to StyleIntents. All fields are optional; the zero
// value of a StyleIntent means "inherit / no style".
//
// InlineCode and BlockCode are used for their respective contexts when set;
// otherwise Code is used as a fallback for both.
type Theme struct {
	Heading     StyleIntent
	Strong      StyleIntent
	Emph        StyleIntent
	InlineCode  StyleIntent
	BlockCode   StyleIntent
	Code        StyleIntent // fallback for InlineCode and BlockCode
	Link        StyleIntent
	Quote       StyleIntent
	HR          StyleIntent
	ListMarker  StyleIntent
	TableHeader StyleIntent
	TableCell   StyleIntent
}

// RenderOptions controls how Markdown is rendered to ANSI. All fields are
// optional; unset pointer fields mean "use the default".
//
// This is a direct port of the RenderOptions interface from src/types.ts.
type RenderOptions struct {
	// Wrap enables hard word-wrapping. Default true.
	Wrap *bool
	// Width is the wrap column. Used only when Wrap is true.
	// Defaults to the TTY width or 80.
	Width *int
	// Color enables ANSI escape output. Defaults to true when stdout is a TTY.
	// When false, all ANSI and OSC sequences are stripped from output.
	Color *bool
	// Hyperlinks enables OSC-8 hyperlinks. Defaults to auto-detect.
	// Forced off when Color is false.
	Hyperlinks *bool
	// Theme is either a ThemeName string or a custom Theme object.
	// Defaults to ThemeDefault.
	ThemeName ThemeName
	Theme     *Theme

	// ListIndent is spaces per nesting level for lists. Default 2.
	ListIndent *int
	// QuotePrefix is the prefix prepended to every blockquote line. Default "│ ".
	QuotePrefix *string

	// TableBorder controls table border style. Default TableBorderUnicode.
	TableBorder TableBorder
	// TablePadding is spaces around cell content. Default 1.
	TablePadding *int
	// TableDense reduces separator rows. Default false.
	TableDense bool
	// TableTruncate truncates cell content to fit column width. Default true.
	TableTruncate *bool
	// TableEllipsis is the truncation marker. Default "…".
	TableEllipsis *string

	// CodeBox draws a box around fenced code blocks. Default true.
	CodeBox *bool
	// CodeGutter shows 1-based line numbers in a gutter. Default false.
	CodeGutter bool
	// CodeWrap wraps long code lines to the render width. Default true.
	CodeWrap *bool

	// Highlighter is an optional syntax-highlighting hook for code blocks.
	Highlighter Highlighter
}

// Bool is a helper to get a *bool option value with a fallback default.
func Bool(p *bool, def bool) bool {
	if p == nil {
		return def
	}
	return *p
}

// Int is a helper to get a *int option value with a fallback default.
func Int(p *int, def int) int {
	if p == nil {
		return def
	}
	return *p
}

// String is a helper to get a *string option value with a fallback default.
func String(p *string, def string) string {
	if p == nil {
		return def
	}
	return *p
}
