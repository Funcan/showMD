package render

import (
	"strings"

	"github.com/funcan/showmd/internal/types"
	"github.com/funcan/showmd/internal/wrap"
	"github.com/yuin/goldmark/ast"
	east "github.com/yuin/goldmark/extension/ast"
)

// tableBoxStyle holds the box-drawing characters for a table border style.
type tableBoxStyle struct {
	topLeft     string
	topRight    string
	bottomLeft  string
	bottomRight string
	hSep        string
	vSep        string
	tSep        string
	mSep        string
	bSep        string
	mLeft       string
	mRight      string
}

var (
	tableBoxUnicode = tableBoxStyle{
		topLeft: "┌", topRight: "┐", bottomLeft: "└", bottomRight: "┘",
		hSep: "─", vSep: "│", tSep: "┬", mSep: "┼", bSep: "┴",
		mLeft: "├", mRight: "┤",
	}
	tableBoxASCII = tableBoxStyle{
		topLeft: "+", topRight: "+", bottomLeft: "+", bottomRight: "+",
		hSep: "-", vSep: "|", tSep: "+", mSep: "+", bSep: "+",
		mLeft: "+", mRight: "+",
	}
)

// renderTable renders a GFM table node.
func renderTable(n ast.Node, ctx *renderContext) string {
	tbl := n.(*east.Table)
	aligns := tbl.Alignments

	headerNode := n.FirstChild()
	if headerNode == nil {
		return ""
	}

	// Render all cells (header + body).
	type row []string
	renderRowCells := func(rowNode ast.Node) row {
		var cells row
		for cell := rowNode.FirstChild(); cell != nil; cell = cell.NextSibling() {
			cells = append(cells, renderInlineContent(cell, ctx))
		}
		return cells
	}

	headerCells := renderRowCells(headerNode)
	var bodyRows []row
	for rowNode := headerNode.NextSibling(); rowNode != nil; rowNode = rowNode.NextSibling() {
		bodyRows = append(bodyRows, renderRowCells(rowNode))
	}

	allRows := append([]row{headerCells}, bodyRows...)

	// Determine column count.
	colCount := 0
	for _, r := range allRows {
		if len(r) > colCount {
			colCount = len(r)
		}
	}
	if colCount == 0 {
		return ""
	}

	pad := ctx.opts.tablePadding
	padStr := strings.Repeat(" ", max0(pad))
	ellipsis := ctx.opts.tableEllipsis
	minContent := max1(len([]rune(ellipsis)) + 1)
	minColWidth := max1(pad*2 + minContent)

	// Calculate initial column widths.
	widths := make([]int, colCount)
	for i := range widths {
		widths[i] = 1
	}
	for _, r := range allRows {
		for idx, cell := range r {
			if idx >= colCount {
				break
			}
			padded := padStr + cell + padStr
			w := min_(maxCol, wrap.VisibleWidth(padded))
			if w > widths[idx] {
				widths[idx] = w
			}
		}
	}

	// Shrink columns if table is too wide.
	if ctx.opts.wrap && ctx.opts.width > 0 {
		totalWidth := sum(widths) + 3*colCount + 1
		over := totalWidth - ctx.opts.width
		for over > 0 {
			maxW := widths[0]
			maxI := 0
			for i, w := range widths {
				if w > maxW {
					maxW = w
					maxI = i
				}
			}
			if widths[maxI] <= minColWidth {
				break
			}
			widths[maxI]--
			over--
		}
	}
	for i := range widths {
		if widths[i] < minColWidth {
			widths[i] = minColWidth
		}
	}

	// renderTableRow renders a row of pre-rendered cell strings.
	// Returns a slice of line-slices (one per row line due to wrapping).
	renderTableRow := func(cells []string, isHeader bool) [][]string {
		linesPerCol := make([][]string, colCount)
		for idx := 0; idx < colCount; idx++ {
			cell := ""
			if idx < len(cells) {
				cell = cells[idx]
			}
			target := max1(widths[idx] - pad*2)
			var content string
			if ctx.opts.tableTruncate {
				content = truncateCell(cell, target, ellipsis)
			} else {
				content = cell
			}
			wrapWidth := target
			if !ctx.opts.wrap {
				wrapWidth = 1<<31 - 1
			}
			wrapped := wrap.WrapText(content, wrapWidth, ctx.opts.wrap)
			align := getAlign(aligns, idx)
			colLines := make([]string, len(wrapped))
			for i, l := range wrapped {
				aligned := padCell(l, target, align)
				padded := padStr + aligned + padStr
				colLines[i] = padCell(padded, widths[idx], east.AlignLeft)
			}
			linesPerCol[idx] = colLines
		}

		// Determine row height (max lines across columns).
		height := 0
		for _, col := range linesPerCol {
			if len(col) > height {
				height = len(col)
			}
		}
		if height == 0 {
			height = 1
		}

		var out [][]string
		for i := 0; i < height; i++ {
			parts := make([]string, colCount)
			for idx := 0; idx < colCount; idx++ {
				align := getAlign(aligns, idx)
				var content string
				if i < len(linesPerCol[idx]) {
					content = linesPerCol[idx][i]
				} else {
					content = padCell("", widths[idx], align)
				}
				if isHeader {
					parts[idx] = ctx.styler.Apply(content, ctx.opts.theme.TableHeader)
				} else {
					parts[idx] = ctx.styler.Apply(content, ctx.opts.theme.TableCell)
				}
			}
			out = append(out, parts)
		}
		return out
	}

	headerRows := renderTableRow(headerCells, true)
	var allBodyRows [][]string
	for _, r := range bodyRows {
		allBodyRows = append(allBodyRows, renderTableRow(r, false)...)
	}

	if ctx.opts.tableBorder == types.TableBorderNone {
		var lines []string
		for _, r := range append(headerRows, allBodyRows...) {
			lines = append(lines, strings.Join(r, " | "))
		}
		return strings.Join(lines, "\n") + "\n\n"
	}

	var box tableBoxStyle
	if ctx.opts.tableBorder == types.TableBorderASCII {
		box = tableBoxASCII
	} else {
		box = tableBoxUnicode
	}

	hLine := func(sepMid, sepLeft, sepRight string) string {
		cols := make([]string, colCount)
		for i, w := range widths {
			cols[i] = strings.Repeat(box.hSep, w)
		}
		return sepLeft + strings.Join(cols, sepMid) + sepRight + "\n"
	}

	top := hLine(box.tSep, box.topLeft, box.topRight)
	mid := hLine(box.mSep, box.mLeft, box.mRight)
	bottom := hLine(box.bSep, box.bottomLeft, box.bottomRight)

	renderFlat := func(rows [][]string) string {
		var sb strings.Builder
		for _, r := range rows {
			sb.WriteString(box.vSep)
			sb.WriteString(strings.Join(r, box.vSep))
			sb.WriteString(box.vSep)
			sb.WriteByte('\n')
		}
		return sb.String()
	}

	var sb strings.Builder
	sb.WriteString(top)
	sb.WriteString(renderFlat(headerRows))
	if !ctx.opts.tableDense {
		sb.WriteString(mid)
	}
	sb.WriteString(renderFlat(allBodyRows))
	sb.WriteString(bottom)
	sb.WriteByte('\n')
	return sb.String()
}

