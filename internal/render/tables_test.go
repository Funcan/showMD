package render_test

import (
	"strings"
	"testing"

	"github.com/funcan/showmd/internal/render"
	"github.com/funcan/showmd/internal/types"
)

// noColorTable is a base option set for table tests.
func noColorTable(extraOpts ...func(*types.RenderOptions)) types.RenderOptions {
	opts := types.RenderOptions{
		Wrap: boolTrue(),
	}
	for _, fn := range extraOpts {
		fn(&opts)
	}
	return opts
}

func tableStrip(md string, width int, opts types.RenderOptions) string {
	return render.Render(md, types.Terminal{Width: width}, opts)
}

func TestTable_LongCellUsesAvailableWidth(t *testing.T) {
	long := strings.Repeat("word ", 12) + "end" // 63 columns
	md := "| A | B |\n| --- | --- |\n| x | " + long + " |\n"
	out := tableStrip(md, 100, noColorTable())
	if !strings.Contains(out, long) {
		t.Errorf("expected untruncated cell when it fits the width, got:\n%s", out)
	}
}

func TestTable_DoesNotLinkifyUnderscores(t *testing.T) {
	md := "| Filename | Size |\n| --- | --- |\n| icon_16x16.png | 16 |\n| icon_16x16@2x.png | 32 |\n"
	opts := noColorTable(func(o *types.RenderOptions) {
		o.TableTruncate = boolFalse()
	})
	out := tableStrip(md, 40, opts)
	for _, line := range strings.Split(out, "\n") {
		if !strings.Contains(line, "icon_16x16") {
			continue
		}
		if strings.Contains(line, "http") {
			t.Errorf("should not linkify underscores in table: %q", line)
		}
		if !strings.Contains(line, "icon_16x16") {
			t.Errorf("expected icon_16x16 in line: %q", line)
		}
	}
}

func TestTable_InlineLinkCells(t *testing.T) {
	md := "| File | Link |\n| --- | --- |\n| icon_16x16.png | https://example.com/icon.png |\n"
	w := 60
	opts := noColorTable(func(o *types.RenderOptions) {
		o.TableTruncate = boolFalse()
	})
	out := tableStrip(md, w, opts)
	if !strings.Contains(out, "icon_16x16.png") {
		t.Errorf("expected filename in output, got %q", out)
	}
	if !strings.Contains(out, "https://example.com/icon.png") {
		t.Errorf("expected URL in output, got %q", out)
	}
}

func TestTable_AutolinkEmailPlain(t *testing.T) {
	md := "| File | Size |\n| --- | --- |\n| icon_16x16@2x.png | 32 |\n"
	opts := noColorTable(func(o *types.RenderOptions) {
		o.TableTruncate = boolFalse()
	})
	out := render.Render(md, types.Terminal{Width: 40, Color: types.ColorTrue, Hyperlinks: true}, opts)
	if !strings.Contains(out, "icon_16x16@2x.png") {
		t.Errorf("expected filename in output, got %q", out)
	}
	if strings.Contains(out, "\x1b]8;;") {
		t.Errorf("should not produce OSC-8 hyperlink for @-style filename, got %q", out)
	}
}

func TestTable_ASCIIBorder(t *testing.T) {
	md := "| h1 | h2 |\n| --- | --- |\n| a | b |\n"
	opts := noColorTable(func(o *types.RenderOptions) {
		o.TableBorder = types.TableBorderASCII
	})
	out := tableStrip(md, 40, opts)
	if !strings.Contains(out, "+") {
		t.Errorf("expected '+' in ASCII border, got %q", out)
	}
	if !strings.Contains(out, "h1") || !strings.Contains(out, "a") {
		t.Errorf("expected cell content, got %q", out)
	}
}

func TestTable_NoBorder_Dense(t *testing.T) {
	md := "| h1 | h2 |\n| --- | --- |\n| a | b |\n"
	p := 0
	opts := noColorTable(func(o *types.RenderOptions) {
		o.TableBorder = types.TableBorderNone
		o.TableDense = true
		o.TablePadding = &p
	})
	out := tableStrip(md, 40, opts)
	if strings.Contains(out, "┌") || strings.Contains(out, "+") {
		t.Errorf("expected no border chars, got %q", out)
	}
	if !strings.Contains(out, "h1") || !strings.Contains(out, "b") {
		t.Errorf("expected cell content, got %q", out)
	}
}

func TestTable_GFMBasic(t *testing.T) {
	md := "| h1 | h2 |\n| --- | --- |\n| a | b |\n"
	w := 30
	opts := noColorTable()
	out := tableStrip(md, w, opts)
	if !strings.Contains(out, "h1") || !strings.Contains(out, "h2") {
		t.Errorf("expected headers, got %q", out)
	}
}

