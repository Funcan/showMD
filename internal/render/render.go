// Package render provides the Markdown → ANSI renderer.
// render.go contains the public API and top-level rendering logic.
package render

import (
	"regexp"
	"strings"

	"github.com/funcan/showmd/internal/theme"
	"github.com/funcan/showmd/internal/types"
	"github.com/funcan/showmd/internal/wrap"
	"github.com/yuin/goldmark/ast"
)

// resolvedOptions holds all render options with defaults applied.
type resolvedOptions struct {
	wrap             bool
	width            int // 0 = no fixed width (only meaningful when wrap=false)
	color            types.ColorProfile
	hyperlinks       bool
	theme            types.Theme
	listIndent       int
	quotePrefix      string
	tableBorder      types.TableBorder
	tablePadding     int
	tableDense       bool
	tableTruncate    bool
	tableEllipsis    string
	codeBox          bool
	codeGutter       bool
	codeWrap         bool
	squashParagraphs bool
	highlighter      types.Highlighter
}

func resolveOptions(term types.Terminal, opts types.RenderOptions) resolvedOptions {
	doWrap := types.Bool(opts.Wrap, true)

	width := term.Width
	if width == 0 && doWrap {
		width = 80
	}

	th := theme.Resolve(opts)

	tableBorder := opts.TableBorder
	if tableBorder == "" {
		tableBorder = types.TableBorderUnicode
	}

	return resolvedOptions{
		wrap:             doWrap,
		width:            width,
		color:            term.Color,
		hyperlinks:       term.Hyperlinks && term.Color != types.ColorNone,
		theme:            th,
		listIndent:       types.Int(opts.ListIndent, 2),
		quotePrefix:      types.String(opts.QuotePrefix, "> "),
		tableBorder:      tableBorder,
		tablePadding:     types.Int(opts.TablePadding, 1),
		tableDense:       opts.TableDense,
		tableTruncate:    types.Bool(opts.TableTruncate, true),
		tableEllipsis:    types.String(opts.TableEllipsis, "…"),
		codeBox:          types.Bool(opts.CodeBox, true),
		codeGutter:       opts.CodeGutter,
		codeWrap:         types.Bool(opts.CodeWrap, true),
		squashParagraphs: opts.SquashParagraphs,
		highlighter:      opts.Highlighter,
	}
}

// renderContext is passed through all rendering functions.
type renderContext struct {
	opts   resolvedOptions
	styler *theme.Styler
	src    []byte
}

// effectiveWidth returns the configured terminal width, defaulting to 80.
func (ctx *renderContext) effectiveWidth() int {
	if ctx.opts.width == 0 {
		return 80
	}
	return ctx.opts.width
}

// Render renders markdown to a string for display on a terminal with the
// given capabilities. It never inspects the real terminal or environment.
func Render(markdown string, term types.Terminal, opts types.RenderOptions) string {
	resolved := resolveOptions(term, opts)
	styler := theme.NewStyler(resolved.color)
	doc := Parse(dedent(markdown))
	ctx := &renderContext{opts: resolved, styler: styler, src: doc.Source}
	output := renderBlocks(doc.Blocks, ctx)
	if resolved.color == types.ColorNone {
		output = wrap.StripANSI(output)
	}
	return output
}

// dedent removes the minimum common leading whitespace from all non-empty lines.
func dedent(s string) string {
	lines := strings.Split(s, "\n")
	minIndent := -1
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			continue
		}
		indent := 0
		for _, ch := range l {
			if ch != ' ' && ch != '\t' {
				break
			}
			indent++
		}
		if minIndent < 0 || indent < minIndent {
			minIndent = indent
		}
	}
	if minIndent <= 0 {
		return s
	}
	out := make([]string, len(lines))
	for i, l := range lines {
		if len(l) < minIndent {
			out[i] = l
		} else {
			out[i] = l[minIndent:]
		}
	}
	return strings.Join(out, "\n")
}

const (
	hrWidth = 40
	maxCol  = 40
)

var reLangLabel = regexp.MustCompile(`^\[([^\]]+)\]$`)

// renderBlocks renders the normalized top-level block slice.
func renderBlocks(blocks []*Block, ctx *renderContext) string {
	var sb strings.Builder
	for i := 0; i < len(blocks); i++ {
		b := blocks[i]
		// Heuristic: paragraph containing only "[lang]" text followed by a
		// no-lang code block → apply the lang and render the code block.
		if b.Kind == ast.KindParagraph && b.Content == "" {
			text := strings.TrimSpace(extractNodeText(b.Node, ctx.src))
			if m := reLangLabel.FindStringSubmatch(text); m != nil && i+1 < len(blocks) {
				next := blocks[i+1]
				if isCode(next) && next.Lang == "" {
					next.Lang = m[1]
					i++
					sb.WriteString(renderBlock(next, ctx, 0))
					continue
				}
			}
		}
		sb.WriteString(renderBlock(b, ctx, 0))
		// Add a blank line between consecutive paragraphs unless squashing.
		if !ctx.opts.squashParagraphs &&
			b.Kind == ast.KindParagraph &&
			i+1 < len(blocks) &&
			blocks[i+1].Kind == ast.KindParagraph {
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}
