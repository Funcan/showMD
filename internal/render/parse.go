// Package render provides the Markdown → ANSI renderer.
// parse.go sets up the goldmark parser with GFM extensions.
package render

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// gm is the shared goldmark instance configured with all GFM extensions:
// Linkify, Table, Strikethrough, TaskList.
var gm = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithParserOptions(
		parser.WithAutoHeadingID(),
	),
)

// ParsedDoc holds a parsed goldmark AST together with the source bytes needed
// to read node text (goldmark stores text as byte-range references into source).
// The Blocks slice is the normalized, renderer-ready sequence of top-level blocks.
type ParsedDoc struct {
	Source []byte
	Root   ast.Node
	Blocks []*Block
}

// Parse parses markdown source, applies normalizations, and returns a ParsedDoc.
func Parse(markdown string) *ParsedDoc {
	src := []byte(markdown)
	reader := text.NewReader(src)
	root := gm.Parser().Parse(reader)
	blocks := Normalize(root, src)
	return &ParsedDoc{
		Source: src,
		Root:   root,
		Blocks: blocks,
	}
}

// CodeBlockContent extracts the text content of a fenced or indented code block
// node as a string, stripping the trailing newline that goldmark appends to the
// last line.
func CodeBlockContent(n ast.Node, src []byte) string {
	lines := n.Lines()
	if lines == nil {
		return ""
	}
	var out []byte
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		out = append(out, seg.Value(src)...)
	}
	// goldmark appends a trailing newline; strip it so content matches the TS
	// behaviour where code.value has no trailing newline.
	if len(out) > 0 && out[len(out)-1] == '\n' {
		out = out[:len(out)-1]
	}
	return string(out)
}

// CodeBlockLang returns the language string for a FencedCodeBlock, or "" for
// indented code blocks or blocks without an info string.
func CodeBlockLang(n ast.Node, src []byte) string {
	fcb, ok := n.(*ast.FencedCodeBlock)
	if !ok {
		return ""
	}
	lang := fcb.Language(src)
	if lang == nil {
		return ""
	}
	return string(lang)
}
