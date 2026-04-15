package render_test

import (
	"strings"
	"testing"

	"github.com/funcan/showmd/internal/render"
	"github.com/yuin/goldmark/ast"
)

// helper: parse and return blocks
func blocks(markdown string) []*render.Block {
	return render.Parse(markdown).Blocks
}

func firstCode(t *testing.T, bks []*render.Block) *render.Block {
	t.Helper()
	for _, b := range bks {
		if b.Kind == ast.KindFencedCodeBlock || b.Kind == ast.KindCodeBlock {
			return b
		}
	}
	t.Fatal("no code block found")
	return nil
}

// --- CodeBlockContent ---

func TestCodeBlockContent_FencedBlock(t *testing.T) {
	doc := render.Parse("```go\nhello\nworld\n```\n")
	cb := firstCode(t, doc.Blocks)
	if !strings.Contains(cb.Content, "hello") {
		t.Errorf("code block content should contain 'hello', got %q", cb.Content)
	}
	if strings.HasSuffix(cb.Content, "\n") {
		t.Errorf("code block content should not have trailing newline, got %q", cb.Content)
	}
}

func TestCodeBlockContent_IndentedBlock(t *testing.T) {
	doc := render.Parse("    line one\n    line two\n")
	cb := firstCode(t, doc.Blocks)
	if !strings.Contains(cb.Content, "line one") {
		t.Errorf("indented code block content: want 'line one', got %q", cb.Content)
	}
}

// --- CodeBlockLang ---

func TestCodeBlockLang_WithLang(t *testing.T) {
	doc := render.Parse("```typescript\nconst x = 1\n```\n")
	cb := firstCode(t, doc.Blocks)
	if cb.Lang != "typescript" {
		t.Errorf("code block lang: want 'typescript', got %q", cb.Lang)
	}
}

func TestCodeBlockLang_NoLang(t *testing.T) {
	doc := render.Parse("```\nno lang\n```\n")
	cb := firstCode(t, doc.Blocks)
	if cb.Lang != "" {
		t.Errorf("code block without lang: want '', got %q", cb.Lang)
	}
}

// --- Normalize: merge adjacent code blocks ---

func TestNormalize_MergesAdjacentSameLangBlocks(t *testing.T) {
	md := "```go\nfunc A() {}\n```\n\n```go\nfunc B() {}\n```\n"
	bks := blocks(md)

	codeBlocks := 0
	for _, b := range bks {
		if b.Kind == ast.KindFencedCodeBlock || b.Kind == ast.KindCodeBlock {
			codeBlocks++
		}
	}
	if codeBlocks != 1 {
		t.Errorf("adjacent same-lang blocks should merge to 1, got %d", codeBlocks)
	}
	cb := firstCode(t, bks)
	if !strings.Contains(cb.Content, "func A()") || !strings.Contains(cb.Content, "func B()") {
		t.Errorf("merged block should contain both functions, got %q", cb.Content)
	}
}

func TestNormalize_DoesNotMergeDifferentLangBlocks(t *testing.T) {
	md := "```go\nfunc A() {}\n```\n\n```python\ndef b(): pass\n```\n"
	bks := blocks(md)

	codeBlocks := 0
	for _, b := range bks {
		if b.Kind == ast.KindFencedCodeBlock || b.Kind == ast.KindCodeBlock {
			codeBlocks++
		}
	}
	if codeBlocks != 2 {
		t.Errorf("different-lang blocks should stay separate, got %d", codeBlocks)
	}
}

func TestNormalize_DoesNotMergeWhenSeparatedByParagraph(t *testing.T) {
	md := "```go\nfunc A() {}\n```\n\nsome text\n\n```go\nfunc B() {}\n```\n"
	bks := blocks(md)

	codeBlocks := 0
	for _, b := range bks {
		if b.Kind == ast.KindFencedCodeBlock || b.Kind == ast.KindCodeBlock {
			codeBlocks++
		}
	}
	if codeBlocks != 2 {
		t.Errorf("blocks separated by paragraph should not merge, got %d", codeBlocks)
	}
}

// --- Normalize: diff detection ---

func TestNormalize_TagsDiffBlock(t *testing.T) {
	diff := "```\ndiff --git a/foo.go b/foo.go\n--- a/foo.go\n+++ b/foo.go\n@@ -1,3 +1,3 @@\n-old\n+new\n```\n"
	bks := blocks(diff)
	cb := firstCode(t, bks)
	if cb.Lang != "diff" {
		t.Errorf("unlabelled diff block should be tagged 'diff', got %q", cb.Lang)
	}
}

func TestNormalize_DoesNotTagNormalCodeAsDiff(t *testing.T) {
	md := "```\nhello world\nfoo bar\n```\n"
	bks := blocks(md)
	cb := firstCode(t, bks)
	if cb.Lang == "diff" {
		t.Errorf("plain code should not be tagged as diff")
	}
}

func TestNormalize_DiffByMarkerDensity(t *testing.T) {
	// 4 out of 5 non-empty lines start with + or -, threshold 60%
	md := "```\n+added line\n+another add\n-removed\n-gone too\nunchanged\n```\n"
	bks := blocks(md)
	cb := firstCode(t, bks)
	if cb.Lang != "diff" {
		t.Errorf("high-density marker block should be tagged as diff, got lang=%q", cb.Lang)
	}
}

func TestNormalize_ExplicitLangNotOverriddenByDiffDetection(t *testing.T) {
	md := "```python\n+added\n-removed\n+more\n-gone\n```\n"
	bks := blocks(md)
	cb := firstCode(t, bks)
	if cb.Lang != "python" {
		t.Errorf("explicit lang should not be overridden by diff detection, got %q", cb.Lang)
	}
}

// --- Parse: document structure ---

func TestParse_HeadingAndParagraph(t *testing.T) {
	bks := blocks("# Hello\n\nA paragraph.\n")
	if len(bks) < 2 {
		t.Fatalf("expected at least 2 blocks, got %d", len(bks))
	}
	if bks[0].Kind != ast.KindHeading {
		t.Errorf("first block should be heading, got %v", bks[0].Kind)
	}
	if bks[1].Kind != ast.KindParagraph {
		t.Errorf("second block should be paragraph, got %v", bks[1].Kind)
	}
}

func TestParse_EmptyInput(t *testing.T) {
	bks := blocks("")
	if len(bks) != 0 {
		t.Errorf("empty input should produce no blocks, got %d", len(bks))
	}
}

func TestParse_ThematicBreak(t *testing.T) {
	bks := blocks("---\n")
	if len(bks) != 1 || bks[0].Kind != ast.KindThematicBreak {
		t.Errorf("'---' should parse as ThematicBreak, got %v", bks)
	}
}

func TestParse_Blockquote(t *testing.T) {
	bks := blocks("> quoted\n")
	if len(bks) != 1 || bks[0].Kind != ast.KindBlockquote {
		t.Errorf("blockquote should parse as KindBlockquote, got %v", bks)
	}
}

func TestParse_List(t *testing.T) {
	bks := blocks("- item one\n- item two\n")
	if len(bks) != 1 || bks[0].Kind != ast.KindList {
		t.Errorf("list should parse as KindList, got %v", bks)
	}
}
