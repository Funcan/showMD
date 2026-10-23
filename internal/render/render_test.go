package render_test

import (
	"github.com/charmbracelet/colorprofile"
	"strings"
	"testing"

	"github.com/funcan/showmd/internal/render"
	"github.com/funcan/showmd/internal/types"
)

// Helper: render with color/hyperlinks disabled for deterministic output.
func renderPlain(md string) string {
	return render.Render(md, types.Terminal{}, types.RenderOptions{})
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
	out := render.Render("```go\nfunc main() {}\n```\n", types.Terminal{}, types.RenderOptions{
		CodeBox: &f,
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
	out := render.Render("| A | B |\n|---|---|\n| 1 | 2 |\n", types.Terminal{}, types.RenderOptions{
		TableBorder: types.TableBorderASCII,
	})
	if !strings.Contains(out, "+") {
		t.Errorf("expected ASCII border, got: %q", out)
	}
}

func TestRender_Table_None(t *testing.T) {
	out := render.Render("| A | B |\n|---|---|\n| 1 | 2 |\n", types.Terminal{}, types.RenderOptions{
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
	if !strings.Contains(out, "> ") {
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

// --- Plain output ---

func TestRender_ZeroTerminalIsPlain(t *testing.T) {
	// The zero Terminal should produce plain text with no ANSI codes.
	out := renderPlain("# Hello\n\nworld\n")
	if strings.Contains(out, "\x1b") {
		t.Errorf("expected no ANSI codes, got: %q", out)
	}
	if !strings.Contains(out, "Hello") {
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
	out := render.Render("A very long paragraph that should not be wrapped at all because wrap is disabled.\n",
		types.Terminal{}, types.RenderOptions{Wrap: &f})
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 1 {
		t.Errorf("expected 1 line when wrap=false, got %d: %q", len(lines), out)
	}
}

func TestRender_WrapEnabled_WrapsAtWidth(t *testing.T) {
	w := 10
	out := render.Render("one two three four five six seven eight nine ten",
		types.Terminal{Width: w}, types.RenderOptions{Wrap: boolTrue()})
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if len(line) > 10 {
			t.Errorf("line exceeds width 10: %q", line)
		}
	}
}

// --- Soft and hard line breaks ---

func TestRender_SoftLineBreak_CollapsesToSpace(t *testing.T) {
	w := 80
	out := strings.TrimRight(render.Render("Hello\nworld",
		types.Terminal{Width: w}, types.RenderOptions{Wrap: boolTrue()}), "\n")
	if out != "Hello world" {
		t.Errorf("soft break should collapse to space, got %q", out)
	}
}

func TestRender_SoftLineBreak_TrimsIndentation(t *testing.T) {
	w := 200
	out := strings.TrimRight(render.Render("Hello\n  world",
		types.Terminal{Width: w}, types.RenderOptions{}), "\n")
	if out != "Hello world" {
		t.Errorf("soft break should trim indentation, got %q", out)
	}
}

func TestRender_HardLineBreak_KeepsSplit(t *testing.T) {
	w := 80
	out := strings.TrimRight(render.Render("line one  \nline two",
		types.Terminal{Width: w}, types.RenderOptions{Wrap: boolTrue()}), "\n")
	if len(strings.Split(out, "\n")) < 2 {
		t.Errorf("hard break (two-space) should produce multiple lines, got %q", out)
	}
}

func TestRender_HardBreakWithSurroundingSoftBreaks(t *testing.T) {
	w := 200
	out := strings.TrimRight(render.Render("a\nb  \nc",
		types.Terminal{Width: w}, types.RenderOptions{}), "\n")
	if !strings.Contains(out, "a b\nc") {
		t.Errorf("expected 'a b\\nc', got %q", out)
	}
}

// --- Inline HTML ---

func TestRender_InlineHTML_Ignored(t *testing.T) {
	out := strings.TrimRight(renderPlain("<div>ignored</div>"), "\n")
	if out != "" {
		t.Errorf("inline HTML should be ignored, got %q", out)
	}
}

// --- Hyperlinks ---

func TestRender_Hyperlinks_URLSuffix_WhenOff(t *testing.T) {
	out := renderPlain("[link](https://example.com)")
	if !strings.Contains(out, "link (https://example.com)") {
		t.Errorf("expected URL suffix when hyperlinks off, got %q", out)
	}
}

func TestRender_Hyperlinks_OSC8_WhenEnabled(t *testing.T) {
	out := render.Render("[x](https://example.com)", types.Terminal{Color: colorprofile.TrueColor, Hyperlinks: true}, types.RenderOptions{
		Wrap: boolFalse(),
	})
	if !strings.HasPrefix(out, "\x1b]8;;https://example.com\x07") || !strings.Contains(out, "\x1b]8;;\x07") {
		t.Errorf("expected OSC-8 hyperlink, got %q", out)
	}
	if !strings.Contains(out, "\x1b[4;34") {
		t.Errorf("expected hyperlink text to carry the theme's link style, got %q", out)
	}
}

func TestRender_Hyperlinks_OSC8_DisabledWhenColorFalse(t *testing.T) {
	out := render.Render("[x](https://example.com)", types.Terminal{Hyperlinks: true}, types.RenderOptions{
		Wrap: boolFalse(),
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
	out := renderPlain("- item 1\n\n- item 2")
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
	w := 200
	out := strings.TrimRight(render.Render(md, types.Terminal{Width: w}, types.RenderOptions{}), "\n")
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
	out := render.Render("`inline`\n\n```\nblock\n```", types.Terminal{Color: colorprofile.TrueColor}, types.RenderOptions{
		Wrap: boolFalse(),
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
	out := render.Render("`x`\n\n```\ny\n```", types.Terminal{Color: colorprofile.TrueColor}, types.RenderOptions{
		Wrap:  boolFalse(),
		Theme: &types.Theme{Code: types.StyleIntent{Color: "red"}},
	})
	if !strings.Contains(out, "\x1b[31m") {
		t.Errorf("expected red from code fallback, got %q", out)
	}
}

func TestRender_DefaultTheme_Colors(t *testing.T) {
	out := render.Render("`inline`\n\n```\nblock\n```\n\n# H", types.Terminal{Color: colorprofile.TrueColor}, types.RenderOptions{
		Wrap:    boolFalse(),
		CodeBox: boolFalse(),
	})
	if !strings.Contains(out, "\x1b[36m") { // cyan inline code
		t.Errorf("expected cyan inline code, got %q", out)
	}
	if !strings.Contains(out, "\x1b[32m") { // green block code
		t.Errorf("expected green block code, got %q", out)
	}
}

func TestRender_HighlighterHook(t *testing.T) {
	out := render.Render("```\ncode\n```", types.Terminal{Color: colorprofile.TrueColor}, types.RenderOptions{
		Wrap:        boolFalse(),
		Highlighter: func(code, _ string) string { return strings.ToUpper(code) },
	})
	if !strings.Contains(out, "CODE") {
		t.Errorf("expected highlighter to uppercase code, got %q", out)
	}
}

// --- SquashParagraphs ---

func TestRender_ParagraphSpacing_BlankLineByDefault(t *testing.T) {
	out := renderPlain("first paragraph\n\nsecond paragraph\n")
	lines := strings.Split(out, "\n")
	hasBlank := false
	for _, l := range lines {
		if l == "" {
			hasBlank = true
			break
		}
	}
	if !hasBlank {
		t.Errorf("expected blank line between paragraphs by default, got: %q", out)
	}
}

func TestRender_ParagraphSpacing_SquashRemovesBlankLine(t *testing.T) {
	out := render.Render("first paragraph\n\nsecond paragraph\n", types.Terminal{}, types.RenderOptions{
		SquashParagraphs: true,
	})
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	for _, l := range lines {
		if l == "" {
			t.Errorf("expected no blank lines when SquashParagraphs=true, got: %q", out)
			return
		}
	}
}

// --- QuotePrefix default ---

func TestRender_BlockquoteDefaultPrefix(t *testing.T) {
	out := renderPlain("> quoted\n")
	if !strings.HasPrefix(strings.TrimLeft(out, "\n"), "> ") {
		t.Errorf("expected default blockquote prefix '> ', got: %q", out)
	}
}

func TestRender_BlockquoteCustomPrefix(t *testing.T) {
	p := "│ "
	out := render.Render("> quoted\n", types.Terminal{}, types.RenderOptions{
		QuotePrefix: &p,
	})
	if !strings.Contains(out, "│ quoted") {
		t.Errorf("expected custom blockquote prefix '│ ', got: %q", out)
	}
}

func boolTrue() *bool  { t := true; return &t }
func boolFalse() *bool { f := false; return &f }

// --- Colour profiles ---

func TestRender_ColorProfileDownsampling(t *testing.T) {
	// The solarized theme renders code blocks in #2aa198.
	cases := []struct {
		profile colorprofile.Profile
		want    string
		notWant string
	}{
		{colorprofile.TrueColor, "38;2;42;161;152", ""},
		{colorprofile.ANSI256, "38;5;", "38;2;"},
		{colorprofile.ANSI, "\x1b[32m", "38;"},
		{colorprofile.ASCII, "", "\x1b[3"},
	}
	for _, c := range cases {
		out := render.Render("```\nx\n```", types.Terminal{Color: c.profile}, types.RenderOptions{
			ThemeName: types.ThemeSolarized,
		})
		if !strings.Contains(out, c.want) {
			t.Errorf("%v: expected %q in %q", c.profile, c.want, out)
		}
		if c.notWant != "" && strings.Contains(out, c.notWant) {
			t.Errorf("%v: did not expect %q in %q", c.profile, c.notWant, out)
		}
	}
}

func TestRender_ASCIIProfileKeepsAttributes(t *testing.T) {
	out := render.Render("**bold**", types.Terminal{Color: colorprofile.ASCII}, types.RenderOptions{})
	if !strings.Contains(out, "\x1b[1m") {
		t.Errorf("expected bold to survive the ASCII profile, got %q", out)
	}
}

func TestRender_WrappedStyleDoesNotLeakIntoQuotePrefix(t *testing.T) {
	out := render.Render("> x **bold text that wraps across lines**", types.Terminal{Width: 20, Color: colorprofile.TrueColor}, types.RenderOptions{})
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected wrapping, got %q", out)
	}
	for i, l := range lines[:len(lines)-1] {
		if !strings.HasSuffix(l, "\x1b[m") {
			t.Errorf("line %d leaves bold open into the next prefix: %q", i, l)
		}
	}
}
