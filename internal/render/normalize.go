package render

import (
	"math"
	"regexp"
	"strings"

	"github.com/yuin/goldmark/ast"
)

// Block is a normalized, renderer-ready top-level block. For most node types
// Node carries the original goldmark AST node and the renderer reads content
// from it directly. For merged or synthesized code blocks, Content and Lang
// are pre-extracted strings and Node may be the first constituent block.
type Block struct {
	Node    ast.Node // original goldmark AST node
	Kind    ast.NodeKind
	Lang    string // effective language for code blocks (may differ from node's info)
	Content string // pre-extracted content; non-empty only for merged/synthetic code blocks
}

// isCode reports whether a Block is a fenced or indented code block.
func isCode(b *Block) bool {
	return b.Kind == ast.KindFencedCodeBlock || b.Kind == ast.KindCodeBlock
}

// Normalize walks the top-level children of the goldmark document, applies
// normalizations equivalent to those in src/render.ts, and returns a flat
// slice of Blocks ready for the renderer.
//
// Normalizations applied (in order):
//  1. Reference-definition paragraphs followed by no-lang code blocks → merge
//  2. Lone indented code blocks that look like reference definitions → paragraph
//  3. Lists whose every item is a single code block → flatten to code block
//  4. Adjacent same-lang code blocks → merge into one
//  5. Unlabelled code blocks that look like diffs → tag lang="diff"
func Normalize(root ast.Node, src []byte) []*Block {
	// Collect top-level children into a raw Block slice.
	raw := collectTopLevel(root, src)

	// Pass 1: merge reference-definition paragraphs with following code blocks,
	// and rewrite lone definition-like code blocks as paragraphs.
	pass1 := normalizeDefinitions(raw, src)

	// Pass 2: merge adjacent same-lang code blocks (and flatten lists-of-code).
	pass2 := mergeAdjacentCode(pass1, src)

	// Pass 3: tag unlabelled code blocks that look like unified diffs.
	pass3 := tagDiffBlocks(pass2)

	return pass3
}

// collectTopLevel converts the document's direct children into Block values.
func collectTopLevel(root ast.Node, src []byte) []*Block {
	var blocks []*Block
	for n := root.FirstChild(); n != nil; n = n.NextSibling() {
		lang := ""
		if n.Kind() == ast.KindFencedCodeBlock || n.Kind() == ast.KindCodeBlock {
			lang = CodeBlockLang(n, src)
		}
		blocks = append(blocks, &Block{
			Node: n,
			Kind: n.Kind(),
			Lang: lang,
		})
	}
	return blocks
}

// --- Pass 1: reference-definition normalisation ---

// reDef matches a paragraph whose text looks like a reference-style link
// definition that may be missing its closing quote:
// [label]: url "title (possibly missing closing quote)
var reDef = regexp.MustCompile(`(?m)^\[[\w\d]+\]:\s+\S+.*"\s*$`)

// reDefStandalone matches an indented code block that is itself a reference def.
var reDefStandalone = regexp.MustCompile(`^\[[\w\d]+\]:\s+\S+`)

func normalizeDefinitions(blocks []*Block, src []byte) []*Block {
	out := make([]*Block, 0, len(blocks))
	for i := 0; i < len(blocks); i++ {
		b := blocks[i]

		// Paragraph followed by a no-lang code block that looks like a
		// continuation of a reference definition.
		if b.Kind == ast.KindParagraph {
			text := extractNodeText(b.Node, src)
			if reDef.MatchString(text) && i+1 < len(blocks) {
				next := blocks[i+1]
				if isCode(next) && next.Lang == "" {
					continuation := reIndent.ReplaceAllString(CodeBlockContent(next.Node, src), " ")
					continuation = strings.Join(strings.Fields(continuation), " ")
					merged := strings.TrimSuffix(strings.TrimRight(text, " \t"), `"`) + `" ` + strings.Trim(continuation, `"`)
					merged = strings.TrimSpace(merged)
					out = append(out, &Block{
						Node:    b.Node,
						Kind:    ast.KindParagraph,
						Content: merged,
					})
					i++ // skip the consumed code block
					continue
				}
			}
		}

		// Lone indented code block that looks like a reference definition.
		if isCode(b) && b.Lang == "" {
			content := CodeBlockContent(b.Node, src)
			stripped := reIndent.ReplaceAllString(content, "")
			stripped = strings.TrimSpace(stripped)
			if reDefStandalone.MatchString(stripped) {
				out = append(out, &Block{
					Node:    b.Node,
					Kind:    ast.KindParagraph,
					Content: stripped,
				})
				continue
			}
		}

		out = append(out, b)
	}
	return out
}

var reIndent = regexp.MustCompile(`(?m)^[ \t>]+`)

