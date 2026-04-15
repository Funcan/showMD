package render_test

import (
	"strings"
	"testing"

	"github.com/funcan/showmd/internal/render"
	"github.com/funcan/showmd/internal/types"
)

// Helper: render with color/hyperlinks disabled for deterministic output.
func renderPlain(md string) string {
	f := false
	return render.Render(md, types.RenderOptions{
		Color:      &f,
		Hyperlinks: &f,
	})
}

func stripPlain(md string) string {
	f := false
	return render.Strip(md, types.RenderOptions{
		Color:      &f,
		Hyperlinks: &f,
	})
}

// --- Headings ---

func TestRender_Heading(t *testing.T) {
	out := renderPlain("# Hello World\n")
	if !strings.Contains(out, "Hello World") {
		t.Errorf("expected heading text, got: %q", out)
	}
	// Heading should have leading newline.
	if !strings.HasPrefix(out, "\n") {
		t.Errorf("expected leading newline before heading, got: %q", out)
	}
}

func TestRender_HeadingLevels(t *testing.T) {
	for _, level := range []int{1, 2, 3, 4, 5, 6} {
		prefix := strings.Repeat("#", level)
		out := renderPlain(prefix + " Level " + strings.Repeat("x", level) + "\n")
		if !strings.Contains(out, "Level") {
			t.Errorf("level %d heading missing text, got: %q", level, out)
		}
	}
}

// --- Paragraphs ---

func TestRender_Paragraph(t *testing.T) {
	out := renderPlain("Hello world.\n")
	if !strings.Contains(out, "Hello world.") {
		t.Errorf("expected paragraph text, got: %q", out)
	}
}

func TestRender_ParagraphBold(t *testing.T) {
	out := renderPlain("This is **bold** text.\n")
	if !strings.Contains(out, "bold") {
		t.Errorf("expected bold text in output, got: %q", out)
	}
}

func TestRender_ParagraphEmphasis(t *testing.T) {
	out := renderPlain("This is _italic_ text.\n")
	if !strings.Contains(out, "italic") {
		t.Errorf("expected italic text in output, got: %q", out)
	}
}

func TestRender_ParagraphInlineCode(t *testing.T) {
	out := renderPlain("Use `fmt.Println()` here.\n")
	if !strings.Contains(out, "fmt.Println()") {
		t.Errorf("expected inline code text, got: %q", out)
	}
}

func TestRender_ParagraphStrikethrough(t *testing.T) {
	out := renderPlain("This is ~~struck~~ text.\n")
	if !strings.Contains(out, "struck") {
		t.Errorf("expected strikethrough text, got: %q", out)
	}
}

// --- HR ---

func TestRender_HR(t *testing.T) {
	out := renderPlain("---\n")
	if !strings.Contains(out, "—") {
		t.Errorf("expected em-dash in HR, got: %q", out)
	}
}

// --- Code blocks ---

func TestRender_FencedCodeBlock_Boxed(t *testing.T) {
	out := renderPlain("```go\nfunc main() {\n\tfmt.Println(\"hi\")\n}\n```\n")
	if !strings.Contains(out, "┌") {
		t.Errorf("expected box-drawing top-left, got: %q", out)
	}
	if !strings.Contains(out, "func main()") {
		t.Errorf("expected code content, got: %q", out)
	}
	if !strings.Contains(out, "[go]") {
		t.Errorf("expected language label, got: %q", out)
	}
}

func TestRender_FencedCodeBlock_NoBox_SingleLine(t *testing.T) {
	// Single-line code blocks should NOT be boxed (useBox requires len(lines) > 1).
	out := renderPlain("```\nhello\n```\n")
	if strings.Contains(out, "┌") {
		t.Errorf("single-line code block should not have box, got: %q", out)
	}
	if !strings.Contains(out, "hello") {
		t.Errorf("expected code content, got: %q", out)
	}
}

func TestRender_CodeBlock_DisabledBox(t *testing.T) {
	f := false
	ff := false
	out := render.Render("```go\nfunc main() {}\n```\n", types.RenderOptions{
		Color:      &ff,
		Hyperlinks: &ff,
		CodeBox:    &f,
	})
	if strings.Contains(out, "┌") {
		t.Errorf("expected no box when CodeBox=false, got: %q", out)
	}
	if !strings.Contains(out, "func main()") {
		t.Errorf("expected code content, got: %q", out)
	}
}

func TestRender_DiffCodeBlock(t *testing.T) {
	md := "```diff\n+added line\n-removed line\n context\n```\n"
	out := renderPlain(md)
	if !strings.Contains(out, "added line") {
		t.Errorf("expected diff content, got: %q", out)
	}
}

