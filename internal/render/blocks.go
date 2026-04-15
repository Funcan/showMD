package render

import (
	"fmt"
	"strings"

	"github.com/funcan/showmd/internal/types"
	"github.com/funcan/showmd/internal/wrap"
	"github.com/yuin/goldmark/ast"
	east "github.com/yuin/goldmark/extension/ast"
)

// renderBlock dispatches a top-level normalized Block to its renderer.
func renderBlock(b *Block, ctx *renderContext) string {
	switch {
	case b.Kind == ast.KindParagraph || b.Kind == ast.KindTextBlock:
		if b.Content != "" {
			return renderParagraphText(b.Content, ctx, 0)
		}
		return renderParagraph(b.Node, ctx, 0)
	case b.Kind == ast.KindHeading:
		return renderHeading(b.Node, ctx)
	case b.Kind == ast.KindThematicBreak:
		return renderHr(ctx)
	case b.Kind == ast.KindBlockquote:
		return renderBlockquote(b.Node, ctx, 0)
	case b.Kind == ast.KindList:
		return renderList(b.Node, ctx, 0)
	case isCode(b):
		return renderCodeBlock(b, ctx)
	case b.Kind == east.KindTable:
		return renderTable(b.Node, ctx)
	case b.Kind == ast.KindHTMLBlock:
		return ""
	default:
		return ""
	}
}

// renderASTNode renders a single goldmark AST node as a block.
// Used for nested content (inside blockquotes, list items, etc.).
func renderASTNode(n ast.Node, ctx *renderContext, indentLevel int, tight bool) string {
	switch n.Kind() {
	case ast.KindParagraph, ast.KindTextBlock:
		return renderParagraph(n, ctx, indentLevel)
	case ast.KindHeading:
		return renderHeading(n, ctx)
	case ast.KindThematicBreak:
		return renderHr(ctx)
	case ast.KindBlockquote:
		return renderBlockquote(n, ctx, indentLevel)
	case ast.KindList:
		return renderList(n, ctx, indentLevel)
	case ast.KindFencedCodeBlock, ast.KindCodeBlock:
		lang := CodeBlockLang(n, ctx.src)
		content := CodeBlockContent(n, ctx.src)
		if lang == "" && looksLikeDiff(content) {
			lang = "diff"
		}
		synth := &Block{Node: n, Kind: n.Kind(), Lang: lang, Content: content}
		return renderCodeBlock(synth, ctx)
	case east.KindTable:
		return renderTable(n, ctx)
	case ast.KindHTMLBlock:
		return ""
	default:
		return ""
	}
}

// renderASTChildren renders all block children of n, applying the [lang] label
// heuristic (paragraph with "[lang]" text before a no-lang code block).
func renderASTChildren(n ast.Node, ctx *renderContext, indentLevel int, tight bool) string {
	// Collect children for lookahead.
	var children []ast.Node
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		children = append(children, child)
	}

	var sb strings.Builder
	for i := 0; i < len(children); i++ {
		child := children[i]
		if child.Kind() == ast.KindParagraph {
			text := strings.TrimSpace(extractNodeText(child, ctx.src))
			if m := reLangLabel.FindStringSubmatch(text); m != nil && i+1 < len(children) {
				next := children[i+1]
				if (next.Kind() == ast.KindFencedCodeBlock || next.Kind() == ast.KindCodeBlock) &&
					CodeBlockLang(next, ctx.src) == "" {
					lang := m[1]
					content := CodeBlockContent(next, ctx.src)
					if lang == "diff" && looksLikeDiff(content) {
						lang = "diff"
					}
					synth := &Block{Node: next, Kind: next.Kind(), Lang: lang, Content: content}
					sb.WriteString(renderCodeBlock(synth, ctx))
					i++ // skip the code block (we skipped the label paragraph by not writing it)
					continue
				}
			}
		}
		sb.WriteString(renderASTNode(child, ctx, indentLevel, tight))
	}
	return sb.String()
}

// renderHeading renders a heading node.
func renderHeading(n ast.Node, ctx *renderContext) string {
	text := renderInlineContent(n, ctx)
	styled := ctx.styler.Apply(text, ctx.opts.theme.Heading)
	return "\n" + styled + "\n"
}

// renderHr renders a thematic break (horizontal rule).
func renderHr(ctx *renderContext) string {
	width := hrWidth
	if ctx.opts.wrap && ctx.opts.width > 0 && ctx.opts.width < hrWidth {
		width = ctx.opts.width
	}
	line := strings.Repeat("—", width)
	return ctx.styler.Apply(line, ctx.opts.theme.HR) + "\n"
}

// renderParagraph renders a paragraph node from its goldmark AST node.
func renderParagraph(n ast.Node, ctx *renderContext, indentLevel int) string {
	raw := renderInlineContent(n, ctx)
	return renderParagraphText(raw, ctx, indentLevel)
}

