package render_test

import (
	"strings"
	"testing"

	"github.com/funcan/showmd/internal/render"
	"github.com/funcan/showmd/internal/types"
	"github.com/funcan/showmd/internal/wrap"
)

// noColorCode returns base options for code block tests.
func noColorCode() types.RenderOptions {
	f := false
	w := 40
	return types.RenderOptions{
		Color:      &f,
		Hyperlinks: &f,
		Wrap:       boolTrue(),
		Width:      &w,
	}
}

func TestCodeBlock_WrapsWhenCodeWrapEnabled(t *testing.T) {
	md := "```\n0123456789ABCDEFG\n```"
	w := 12
	opts := noColorCode()
	opts.Width = &w
	opts.CodeBox = boolFalse()
	out := render.Strip(md, opts)
	firstLine := strings.Split(out, "\n")[0]
	if len(firstLine) > 12 {
		t.Errorf("line should fit within width 12, got %q (len %d)", firstLine, len(firstLine))
	}
	if !strings.Contains(out, "0123456789") {
		t.Errorf("expected code content, got %q", out)
	}
}

func TestCodeBlock_NoWrap_AllowsOverflow(t *testing.T) {
	md := "```\n0123456789ABCDEFG\n```"
	w := 10
	opts := noColorCode()
	opts.Width = &w
	opts.CodeWrap = boolFalse()
	opts.CodeBox = boolFalse()
	out := render.Strip(md, opts)
	firstLine := strings.Split(out, "\n")[0]
	if len(firstLine) <= 15 {
		t.Errorf("expected overflow line (>15 chars) when codeWrap=false, got %q", firstLine)
	}
}

func TestCodeBlock_GutterWithNoBox(t *testing.T) {
	md := "```\nline1\nline2\n```"
	tr := true
	out := render.Render(md, types.RenderOptions{
		Color:      &tr,
		Hyperlinks: boolFalse(),
		CodeBox:    boolFalse(),
		CodeGutter: true,
		Wrap:       boolFalse(),
	})
	plain := wrap.StripANSI(out)
	if !strings.Contains(plain, "1") || !strings.Contains(plain, "line1") {
		t.Errorf("expected gutter with line numbers, got %q", plain)
	}
}

func TestCodeBlock_GutterMultiDigit(t *testing.T) {
	lines := make([]string, 12)
	for i := range lines {
		lines[i] = "l" + string(rune('0'+i+1))
	}
	md := "```\n" + strings.Join(lines, "\n") + "\n```"
	tr := true
	out := render.Render(md, types.RenderOptions{
		Color:      &tr,
		Hyperlinks: boolFalse(),
		CodeGutter: true,
		Wrap:       boolFalse(),
	})
	plain := wrap.StripANSI(out)
	if !strings.Contains(plain, "12") {
		t.Errorf("expected 2-digit line number, got %q", plain)
	}
}

func TestCodeBlock_LangLabelInHeader(t *testing.T) {
	md := "```bash\nthis line is definitely longer than the label\nand still flows\n```"
	tr := true
	out := render.Render(md, types.RenderOptions{
		Color:      &tr,
		Hyperlinks: boolFalse(),
		Wrap:       boolFalse(),
	})
	firstLine := wrap.StripANSI(strings.Split(out, "\n")[0])
	if !strings.HasPrefix(firstLine, "┌") {
		t.Errorf("expected box top, got %q", firstLine)
	}
	if !strings.Contains(firstLine, "[bash]") {
		t.Errorf("expected [bash] label, got %q", firstLine)
	}
	// Border should be dimmed
	if !strings.Contains(strings.Split(out, "\n")[0], "\x1b[2m") {
		t.Errorf("expected dim border, raw: %q", strings.Split(out, "\n")[0])
	}
}

func TestCodeBlock_NoLabelWhenNoLang(t *testing.T) {
	md := "```\nfoo\nbar\n```"
	out := render.Render(md, types.RenderOptions{
		Color:      boolFalse(),
		Hyperlinks: boolFalse(),
		Wrap:       boolFalse(),
	})
	firstLine := strings.Split(out, "\n")[0]
	if !strings.HasPrefix(firstLine, "┌") {
		t.Errorf("expected box top, got %q", firstLine)
	}
	if strings.Contains(firstLine, "[") {
		t.Errorf("should not have lang label when no language, got %q", firstLine)
	}
}