// --- Tables ---

func TestRender_Table_Unicode(t *testing.T) {
	md := "| A | B |\n|---|---|\n| 1 | 2 |\n"
	out := renderPlain(md)
	if !strings.Contains(out, "┌") {
		t.Errorf("expected unicode box in table, got: %q", out)
	}
	if !strings.Contains(out, "A") || !strings.Contains(out, "B") {
		t.Errorf("expected header cells, got: %q", out)
	}
	if !strings.Contains(out, "1") || !strings.Contains(out, "2") {
		t.Errorf("expected body cells, got: %q", out)
	}
}

func TestRender_Table_ASCII(t *testing.T) {
	f := false
	out := render.Render("| A | B |\n|---|---|\n| 1 | 2 |\n", types.RenderOptions{
		Color:       &f,
		Hyperlinks:  &f,
		TableBorder: types.TableBorderASCII,
	})
	if !strings.Contains(out, "+") {
		t.Errorf("expected ASCII border, got: %q", out)
	}
}

func TestRender_Table_None(t *testing.T) {
	f := false
	out := render.Render("| A | B |\n|---|---|\n| 1 | 2 |\n", types.RenderOptions{
		Color:       &f,
		Hyperlinks:  &f,
		TableBorder: types.TableBorderNone,
	})
	if strings.Contains(out, "┌") || strings.Contains(out, "+") {
		t.Errorf("expected no border, got: %q", out)
	}
	if !strings.Contains(out, "A") {
		t.Errorf("expected cell content, got: %q", out)
	}
}

// --- Lists ---

func TestRender_UnorderedList(t *testing.T) {
	out := renderPlain("- alpha\n- beta\n- gamma\n")
	for _, item := range []string{"alpha", "beta", "gamma"} {
		if !strings.Contains(out, item) {
			t.Errorf("expected list item %q, got: %q", item, out)
		}
	}
	if !strings.Contains(out, "-") {
		t.Errorf("expected bullet marker, got: %q", out)
	}
}

func TestRender_OrderedList(t *testing.T) {
	out := renderPlain("1. first\n2. second\n3. third\n")
	if !strings.Contains(out, "first") || !strings.Contains(out, "second") {
		t.Errorf("expected ordered list items, got: %q", out)
	}
	if !strings.Contains(out, "1.") {
		t.Errorf("expected ordered marker, got: %q", out)
	}
}

func TestRender_TaskList(t *testing.T) {
	out := renderPlain("- [x] done\n- [ ] todo\n")
	if !strings.Contains(out, "[x]") {
		t.Errorf("expected checked checkbox, got: %q", out)
	}
	if !strings.Contains(out, "[ ]") {
		t.Errorf("expected unchecked checkbox, got: %q", out)
	}
	if !strings.Contains(out, "done") || !strings.Contains(out, "todo") {
		t.Errorf("expected task text, got: %q", out)
	}
}

func TestRender_NestedList(t *testing.T) {
	out := renderPlain("- parent\n  - child\n")
	if !strings.Contains(out, "parent") || !strings.Contains(out, "child") {
		t.Errorf("expected nested list items, got: %q", out)
	}
}

// --- Blockquote ---

func TestRender_Blockquote(t *testing.T) {
	out := renderPlain("> This is a quote.\n")
	if !strings.Contains(out, "This is a quote.") {
		t.Errorf("expected quote text, got: %q", out)
	}
	if !strings.Contains(out, "│") {
		t.Errorf("expected quote prefix, got: %q", out)
	}
}

// --- Links ---

func TestRender_LinkWithLabel(t *testing.T) {
	out := renderPlain("[example](https://example.com)\n")
	if !strings.Contains(out, "example") {
		t.Errorf("expected link label, got: %q", out)
	}
	if !strings.Contains(out, "https://example.com") {
		t.Errorf("expected link URL in output, got: %q", out)
	}
}

func TestRender_LinkMailto(t *testing.T) {
	out := renderPlain("[email](mailto:foo@example.com)\n")
	// Mailto links should show as plain text.
	if !strings.Contains(out, "email") {
		t.Errorf("expected link label, got: %q", out)
	}
}

// --- Strip ---

func TestStrip_RemovesANSI(t *testing.T) {
	// Strip should produce plain text with no ANSI codes.
	out := stripPlain("# Hello\n\nworld\n")
	if strings.Contains(out, "\x1b") {
		t.Errorf("Strip should remove ANSI codes, got: %q", out)
	}
	if !strings.Contains(out, "Hello") {
		t.Errorf("expected heading text, got: %q", out)
	}
}