// renderParagraphText renders a paragraph from pre-rendered inline text.
func renderParagraphText(raw string, ctx *renderContext, indentLevel int) string {
	text := normalizeParagraphInlineText(raw)
	prefix := strings.Repeat(" ", ctx.opts.listIndent*indentLevel)
	rawLines := strings.Split(text, "\n")

	// Group definition-like lines together (mirrors TS renderParagraph logic).
	var normalized []string
	inDefinitions := false
	for _, line := range rawLines {
		if reRefDef.MatchString(line) && len(normalized) > 0 && normalized[len(normalized)-1] != "" {
			normalized = append(normalized, "")
		}
		if reRefDef.MatchString(line) {
			inDefinitions = true
			normalized = append(normalized, line)
			continue
		}
		if inDefinitions && strings.TrimSpace(line) == "" {
			continue
		}
		inDefinitions = false
		normalized = append(normalized, line)
	}

	width := ctx.opts.width
	if width == 0 {
		width = 80
	}

	var sb strings.Builder
	for _, l := range normalized {
		for _, part := range wrap.WrapWithPrefix(l, width, ctx.opts.wrap, prefix) {
			sb.WriteString(part)
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}

// renderBlockquote renders a blockquote node.
func renderBlockquote(n ast.Node, ctx *renderContext, indentLevel int) string {
	inner := renderASTChildren(n, ctx, indentLevel, false)
	prefix := ctx.styler.Apply(ctx.opts.quotePrefix, ctx.opts.theme.Quote)
	text := strings.TrimRight(inner, "\n")
	width := ctx.opts.width
	if width == 0 {
		width = 80
	}
	wrapped := wrap.WrapWithPrefix(text, width, ctx.opts.wrap, prefix)
	var sb strings.Builder
	for _, l := range wrapped {
		sb.WriteString(l)
		sb.WriteByte('\n')
	}
	return sb.String()
}

// renderList renders a list node.
func renderList(n ast.Node, ctx *renderContext, indentLevel int) string {
	list := n.(*ast.List)
	tight := list.IsTight
	ordered := list.IsOrdered()
	start := list.Start

	var sb strings.Builder
	idx := 0
	for item := list.FirstChild(); item != nil; item = item.NextSibling() {
		sb.WriteString(renderListItem(item, ctx, indentLevel, tight, ordered, start, idx))
		idx++
	}
	return sb.String()
}

// renderListItem renders a single list item.
func renderListItem(n ast.Node, ctx *renderContext, indentLevel int, tight bool, ordered bool, start int, idx int) string {
	isTask, checked := isTaskItem(n)

	var marker string
	if isTask {
		if checked {
			marker = "[x]"
		} else {
			marker = "[ ]"
		}
	} else if ordered {
		marker = fmt.Sprintf("%d.", start+idx)
	} else {
		marker = "-"
	}
	markerStyled := ctx.styler.Apply(marker, ctx.opts.theme.ListMarker)

	content := renderASTChildren(n, ctx, indentLevel+1, tight)
	content = strings.TrimRight(content, "\n")
	contentLines := strings.Split(content, "\n")

	// Drop leading blank lines.
	for len(contentLines) > 0 && strings.TrimSpace(contentLines[0]) == "" {
		contentLines = contentLines[1:]
	}

	indent := strings.Repeat(" ", ctx.opts.listIndent*indentLevel)
	continuationIndent := indent + strings.Repeat(" ", ctx.opts.listIndent)
	firstBullet := indent + markerStyled + " "

	var lines []string
	for i, line := range contentLines {
		clean := strings.TrimLeft(line, " \t")
		if i == 0 {
			lines = append(lines, firstBullet+clean)
		} else {
			lines = append(lines, continuationIndent+clean)
		}
	}
	if !tight {
		lines = append(lines, "")
	}

	var sb strings.Builder
	for _, l := range lines {
		sb.WriteString(l)
		sb.WriteByte('\n')
	}
	return sb.String()
}

// isTaskItem reports whether n is a task list item, and whether it is checked.
func isTaskItem(n ast.Node) (isTask bool, checked bool) {
	firstBlock := n.FirstChild()
	if firstBlock == nil {
		return false, false
	}
	firstInline := firstBlock.FirstChild()
	if firstInline == nil {
		return false, false
	}
	if firstInline.Kind() == east.KindTaskCheckBox {
		cb := firstInline.(*east.TaskCheckBox)
		return true, cb.IsChecked
	}
	return false, false
}

// renderCodeBlock renders a fenced or indented code block.
func renderCodeBlock(b *Block, ctx *renderContext) string {
	codeStyle := ctx.opts.theme.BlockCode
	if codeStyle == (types.StyleIntent{}) {
		codeStyle = ctx.opts.theme.Code
	}

	content := b.Content
	if content == "" && b.Node != nil {
		content = CodeBlockContent(b.Node, ctx.src)
	}

	lines := strings.Split(content, "\n")
	isDiff := b.Lang == "diff"
	gutterWidth := 0
	if ctx.opts.codeGutter {
		gutterWidth = len(fmt.Sprintf("%d", len(lines))) + 2
	}
	shouldWrap := !isDiff && ctx.opts.codeWrap
	useBox := ctx.opts.codeBox && len(lines) > 1
	boxPadding := 0
	if useBox {
		boxPadding = 4
	}

	wrapLimit := 0
	if shouldWrap && ctx.opts.wrap && ctx.opts.width > 0 {
		limit := ctx.opts.width - boxPadding - gutterWidth
		if limit < 1 {
			limit = 1
		}
		wrapLimit = limit
	}

	var contentLines []string
	for lineIdx, line := range lines {
		var segments []string
		if wrapLimit > 0 {
			segments = wrapCodeLine(line, wrapLimit)
		} else {
			segments = []string{line}
		}
		for segIdx, seg := range segments {
			var highlighted string
			if ctx.opts.highlighter != nil {
				highlighted = ctx.opts.highlighter(seg, b.Lang)
			} else {
				highlighted = ctx.styler.Apply(seg, codeStyle)
			}
			if ctx.opts.codeGutter {
				var num string
				if segIdx == 0 {
					num = fmt.Sprintf("%*d", gutterWidth-2, lineIdx+1)
				} else {
					num = strings.Repeat(" ", gutterWidth-1)
				}
				highlighted = ctx.styler.Apply(num, types.StyleIntent{Dim: true}) + " " + highlighted
			}
			contentLines = append(contentLines, highlighted)
		}
	}

	if !useBox {
		return strings.Join(contentLines, "\n") + "\n\n"
	}

	// Boxed rendering.
	maxLine := 0
	for _, l := range contentLines {
		if w := wrap.VisibleWidth(l); w > maxLine {
			maxLine = w
		}
	}

	labelRaw := ""
	if b.Lang != "" {
		labelRaw = "[" + b.Lang + "]"
	}
	labelStyled := ""
	if labelRaw != "" {
		labelStyled = ctx.styler.Apply(labelRaw, types.StyleIntent{Dim: true})
	}

	wrapTarget := maxLine
	if ctx.opts.codeWrap && ctx.opts.width > 0 {
		t := ctx.opts.width - 4
		if t < 1 {
			t = 1
		}
		if maxLine < t {
			wrapTarget = maxLine
		} else {
			wrapTarget = t
		}
	}

	minInner := 0
	if b.Lang != "" {
		minInner = len(b.Lang) + 2
	}

	innerWidth := wrapTarget
	if !ctx.opts.codeWrap {
		innerWidth = maxLine
	}
	if minInner > innerWidth {
		innerWidth = minInner
	}
	if len(labelRaw) > innerWidth {
		innerWidth = len(labelRaw)
	}

	var topRaw string
	if len(labelRaw) > 0 {
		topPadding := innerWidth - len(labelRaw) + 1
		if topPadding < 0 {
			topPadding = 0
		}
		topRaw = "┌ " + labelStyled + strings.Repeat("─", topPadding) + "┐"
	} else {
		topRaw = "┌ " + strings.Repeat("─", innerWidth) + " ┐"
	}
	bottomRaw := "└" + strings.Repeat("─", innerWidth+2) + "┘"

	top := ctx.styler.Apply(topRaw, types.StyleIntent{Dim: true})
	bottom := ctx.styler.Apply(bottomRaw, types.StyleIntent{Dim: true})

	var boxLines []string
	for _, ln := range contentLines {
		pad := innerWidth - wrap.VisibleWidth(ln)
		if pad < 0 {
			pad = 0
		}
		left := ctx.styler.Apply("│ ", types.StyleIntent{Dim: true})
		right := ctx.styler.Apply(" │", types.StyleIntent{Dim: true})
		boxLines = append(boxLines, left+ln+strings.Repeat(" ", pad)+right)
	}

	return top + "\n" + strings.Join(boxLines, "\n") + "\n" + bottom + "\n\n"
}

// wrapCodeLine hard-wraps a single code line at the character level,
// accounting for wide (CJK/emoji) characters.
func wrapCodeLine(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}
	var parts []string
	var current strings.Builder
	currentWidth := 0
	for _, r := range text {
		rw := wrap.VisibleWidth(string(r))
		if currentWidth+rw > width {
			parts = append(parts, current.String())
			current.Reset()
			currentWidth = 0
		}
		current.WriteRune(r)
		currentWidth += rw
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	if len(parts) == 0 {
		return []string{""}
	}
	return parts
}
