// Package stream provides a stateful streaming Markdown renderer.
// It processes Markdown arriving in chunks (e.g. from an LLM stream), buffering
// incomplete fenced code blocks and tables so they render atomically.
//
// This is a direct port of src/stream.ts from Markdansi.
package stream

import (
	"regexp"
	"strings"
)

// SpacingMode controls how blank lines are emitted.
type SpacingMode string

const (
	// SpacingPreserve emits blank lines exactly as received.
	SpacingPreserve SpacingMode = "preserve"
	// SpacingSingle collapses consecutive blank lines to a single blank line.
	SpacingSingle SpacingMode = "single"
	// SpacingTight drops all blank lines (dense output).
	SpacingTight SpacingMode = "tight"
)

// fenceState tracks an open fenced code block.
type fenceState struct {
	char byte // '`' or '~'
	len  int
}

// Options configures a Streamer.
type Options struct {
	// Render converts a Markdown fragment to ANSI text. Must be pure (no cursor
	// control) and must not rely on prior terminal state.
	Render func(markdown string) string
	// Spacing controls blank-line emission. Defaults to SpacingSingle.
	Spacing SpacingMode
}

// Streamer holds the state of a hybrid streaming Markdown renderer.
// Create one with New; do not copy after first use.
type Streamer struct {
	renderFn func(string) string
	spacing  SpacingMode

	buffer      string
	blankStreak int
	started     bool

	heldTableHeader *string
	inTable         bool
	tableBuffer     string

	fence       *fenceState
	fenceBuffer string
}

// New creates a Streamer with the given options.
func New(opts Options) *Streamer {
	spacing := opts.Spacing
	if spacing == "" {
		spacing = SpacingSingle
	}
	return &Streamer{
		renderFn: opts.Render,
		spacing:  spacing,
	}
}

// Push appends a Markdown delta (chunk) and returns any ANSI output ready to write.
// The caller should append all returned strings to the terminal output in order.
func (s *Streamer) Push(delta string) string {
	if delta == "" {
		return ""
	}
	s.buffer += normalizeNewlines(delta)
	var out strings.Builder
	for {
		idx := strings.IndexByte(s.buffer, '\n')
		if idx < 0 {
			break
		}
		line := s.buffer[:idx]
		s.buffer = s.buffer[idx+1:]
		out.WriteString(s.processLine(line))
	}
	return out.String()
}

// Finish flushes any remaining buffered content and finalizes the stream.
// An optional finalDelta may be provided; it is pushed before flushing.
func (s *Streamer) Finish(finalDelta string) string {
	var out strings.Builder
	if finalDelta != "" {
		out.WriteString(s.Push(finalDelta))
	}
	if s.buffer != "" {
		out.WriteString(s.processLine(s.buffer))
		s.buffer = ""
	}
	out.WriteString(s.flushHeldHeader())
	out.WriteString(s.flushFence())
	out.WriteString(s.flushTable())
	return out.String()
}

// Reset zeros all internal state so the Streamer can be reused.
func (s *Streamer) Reset() {
	s.buffer = ""
	s.blankStreak = 0
	s.started = false
	s.heldTableHeader = nil
	s.inTable = false
	s.tableBuffer = ""
	s.fence = nil
	s.fenceBuffer = ""
}

// --- internal helpers ---

var newlineReplacer = strings.NewReplacer("\r\n", "\n", "\r", "\n")

func normalizeNewlines(s string) string {
	return newlineReplacer.Replace(s)
}

var reFenceStart = regexp.MustCompile("^(```+|~~~+)")

func parseFenceStart(line string) *fenceState {
	trimmed := strings.TrimLeft(line, " \t")
	m := reFenceStart.FindString(trimmed)
	if m == "" {
		return nil
	}
	ch := m[0]
	if ch != '`' {
		ch = '~'
	}
	return &fenceState{char: ch, len: len(m)}
}

func isFenceEnd(line string, f *fenceState) bool {
	trimmed := strings.TrimLeft(line, " \t")
	token := strings.Repeat(string(f.char), f.len)
	return strings.HasPrefix(trimmed, token)
}