// --- CreateRenderer ---

func TestCreateRenderer(t *testing.T) {
	f := false
	renderer := render.CreateRenderer(types.RenderOptions{
		Color:      &f,
		Hyperlinks: &f,
	})
	out := renderer("# Test\n")
	if !strings.Contains(out, "Test") {
		t.Errorf("expected heading text, got: %q", out)
	}
}

// --- Dedent ---

func TestRender_Dedent(t *testing.T) {
	// Indented markdown should be dedented before parsing.
	out := renderPlain("    # Heading\n\n    paragraph\n")
	if !strings.Contains(out, "Heading") {
		t.Errorf("expected dedented heading, got: %q", out)
	}
	if !strings.Contains(out, "paragraph") {
		t.Errorf("expected dedented paragraph, got: %q", out)
	}
}

// --- Lang label heuristic ---

func TestRender_LangLabelHeuristic(t *testing.T) {
	md := "[go]\n\n```\nfunc main() {}\n```\n"
	out := renderPlain(md)
	// The [go] paragraph should not appear and the code block should be labelled go.
	if strings.Contains(out, "[go]\n") {
		t.Errorf("lang label paragraph should be consumed, got: %q", out)
	}
	if !strings.Contains(out, "func main()") {
		t.Errorf("expected code content, got: %q", out)
	}
}

// --- Wrap options ---

func TestRender_WrapDisabled(t *testing.T) {
	f := false
	ff := false
	out := render.Render("A very long paragraph that should not be wrapped at all because wrap is disabled.\n",
		types.RenderOptions{Color: &ff, Hyperlinks: &ff, Wrap: &f})
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 1 {
		t.Errorf("expected 1 line when wrap=false, got %d: %q", len(lines), out)
	}
}

func TestRender_WrapEnabled_WrapsAtWidth(t *testing.T) {
	f := false
	w := 10
	out := render.Strip("one two three four five six seven eight nine ten",
		types.RenderOptions{Color: &f, Hyperlinks: &f, Wrap: boolTrue(), Width: &w})
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if len(line) > 10 {
			t.Errorf("line exceeds width 10: %q", line)
		}
	}
}

// --- Soft and hard line breaks ---

func TestRender_SoftLineBreak_CollapsesToSpace(t *testing.T) {
	f := false
	w := 80
	out := strings.TrimRight(render.Strip("Hello\nworld",
		types.RenderOptions{Color: &f, Hyperlinks: &f, Wrap: boolTrue(), Width: &w}), "\n")
	if out != "Hello world" {
		t.Errorf("soft break should collapse to space, got %q", out)
	}
}

func TestRender_SoftLineBreak_TrimsIndentation(t *testing.T) {
	f := false
	w := 200
	out := strings.TrimRight(render.Strip("Hello\n  world",
		types.RenderOptions{Color: &f, Hyperlinks: &f, Width: &w}), "\n")
	if out != "Hello world" {
		t.Errorf("soft break should trim indentation, got %q", out)
	}
}

func TestRender_HardLineBreak_KeepsSplit(t *testing.T) {
	f := false
	w := 80
	out := strings.TrimRight(render.Strip("line one  \nline two",
		types.RenderOptions{Color: &f, Hyperlinks: &f, Wrap: boolTrue(), Width: &w}), "\n")
	if len(strings.Split(out, "\n")) < 2 {
		t.Errorf("hard break (two-space) should produce multiple lines, got %q", out)
	}
}

func TestRender_HardBreakWithSurroundingSoftBreaks(t *testing.T) {
	f := false
	w := 200
	out := strings.TrimRight(render.Strip("a\nb  \nc",
		types.RenderOptions{Color: &f, Hyperlinks: &f, Width: &w}), "\n")
	if !strings.Contains(out, "a b\nc") {
		t.Errorf("expected 'a b\\nc', got %q", out)
	}
}

// --- Inline HTML ---

func TestRender_InlineHTML_Ignored(t *testing.T) {
	out := strings.TrimRight(stripPlain("<div>ignored</div>"), "\n")
	if out != "" {
		t.Errorf("inline HTML should be ignored, got %q", out)
	}
}

// --- Hyperlinks ---

func TestRender_Hyperlinks_URLSuffix_WhenOff(t *testing.T) {
	out := stripPlain("[link](https://example.com)")
	if !strings.Contains(out, "link (https://example.com)") {
		t.Errorf("expected URL suffix when hyperlinks off, got %q", out)
	}
}

