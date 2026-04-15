// Package wrap provides ANSI-aware text wrapping utilities.
// It is a direct port of src/wrap.ts from Markdansi.
package wrap

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
)

// ansiEscape matches ANSI CSI sequences (colors, attributes) and OSC sequences
// (e.g. OSC-8 hyperlinks), so we can strip them before measuring visible width.
var ansiEscape = regexp.MustCompile(
	`\x1b(?:` +
		`\[[0-9;?]*[A-Za-z]` + // CSI sequences: ESC [ ... letter
		`|` +
		`\][^\x07\x1b]*(?:\x07|\x1b\\)` + // OSC sequences: ESC ] ... BEL or ST
		`|` +
		`[@-Z\\-_]` + // Fe sequences: ESC + single char
		`)`,
)

// StripANSI removes all ANSI escape sequences from s.
func StripANSI(s string) string {
	return ansiEscape.ReplaceAllString(s, "")
}

// VisibleWidth returns the visible terminal width of s, ignoring any ANSI
// escape codes and accounting for wide (CJK/emoji) characters.
func VisibleWidth(s string) int {
	return runewidth.StringWidth(StripANSI(s))
}

// orphanWords is the set of words that should not be left dangling at the end
// of a line. Matches the TS orphanPhraseTail regex.
var (
	// preposition+article pair that should move together.
	reOrphanPhrase = regexp.MustCompile(`(?i)\b(with|in|on|of|to|for)\s+(a|an|the)$`)
	// single orphan word.
	reOrphanSingle = regexp.MustCompile(`(?i)\b(a|an|the|to|of|with|and|or|in|on|for)$`)
)

// trimTrailingSpace removes trailing whitespace from s (equivalent to the TS
// trimEndSpaces helper).
func trimTrailingSpace(s string) string {
	return strings.TrimRight(s, " \t")
}

// orphanPhraseTail returns the trailing orphan phrase that should be pulled to
// the next line, or "" if there is none. The returned string may be a
// preposition+article pair ("in the") or a single article/preposition.
// It only matches whole words not followed by punctuation.
func orphanPhraseTail(s string) string {
	trimmed := trimTrailingSpace(s)
	if m := reOrphanPhrase.FindString(trimmed); m != "" {
		return m
	}
	if m := reOrphanSingle.FindString(trimmed); m != "" {
		return m
	}
	return ""
}

// WrapText wraps a single (already-inlined) paragraph string into lines no
// wider than width visible columns. It breaks only on whitespace; words longer
// than width overflow rather than being broken. When wrap is false or width ≤ 0
// the original string is returned as a single-element slice.
//
// The orphan-prevention heuristic: if a line would end with a bare article or
// preposition (or a preposition+article pair), those words are pulled to the
// next line together with the overflowing word, matching the TS behaviour.
func WrapText(text string, width int, wrap bool) []string {
	if !wrap || width <= 0 {
		return []string{text}
	}

	// Split on whitespace, keeping the separators (mirrors TS split(/(\s+)/)).
	tokens := splitKeepSep(text)

	var lines []string
	current := ""
	currentWidth := 0

	for _, tok := range tokens {
		w := VisibleWidth(tok)
		isSpace := strings.TrimSpace(tok) == ""

		if current != "" && !isSpace && currentWidth+w > width {
			// This token would overflow — try orphan rescue first.
			nextWord := strings.TrimLeft(tok, " \t")
			currentNoTrail := trimTrailingSpace(current)
			tail := orphanPhraseTail(currentNoTrail)

			if tail != "" && utf8.RuneCountInString(currentNoTrail) > utf8.RuneCountInString(tail) {
				base := trimTrailingSpace(currentNoTrail[:len(currentNoTrail)-len(tail)])
				if base != "" {
					lines = append(lines, base)
					current = tail + " " + nextWord
					currentWidth = VisibleWidth(current)
					continue
				}
			}

			lines = append(lines, currentNoTrail)
			current = nextWord
			currentWidth = VisibleWidth(current)
			continue
		}

		current += tok
		currentWidth = VisibleWidth(current)
	}

	if current != "" {
		lines = append(lines, trimTrailingSpace(current))
	}
	if len(lines) == 0 {
		lines = append(lines, "")
	}
	return lines
}

// splitKeepSep splits text on whitespace runs, keeping the separators as
// separate tokens. This mirrors the TS `text.split(/(\s+)/).filter(w => w.length > 0)`.
func splitKeepSep(text string) []string {
	var tokens []string
	rest := text
	for rest != "" {
		// Find first whitespace run.
		i := strings.IndexAny(rest, " \t\n\r")
		if i < 0 {
			tokens = append(tokens, rest)
			break
		}
		if i > 0 {
			tokens = append(tokens, rest[:i])
		}
		rest = rest[i:]
		// Consume the whitespace run.
		j := 0
		for j < len(rest) && (rest[j] == ' ' || rest[j] == '\t' || rest[j] == '\n' || rest[j] == '\r') {
			j++
		}
		tokens = append(tokens, rest[:j])
		rest = rest[j:]
	}
	return tokens
}

// WrapWithPrefix wraps text and prepends prefix to every output line. When
// wrap is false, the text is split on existing newlines and prefix is prepended
// to each. The effective wrapping width is reduced by the visible width of
// prefix.
func WrapWithPrefix(text string, width int, wrap bool, prefix string) []string {
	if !wrap {
		var out []string
		for _, line := range strings.Split(text, "\n") {
			out = append(out, prefix+line)
		}
		return out
	}
	effectiveWidth := width - VisibleWidth(prefix)
	if effectiveWidth < 1 {
		effectiveWidth = 1
	}
	var out []string
	for _, line := range strings.Split(text, "\n") {
		for _, part := range WrapText(line, effectiveWidth, wrap) {
			out = append(out, prefix+part)
		}
	}
	return out
}
