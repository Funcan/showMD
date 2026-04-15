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