func TestTable_CellWrapping(t *testing.T) {
	md := "| h1 | h2 |\n| --- | --- |\n| a b c d e f | g |\n"
	w := 15
	opts := noColorTable(func(o *types.RenderOptions) {
		o.TableTruncate = boolFalse()
	})
	out := tableStrip(md, w, opts)
	lines := strings.Split(strings.TrimSpace(out), "\n")
	borderLines := 0
	for _, l := range lines {
		if strings.Contains(l, "│") || strings.Contains(l, "|") {
			borderLines++
		}
	}
	if borderLines <= 3 {
		t.Errorf("expected wrapped rows to add lines, got %d border lines:\n%s", borderLines, out)
	}
}

func TestTable_WrappedCells_SecondColPreserved(t *testing.T) {
	md := "| h1 | h2 |\n| --- | --- |\n| a b c d e f | g |\n"
	w := 15
	opts := noColorTable(func(o *types.RenderOptions) {
		o.TableTruncate = boolFalse()
	})
	out := tableStrip(md, w, opts)
	found := false
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "g") && strings.Contains(line, "a") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'g' and 'a' on same row in wrapped table:\n%s", out)
	}
}

func TestTable_NoTruncationWhenContentFits(t *testing.T) {
	md := "| Provider | Enabled | Configured | Detail |\n| --- | --- | --- | --- |\n| WhatsApp | ON | WARN | not linked |\n"
	w := 120
	opts := noColorTable()
	out := tableStrip(md, w, opts)
	if !strings.Contains(out, "Provider") {
		t.Errorf("expected header, got %q", out)
	}
	if strings.Contains(out, "Provi…") {
		t.Errorf("should not truncate when content fits, got %q", out)
	}
}

func TestTable_LongWordOverflows_NoBreak(t *testing.T) {
	word := "Supercalifragilistic"
	md := "| h1 | h2 |\n| --- | --- |\n| " + word + " | x |\n"
	w := 10
	opts := noColorTable(func(o *types.RenderOptions) {
		o.TableTruncate = boolFalse()
	})
	out := tableStrip(md, w, opts)
	if !strings.Contains(out, word) {
		t.Errorf("long word should overflow without hard break, got %q", out)
	}
}

func TestTable_AlignmentMarkers(t *testing.T) {
	md := "| h1 | h2 |\n| :-- | --: |\n| left | right |\n"
	w := 30
	opts := noColorTable(func(o *types.RenderOptions) {
		o.TableTruncate = boolFalse()
	})
	out := tableStrip(md, w, opts)
	found := false
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "left") && strings.Contains(line, "right") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected left and right on same row, got %q", out)
	}
}

func TestTable_Truncation_DefaultOn(t *testing.T) {
	md := "| col | col2 |\n| --- | --- |\n| Supercalifragilistic | short |\n"
	w := 18
	opts := noColorTable()
	out := tableStrip(md, w, opts)
	if !strings.Contains(out, "…") {
		t.Errorf("expected truncation ellipsis at tight width, got %q", out)
	}
}

func TestTable_ExtremelySmallWidth_HasEllipsis(t *testing.T) {
	md := "| h |\n| - |\n| verylong |\n"
	w := 4
	p := 0
	opts := noColorTable(func(o *types.RenderOptions) {
		o.TablePadding = &p
	})
	out := tableStrip(md, w, opts)
	if !strings.Contains(out, "…") {
		t.Errorf("expected ellipsis at extreme width, got %q", out)
	}
}

func TestTable_Alignment_LeftCenterRight(t *testing.T) {
	md := "| l | c | r |\n| :-- | :-: | --: |\n| a | b | c |\n"
	w := 40
	opts := noColorTable()
	out := tableStrip(md, w, opts)
	var found string
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "a") && strings.Contains(line, "c") {
			found = line
			break
		}
	}
	if found == "" {
		t.Errorf("expected row with all three cells, got %q", out)
	}
	if !strings.Contains(found, " b ") {
		t.Errorf("expected center cell 'b' to be padded, got %q", found)
	}
}

func TestTable_CustomPadding(t *testing.T) {
	md := "| h1 | h2 |\n| --- | --- |\n| a | b |\n"
	p := 2
	opts := noColorTable(func(o *types.RenderOptions) {
		o.TablePadding = &p
	})
	out := tableStrip(md, 40, opts)
	if !strings.Contains(out, "h1") {
		t.Errorf("expected table content with padding 2, got %q", out)
	}
}