func looksLikeTableHeader(line string) bool {
	if !strings.Contains(line, "|") {
		return false
	}
	for _, r := range line {
		if r != ' ' && r != '|' && r != '\t' {
			return true
		}
	}
	return false
}

var reTableSep = regexp.MustCompile(`^\|?(?:\s*:?-+:?\s*\|)+\s*:?-+:?\s*\|?$`)

func isTableSeparator(line string) bool {
	trimmed := strings.TrimSpace(line)
	if !strings.Contains(trimmed, "-") {
		return false
	}
	return reTableSep.MatchString(trimmed)
}

func looksLikeTableRow(line string) bool {
	return looksLikeTableHeader(line)
}

// normalizeRenderedFragment strips leading newlines (headings prepend one) and
// normalizes to exactly one trailing newline, so the streamer owns spacing.
func normalizeRenderedFragment(rendered string) string {
	s := strings.TrimLeft(rendered, "\n")
	s = strings.TrimRight(s, "\n")
	return s + "\n"
}

func (s *Streamer) emitBlankLine() string {
	if !s.started {
		return ""
	}
	if s.spacing == SpacingTight {
		return ""
	}
	if s.spacing == SpacingSingle && s.blankStreak >= 1 {
		return ""
	}
	s.blankStreak++
	return "\n"
}

func (s *Streamer) emitRendered(markdown string) string {
	if markdown == "" {
		return ""
	}
	s.blankStreak = 0
	s.started = true
	return normalizeRenderedFragment(s.renderFn(markdown))
}

func (s *Streamer) flushHeldHeader() string {
	if s.heldTableHeader == nil {
		return ""
	}
	md := *s.heldTableHeader
	s.heldTableHeader = nil
	return s.emitRendered(md)
}

func (s *Streamer) flushTable() string {
	if !s.inTable {
		return ""
	}
	s.inTable = false
	md := s.tableBuffer
	s.tableBuffer = ""
	return s.emitRendered(md)
}

func (s *Streamer) flushFence() string {
	if s.fence == nil {
		return ""
	}
	s.fence = nil
	md := s.fenceBuffer
	s.fenceBuffer = ""
	return s.emitRendered(md)
}

var reFenceOnly = regexp.MustCompile(`^(` + "```+" + `|~~~+)\s*$`)

func (s *Streamer) processLine(line string) string {
	// Fence mode: buffer everything until the closing fence.
	if s.fence != nil {
		s.fenceBuffer += line + "\n"
		if isFenceEnd(line, s.fence) {
			return s.flushFence()
		}
		return ""
	}

	// Table mode: buffer table rows; flush when the table ends.
	if s.inTable {
		if strings.TrimSpace(line) == "" {
			return s.flushTable() + s.emitBlankLine()
		}
		if !looksLikeTableRow(line) {
			return s.flushTable() + s.processLine(line)
		}
		s.tableBuffer += line + "\n"
		return ""
	}

	// Blank line: flush any held header and emit spacing.
	if strings.TrimSpace(line) == "" {
		return s.flushHeldHeader() + s.emitBlankLine()
	}

	// Fence start: flush held header and enter fence mode.
	if fs := parseFenceStart(line); fs != nil {
		out := s.flushHeldHeader()
		s.fence = fs
		s.fenceBuffer = line + "\n"
		// Handle a bare closing-fence-only line (e.g. "```" with no lang on open).
		trimmed := strings.TrimLeft(line, " \t")
		if isFenceEnd(line, fs) && reFenceOnly.MatchString(trimmed) {
			return out + s.flushFence()
		}
		return out
	}

	// We held a possible table header; check if this line is a separator.
	if s.heldTableHeader != nil {
		if isTableSeparator(line) && looksLikeTableHeader(*s.heldTableHeader) {
			s.inTable = true
			s.tableBuffer = *s.heldTableHeader + "\n" + line + "\n"
			s.heldTableHeader = nil
			return ""
		}
		out := s.flushHeldHeader()
		return out + s.processLine(line)
	}

	// Potential table header: delay emission until we see the next line.
	if looksLikeTableHeader(line) {
		s.heldTableHeader = &line
		return ""
	}

	// Normal line: render immediately.
	return s.emitRendered(line)
}