func TestCodeBlock_LongLangLabel_HeaderWidthAdapts(t *testing.T) {
	md := "```superlonglanguageid\nfoo\nbar\n```"
	out := render.Render(md, types.RenderOptions{
		Color:      boolFalse(),
		Hyperlinks: boolFalse(),
		Wrap:       boolFalse(),
	})
	lines := strings.Split(out, "\n")
	top := lines[0]
	body := lines[1]
	if len(top) < len(body)-2 {
		t.Errorf("header should be >= body width, top=%d body=%d:\n%s", len(top), len(body), out)
	}
	if !strings.Contains(top, "[superlonglanguageid]") {
		t.Errorf("expected long label in header, got %q", top)
	}
}

func TestCodeBlock_NoBlankLineBeforeBox(t *testing.T) {
	md := "```bash\nfoo\nbar\n```"
	out := render.Render(md, types.RenderOptions{
		Color:      boolFalse(),
		Hyperlinks: boolFalse(),
		Wrap:       boolFalse(),
	})
	if !strings.HasPrefix(out, "┌") {
		t.Errorf("expected box to start at beginning (no blank line), got %q", out)
	}
}

func TestCodeBlock_ReferenceContinuation_NotBoxed(t *testing.T) {
	md := "[1]: https://example.com/icon \"\n\t Icon Composer Notes \n\t\n\""
	out := render.Render(md, types.RenderOptions{
		Color:      boolFalse(),
		Hyperlinks: boolFalse(),
		Wrap:       boolTrue(),
	})
	if strings.Contains(out, "┌") {
		t.Errorf("reference continuation should not be boxed, got %q", out)
	}
	if !strings.Contains(out, `[1]: https://example.com/icon "`) {
		t.Errorf("expected reference definition, got %q", out)
	}
}

func TestCodeBlock_DefinitionHasBlankLineBefore(t *testing.T) {
	md := "Body line.\n[1]: https://example.com \"Title\"\nNext."
	out := render.Render(md, types.RenderOptions{
		Color:      boolFalse(),
		Hyperlinks: boolFalse(),
		Wrap:       boolTrue(),
	})
	lines := strings.Split(out, "\n")
	if lines[0] != "Body line." {
		t.Errorf("expected 'Body line.' first, got %q", lines[0])
	}
	if lines[1] != "" {
		t.Errorf("expected blank line before definition, got %q", lines[1])
	}
	if lines[2] != `[1]: https://example.com "Title"` {
		t.Errorf("expected definition on line 2, got %q", lines[2])
	}
	if lines[3] != "Next." {
		t.Errorf("expected Next. on line 3, got %q", lines[3])
	}
}

func TestCodeBlock_ListOfBlocks_CollapsedToOne(t *testing.T) {
	md := "- ```\n  first\n  ```\n- ```\n  second\n  ```"
	out := render.Render(md, types.RenderOptions{
		Color:      boolFalse(),
		Hyperlinks: boolFalse(),
		Wrap:       boolFalse(),
	})
	boxes := strings.Count(out, "┌")
	if boxes != 1 {
		t.Errorf("expected 1 merged code box, got %d:\n%s", boxes, out)
	}
	if !strings.Contains(out, "first") || !strings.Contains(out, "second") {
		t.Errorf("expected both code blocks in merged box, got %q", out)
	}
}

func TestCodeBlock_DiffTagged_NoWrap(t *testing.T) {
	md := "```\n--- a/foo\n+++ b/foo\n@@ -1 +1 @@\n- a very very very very long line\n+ another very very very very long line\n```"
	w := 20
	out := render.Render(md, types.RenderOptions{
		Color:      boolFalse(),
		Hyperlinks: boolFalse(),
		Wrap:       boolTrue(),
		Width:      &w,
	})
	if !strings.Contains(out, "[diff]") {
		t.Errorf("expected [diff] label, got %q", out)
	}
	var longLine string
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "very very very very long line") {
			longLine = line
			break
		}
	}
	if longLine == "" {
		t.Errorf("expected long diff line, got %q", out)
	}
	visible := wrap.VisibleWidth(longLine)
	if visible <= 30 {
		t.Errorf("diff lines should not be wrapped, got visible width %d: %q", visible, longLine)
	}
}

func TestCodeBlock_SingleLine_NoBox(t *testing.T) {
	md := "```\nsolo\n```"
	out := render.Render(md, types.RenderOptions{
		Color:      boolFalse(),
		Hyperlinks: boolFalse(),
		Wrap:       boolFalse(),
	})
	if strings.Contains(out, "┌") {
		t.Errorf("single-line code should not be boxed, got %q", out)
	}
	if strings.TrimSpace(out) != "solo" {
		t.Errorf("expected 'solo', got %q", out)
	}
}