func TestRender_Hyperlinks_OSC8_WhenEnabled(t *testing.T) {
	tr := true
	out := render.Render("[x](https://example.com)", types.RenderOptions{
		Color:      &tr,
		Hyperlinks: &tr,
		Wrap:       boolFalse(),
	})
	if !strings.Contains(out, "\x1b]8;;https://example.com\x07x\x1b]8;;\x07") {
		t.Errorf("expected OSC-8 hyperlink, got %q", out)
	}
}

func TestRender_Hyperlinks_OSC8_DisabledWhenColorFalse(t *testing.T) {
	fl := false
	tr := true
	out := render.Render("[x](https://example.com)", types.RenderOptions{
		Color:      &fl,
		Hyperlinks: &tr,
		Wrap:       boolFalse(),
	})
	if strings.Contains(out, "\x1b]8;;") {
		t.Errorf("OSC-8 should be disabled when color=false, got %q", out)
	}
	if !strings.Contains(out, "x (https://example.com)") {
		t.Errorf("expected plain URL suffix when color=false, got %q", out)
	}
}

// --- Loose lists ---

func TestRender_LooseList_HasBlankLineBetweenItems(t *testing.T) {
	out := stripPlain("- item 1\n\n- item 2")
	blanks := 0
	for _, l := range strings.Split(out, "\n") {
		if l == "" {
			blanks++
		}
	}
	if blanks == 0 {
		t.Errorf("loose list should have blank lines, got %q", out)
	}
}

func TestRender_List_SoftBreakInsideItem(t *testing.T) {
	md := "- Section IV: A concluding line that was \"typed on 2025-12-18 with a\n  stubborn cursor.\""
	f := false
	w := 200
	out := strings.TrimRight(render.Strip(md, types.RenderOptions{Color: &f, Hyperlinks: &f, Width: &w}), "\n")
	if !strings.Contains(out, "with a stubborn cursor.") {
		t.Errorf("soft break inside list item should collapse, got %q", out)
	}
	if strings.Contains(out, "\n\n") {
		t.Errorf("should not have double blank lines, got %q", out)
	}
}

// --- Theme/color ---

func TestRender_Theme_InlineVsBlockCodeDistinct(t *testing.T) {
	inlineColor := "\x1b[31m" // red
	blockColor := "\x1b[32m"  // green
	tr := true
	fl := false
	out := render.Render("`inline`\n\n```\nblock\n```", types.RenderOptions{
		Color:      &tr,
		Hyperlinks: &fl,
		Wrap:       boolFalse(),
		Theme: &types.Theme{
			InlineCode: types.StyleIntent{Color: "red"},
			BlockCode:  types.StyleIntent{Color: "green"},
		},
	})
	if !strings.Contains(out, inlineColor) {
		t.Errorf("expected red inline code, got %q", out)
	}
	if !strings.Contains(out, blockColor) {
		t.Errorf("expected green block code, got %q", out)
	}
}

func TestRender_Theme_CodeFallback(t *testing.T) {
	tr := true
	fl := false
	out := render.Render("`x`\n\n```\ny\n```", types.RenderOptions{
		Color:      &tr,
		Hyperlinks: &fl,
		Wrap:       boolFalse(),
		Theme:      &types.Theme{Code: types.StyleIntent{Color: "red"}},
	})
	if !strings.Contains(out, "\x1b[31m") {
		t.Errorf("expected red from code fallback, got %q", out)
	}
}

func TestRender_DefaultTheme_Colors(t *testing.T) {
	tr := true
	fl := false
	out := render.Render("`inline`\n\n```\nblock\n```\n\n# H", types.RenderOptions{
		Color:      &tr,
		Hyperlinks: &fl,
		Wrap:       boolFalse(),
		CodeBox:    boolFalse(),
	})
	if !strings.Contains(out, "\x1b[36m") { // cyan inline code
		t.Errorf("expected cyan inline code, got %q", out)
	}
	if !strings.Contains(out, "\x1b[32m") { // green block code
		t.Errorf("expected green block code, got %q", out)
	}
}

func TestRender_HighlighterHook(t *testing.T) {
	tr := true
	fl := false
	out := render.Render("```\ncode\n```", types.RenderOptions{
		Color:       &tr,
		Hyperlinks:  &fl,
		Wrap:        boolFalse(),
		Highlighter: func(code, _ string) string { return strings.ToUpper(code) },
	})
	if !strings.Contains(out, "CODE") {
		t.Errorf("expected highlighter to uppercase code, got %q", out)
	}
}

// helpers for bool pointers not in renderPlain scope

func boolTrue() *bool  { t := true; return &t }
func boolFalse() *bool { f := false; return &f }