// padCell pads text to width columns with the given alignment.
func padCell(text string, width int, align east.Alignment) string {
	visible := wrap.VisibleWidth(text)
	pad := width - visible
	if pad <= 0 {
		return text
	}
	switch align {
	case east.AlignRight:
		return strings.Repeat(" ", pad) + text
	case east.AlignCenter:
		left := pad / 2
		right := pad - left
		return strings.Repeat(" ", left) + text + strings.Repeat(" ", right)
	default:
		return text + strings.Repeat(" ", pad)
	}
}

// truncateCell truncates text to width visible columns, adding ellipsis.
func truncateCell(text string, width int, ellipsis string) string {
	if wrap.VisibleWidth(text) <= width {
		return text
	}
	ellipsisRunes := []rune(ellipsis)
	if width <= len(ellipsisRunes) {
		if width < len(ellipsisRunes) {
			return string(ellipsisRunes[:width])
		}
		return ellipsis
	}
	target := width - len(ellipsisRunes)
	runes := []rune(text)
	if target > len(runes) {
		target = len(runes)
	}
	return string(runes[:target]) + ellipsis
}

// getAlign returns the alignment for column idx, defaulting to AlignLeft.
func getAlign(aligns []east.Alignment, idx int) east.Alignment {
	if idx < len(aligns) {
		return aligns[idx]
	}
	return east.AlignLeft
}

// Helper math functions.

func max0(n int) int {
	if n < 0 {
		return 0
	}
	return n
}

func max1(n int) int {
	if n < 1 {
		return 1
	}
	return n
}

func min_(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func sum(ns []int) int {
	total := 0
	for _, n := range ns {
		total += n
	}
	return total
}
