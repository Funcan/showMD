package stream_test

import (
	"strings"
	"testing"

	"github.com/funcan/showmd/internal/render"
	"github.com/funcan/showmd/internal/stream"
	"github.com/funcan/showmd/internal/types"
)

// renderNoColor mirrors the TS renderNoColor helper used in the stream tests.
func renderNoColor(md string) string {
	f := false
	w := 60
	return render.Render(md, types.RenderOptions{
		Width:      &w,
		Wrap:       boolPtr(true),
		Color:      &f,
		Hyperlinks: &f,
	})
}

func boolPtr(b bool) *bool { return &b }

func newStreamer(spacing stream.SpacingMode) *stream.Streamer {
	return stream.New(stream.Options{
		Render:  renderNoColor,
		Spacing: spacing,
	})
}

func TestStream_BuffersUntilNewlineForRegularLines(t *testing.T) {
	s := newStreamer(stream.SpacingPreserve)
	if got := s.Push("Hello"); got != "" {
		t.Errorf("expected empty before newline, got %q", got)
	}
	got := s.Push(" world\n")
	if !strings.Contains(got, "Hello world") {
		t.Errorf("expected 'Hello world' in output, got %q", got)
	}
}

func TestStream_BuffersFencedBlockUntilClosingFence(t *testing.T) {
	s := newStreamer(stream.SpacingPreserve)
	if got := s.Push("```txt\n"); got != "" {
		t.Errorf("expected empty on fence open, got %q", got)
	}
	if got := s.Push("line 1\n"); got != "" {
		t.Errorf("expected empty inside fence, got %q", got)
	}
	got := s.Push("```\n")
	if got == "" {
		t.Error("expected output on fence close, got empty")
	}
}

func TestStream_BuffersTableUntilNonTableLine(t *testing.T) {
	s := newStreamer(stream.SpacingPreserve)
	if got := s.Push("| A | B |\n"); got != "" {
		t.Errorf("expected empty on header, got %q", got)
	}
	if got := s.Push("|---|---|\n"); got != "" {
		t.Errorf("expected empty on separator, got %q", got)
	}
	if got := s.Push("| 1 | 2 |\n"); got != "" {
		t.Errorf("expected empty on body row, got %q", got)
	}
	out := s.Push("\n")
	if !strings.Contains(out, "A") || !strings.Contains(out, "B") {
		t.Errorf("expected table content after blank line, got %q", out)
	}
}

func TestStream_TrimsLeadingNewlineFromHeadingFragment(t *testing.T) {
	s := newStreamer(stream.SpacingPreserve)
	out := s.Push("## Heading\n")
	if strings.HasPrefix(out, "\n") {
		t.Errorf("output should not start with newline, got %q", out)
	}
	if !strings.Contains(out, "Heading\n") {
		t.Errorf("expected 'Heading' in output, got %q", out)
	}
}

func TestStream_NoExtraBlankLinesAroundHeadings(t *testing.T) {
	s := newStreamer(stream.SpacingSingle)
	out := s.Push("## Overview\n\n- One\n\n## Key Evidence\n- Two\n") + s.Finish("")
	if strings.Contains(out, "\n\n\n") {
		t.Errorf("should not have triple newline, got %q", out)
	}
}

func TestStream_CollapsesConsecutiveBlankLinesInSingleMode(t *testing.T) {
	s := newStreamer(stream.SpacingSingle)
	out := s.Push("A\n\n\nB\n") + s.Finish("")
	if !strings.Contains(out, "A\n\nB\n") {
		t.Errorf("expected collapsed blank lines, got %q", out)
	}
	if strings.Contains(out, "A\n\n\nB\n") {
		t.Errorf("should not have triple newline, got %q", out)
	}
}

func TestStream_DropsLeadingBlankLinesBeforeFirstContent(t *testing.T) {
	s := newStreamer(stream.SpacingSingle)
	out := s.Push("\n\nA\n") + s.Finish("")
	if strings.HasPrefix(out, "\n") {
		t.Errorf("output should not start with newline, got %q", out)
	}
	if !strings.Contains(out, "A\n") {
		t.Errorf("expected 'A' in output, got %q", out)
	}
}