// extractNodeText walks inline children of a node and concatenates their text
// segments. Used to read the text content of a paragraph for pattern matching.
func extractNodeText(n ast.Node, src []byte) string {
	var sb strings.Builder
	_ = ast.Walk(n, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if node.Kind() == ast.KindText {
			t := node.(*ast.Text)
			sb.Write(t.Segment.Value(src))
			if t.HardLineBreak() || t.SoftLineBreak() {
				sb.WriteByte('\n')
			}
		}
		return ast.WalkContinue, nil
	})
	return sb.String()
}

// --- Pass 2: merge adjacent same-lang code blocks ---

func mergeAdjacentCode(blocks []*Block, src []byte) []*Block {
	out := make([]*Block, 0, len(blocks))
	var pending *Block

	flush := func() {
		if pending != nil {
			out = append(out, pending)
			pending = nil
		}
	}

	for _, b := range blocks {
		// A list whose every item is a single code block collapses to one code block.
		if b.Kind == ast.KindList {
			if flat := flattenListOfCode(b.Node, src); flat != nil {
				b = flat
			}
		}

		if isCode(b) {
			if pending != nil && pending.Lang == b.Lang {
				// Merge: append content with a blank line separator.
				pendingContent := blockContent(pending, src)
				nextContent := blockContent(b, src)
				pending = &Block{
					Node:    pending.Node,
					Kind:    pending.Kind,
					Lang:    pending.Lang,
					Content: pendingContent + "\n" + nextContent,
				}
				continue
			}
			flush()
			pending = b
			// Ensure Content is populated so the renderer doesn't need to re-extract.
			if pending.Content == "" {
				pending.Content = CodeBlockContent(pending.Node, src)
			}
			continue
		}

		flush()
		out = append(out, b)
	}
	flush()
	return out
}

// blockContent returns the pre-extracted content of a Block, falling back to
// reading it from the goldmark AST if not already stored.
func blockContent(b *Block, src []byte) string {
	if b.Content != "" {
		return b.Content
	}
	return CodeBlockContent(b.Node, src)
}

// flattenListOfCode returns a synthetic code Block if every list item in n
// contains exactly one fenced/indented code block, otherwise nil.
// This mirrors the TS flattenCodeList function.
func flattenListOfCode(n ast.Node, src []byte) *Block {
	var codes []*Block
	for item := n.FirstChild(); item != nil; item = item.NextSibling() {
		if item.Kind() != ast.KindListItem {
			return nil
		}
		child := item.FirstChild()
		if child == nil || child.NextSibling() != nil {
			return nil // item has 0 or 2+ children
		}
		if child.Kind() != ast.KindFencedCodeBlock && child.Kind() != ast.KindCodeBlock {
			return nil
		}
		codes = append(codes, &Block{
			Node: child,
			Kind: child.Kind(),
			Lang: CodeBlockLang(child, src),
		})
	}
	if len(codes) == 0 {
		return nil
	}
	// All items must share the same language.
	lang := codes[0].Lang
	for _, c := range codes[1:] {
		if c.Lang != lang {
			lang = ""
			break
		}
	}
	var parts []string
	for _, c := range codes {
		parts = append(parts, CodeBlockContent(c.Node, src))
	}
	return &Block{
		Node:    n,
		Kind:    ast.KindFencedCodeBlock,
		Lang:    lang,
		Content: strings.Join(parts, "\n"),
	}
}

// --- Pass 3: tag diff blocks ---

var (
	reDiffHeader = regexp.MustCompile(`(?m)^(diff --git|--- a/|\+\+\+ b/|@@ )`)
	reDiffMarker = regexp.MustCompile(`(?m)^[+\-@]`)
)

// looksLikeDiff reports whether text looks like a unified diff.
// Mirrors the TS looksLikeDiff function.
func looksLikeDiff(text string) bool {
	lines := strings.Split(text, "\n")
	// Trim each line for comparison.
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if strings.HasPrefix(trimmed, "diff --git") ||
			strings.HasPrefix(trimmed, "--- a/") ||
			strings.HasPrefix(trimmed, "+++ b/") ||
			strings.HasPrefix(trimmed, "@@ ") {
			return true
		}
	}
	var nonEmpty []string
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			nonEmpty = append(nonEmpty, strings.TrimSpace(l))
		}
	}
	if len(nonEmpty) < 3 {
		return false
	}
	markers := 0
	for _, l := range nonEmpty {
		if reDiffMarker.MatchString(l) {
			markers++
		}
	}
	threshold := math.Max(3, math.Ceil(float64(len(nonEmpty))*0.6))
	return float64(markers) >= threshold
}

func tagDiffBlocks(blocks []*Block) []*Block {
	for _, b := range blocks {
		if isCode(b) && b.Lang == "" && looksLikeDiff(b.Content) {
			b.Lang = "diff"
		}
	}
	return blocks
}
