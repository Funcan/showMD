package render

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/funcan/showmd/internal/hyperlink"
	"github.com/funcan/showmd/internal/types"
	"github.com/yuin/goldmark/ast"
	east "github.com/yuin/goldmark/extension/ast"
)

// hardBreak is the sentinel rune used for hard line breaks within inline text.
// Mirrors the TS HARD_BREAK = "\u000B" (vertical tab).
const hardBreak = '\v'

// renderInlineContent renders all inline children of n into a string.
func renderInlineContent(n ast.Node, ctx *renderContext) string {
	var sb strings.Builder
	renderInlineChildren(n, ctx, &sb)
	return sb.String()
}

// renderInlineChildren iterates the direct children of n and renders inline nodes.
func renderInlineChildren(n ast.Node, ctx *renderContext, sb *strings.Builder) {
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		renderInlineSingle(child, ctx, sb)
	}
}

// renderInlineSingle renders a single inline node.
func renderInlineSingle(child ast.Node, ctx *renderContext, sb *strings.Builder) {
	switch child.Kind() {
	case ast.KindText:
		t := child.(*ast.Text)
		sb.Write(t.Segment.Value(ctx.src))
		if t.HardLineBreak() {
			sb.WriteRune(hardBreak)
		} else if t.SoftLineBreak() {
			sb.WriteByte('\n')
		}

	case ast.KindEmphasis:
		em := child.(*ast.Emphasis)
		inner := innerInlineText(em, ctx)
		if em.Level == 1 {
			sb.WriteString(ctx.styler.Apply(inner, ctx.opts.theme.Emph))
		} else {
			sb.WriteString(ctx.styler.Apply(inner, ctx.opts.theme.Strong))
		}

	case ast.KindCodeSpan:
		val := codeSpanText(child, ctx.src)
		codeStyle := ctx.opts.theme.InlineCode
		if codeStyle == (types.StyleIntent{}) {
			codeStyle = ctx.opts.theme.Code
		}
		sb.WriteString(ctx.styler.Apply(val, codeStyle))

	case ast.KindLink:
		sb.WriteString(renderLink(child.(*ast.Link), ctx))

	case ast.KindAutoLink:
		sb.WriteString(renderAutoLink(child.(*ast.AutoLink), ctx))

	case ast.KindImage:
		// Render alt text only.
		renderInlineChildren(child, ctx, sb)

	case ast.KindRawHTML:
		// Skip raw HTML in inline context.

	case east.KindStrikethrough:
		inner := innerInlineText(child, ctx)
		sb.WriteString(ctx.styler.Apply(inner, types.StyleIntent{Strike: true}))

	case east.KindTaskCheckBox:
		// Task checkbox is displayed as the list item marker; omit here.

	case ast.KindString:
		s := child.(*ast.String)
		sb.Write(s.Value)

	default:
		// Recurse into unknown inline container nodes to capture their text.
		renderInlineChildren(child, ctx, sb)
	}
}

// innerInlineText renders inline children of n into a string.
func innerInlineText(n ast.Node, ctx *renderContext) string {
	var sb strings.Builder
	renderInlineChildren(n, ctx, &sb)
	return sb.String()
}

// codeSpanText extracts the raw text content of an ast.CodeSpan node.
func codeSpanText(n ast.Node, src []byte) string {
	var sb strings.Builder
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if c.Kind() == ast.KindText {
			t := c.(*ast.Text)
			sb.Write(t.Segment.Value(src))
		}
	}
	return sb.String()
}

// renderLink renders a Markdown link node.
func renderLink(n *ast.Link, ctx *renderContext) string {
	label := innerInlineText(n, ctx)
	url := string(n.Destination)
	if label == "" {
		label = url
	}
	if url == "" {
		return label
	}
	// Mailto autolinks: treat as plain text.
	if strings.HasPrefix(url, "mailto:") {
		return label
	}
	if ctx.opts.hyperlinks {
		return hyperlink.OSC8(url, label)
	}
	if label != url {
		return ctx.styler.Apply(label, ctx.opts.theme.Link) +
			ctx.styler.Apply(" ("+url+")", types.StyleIntent{Dim: true})
	}
	return ctx.styler.Apply(label, ctx.opts.theme.Link)
}

// renderAutoLink renders an autolink (bare URL or email).
func renderAutoLink(n *ast.AutoLink, ctx *renderContext) string {
	url := string(n.URL(ctx.src))
	label := url
	if n.AutoLinkType == ast.AutoLinkEmail {
		return label
	}
	if strings.HasPrefix(url, "mailto:") {
		return label
	}
	if ctx.opts.hyperlinks && url != "" {
		return hyperlink.OSC8(url, label)
	}
	return ctx.styler.Apply(label, ctx.opts.theme.Link)
}

// normalizeParagraphInlineText converts raw inline text (with \n for soft
// breaks and hardBreak sentinel for hard breaks) into final paragraph text.
//
// Soft breaks between non-empty, non-definition lines collapse to a space;
// hard breaks become literal newlines.
func normalizeParagraphInlineText(text string) string {
	if !strings.ContainsRune(text, '\n') && !strings.ContainsRune(text, hardBreak) {
		return text
	}

	type seg struct {
		text     string
		isHard   bool // true = hard break follows this segment
		hasBreak bool // true = any break follows (used to identify last segment)
	}

	var segs []seg
	var current strings.Builder
	for i := 0; i < len(text); {
		r, size := utf8.DecodeRuneInString(text[i:])
		i += size
		switch r {
		case '\n':
			segs = append(segs, seg{text: current.String(), isHard: false, hasBreak: true})
			current.Reset()
		case hardBreak:
			segs = append(segs, seg{text: current.String(), isHard: true, hasBreak: true})
			current.Reset()
		default:
			current.WriteRune(r)
		}
	}
	segs = append(segs, seg{text: current.String()})

	var sb strings.Builder
	sb.WriteString(segs[0].text)

	for i := 0; i < len(segs)-1; i++ {
		s := segs[i]
		right := segs[i+1].text

		if s.isHard {
			sb.WriteByte('\n')
			sb.WriteString(right)
			continue
		}

		// Soft break: collapse to space unless either side is empty or looks like a definition.
		left := s.text
		leftTrim := strings.TrimLeft(left, " \t")
		rightTrim := strings.TrimLeft(right, " \t")
		keepNewline := left == "" || right == "" || isRefDef(leftTrim) || isRefDef(rightTrim)
		if keepNewline {
			sb.WriteByte('\n')
			sb.WriteString(rightTrim)
		} else {
			sb.WriteByte(' ')
			sb.WriteString(rightTrim)
		}
	}
	return sb.String()
}

// reRefDef matches a reference-style link definition: [label]: url
var reRefDef = regexp.MustCompile(`^\[[^\]]+\]:\s+\S`)

// isRefDef reports whether s looks like a reference-style link definition.
func isRefDef(s string) bool {
	return reRefDef.MatchString(s)
}