func TestStream_PreservesConsecutiveBlankLinesInPreserveMode(t *testing.T) {
	s := newStreamer(stream.SpacingPreserve)
	out := s.Push("A\n\n\nB\n") + s.Finish("")
	if !strings.Contains(out, "A\n\n\nB\n") {
		t.Errorf("expected three newlines preserved, got %q", out)
	}
}

func TestStream_FlushesHeldTableHeaderWhenNextLineNotSeparator(t *testing.T) {
	s := newStreamer(stream.SpacingPreserve)
	if got := s.Push("| A |\n"); got != "" {
		t.Errorf("expected empty on potential header, got %q", got)
	}
	out := s.Push("Not a table\n")
	idxA := strings.Index(out, "A")
	idxNot := strings.Index(out, "Not a table")
	if idxA < 0 {
		t.Errorf("expected 'A' in output, got %q", out)
	}
	if idxNot < 0 {
		t.Errorf("expected 'Not a table' in output, got %q", out)
	}
	if idxNot <= idxA {
		t.Errorf("'A' should appear before 'Not a table', got %q", out)
	}
}

func TestStream_FlushesTableOnNonRowLineAndContinuesRendering(t *testing.T) {
	s := newStreamer(stream.SpacingPreserve)
	s.Push("| A | B |\n")
	s.Push("|---|---|\n")
	s.Push("| 1 | 2 |\n")
	out := s.Push("Next\n")
	if !strings.Contains(out, "A") {
		t.Errorf("expected table content, got %q", out)
	}
	if !strings.Contains(out, "Next") {
		t.Errorf("expected 'Next' in output, got %q", out)
	}
}

func TestStream_FlushesUnterminatedFencedBlockOnFinish(t *testing.T) {
	s := newStreamer(stream.SpacingPreserve)
	s.Push("```txt\n")
	s.Push("code\n")
	out := s.Finish("")
	if !strings.Contains(out, "code") {
		t.Errorf("expected 'code' in flushed output, got %q", out)
	}
}

func TestStream_FlushesFinalBufferedLineOnFinish(t *testing.T) {
	s := newStreamer(stream.SpacingPreserve)
	if got := s.Push("Hello"); got != "" {
		t.Errorf("expected empty before newline, got %q", got)
	}
	out := s.Finish("")
	if !strings.Contains(out, "Hello\n") {
		t.Errorf("expected 'Hello' in finish output, got %q", out)
	}
}

func TestStream_ResetClearsBufferedState(t *testing.T) {
	s := newStreamer(stream.SpacingPreserve)
	s.Push("| A | B |\n")
	s.Reset()
	out := s.Push("Hi\n") + s.Finish("")
	if !strings.Contains(out, "Hi\n") {
		t.Errorf("expected 'Hi' after reset, got %q", out)
	}
	if strings.Contains(out, "A") {
		t.Errorf("should not contain pre-reset state 'A', got %q", out)
	}
}

func TestStream_NormalizesCRLFNewlines(t *testing.T) {
	s := newStreamer(stream.SpacingPreserve)
	out := s.Push("A\r\nB\r\n") + s.Finish("")
	if !strings.Contains(out, "A\n") {
		t.Errorf("expected 'A' with LF, got %q", out)
	}
	if !strings.Contains(out, "B\n") {
		t.Errorf("expected 'B' with LF, got %q", out)
	}
	if strings.Contains(out, "\r") {
		t.Errorf("should not contain CR, got %q", out)
	}
}

func TestStream_TightModeDropsAllBlankLines(t *testing.T) {
	s := newStreamer(stream.SpacingTight)
	out := s.Push("A\n\n\nB\n") + s.Finish("")
	if strings.Contains(out, "\n\n") {
		t.Errorf("tight mode should drop blank lines, got %q", out)
	}
	if !strings.Contains(out, "A") || !strings.Contains(out, "B") {
		t.Errorf("expected content in tight output, got %q", out)
	}
}
