package wrap_test

import (
	"strings"
	"testing"

	"github.com/funcan/showmd/internal/wrap"
)

// --- StripANSI ---

func TestStripANSI_RemovesColorCodes(t *testing.T) {
	input := "\x1b[31mhello\x1b[0m"
	got := wrap.StripANSI(input)
	if got != "hello" {
		t.Errorf("StripANSI: want %q, got %q", "hello", got)
	}
}

func TestStripANSI_RemovesOSC8Hyperlink(t *testing.T) {
	input := "\x1b]8;;https://example.com\x07click here\x1b]8;;\x07"
	got := wrap.StripANSI(input)
	if got != "click here" {
		t.Errorf("StripANSI OSC-8: want %q, got %q", "click here", got)
	}
}

func TestStripANSI_PlainString_Unchanged(t *testing.T) {
	input := "hello world"
	got := wrap.StripANSI(input)
	if got != input {
		t.Errorf("StripANSI plain: want %q, got %q", input, got)
	}
}

// --- VisibleWidth ---

func TestVisibleWidth_PlainASCII(t *testing.T) {
	if w := wrap.VisibleWidth("hello"); w != 5 {
		t.Errorf("VisibleWidth ascii: want 5, got %d", w)
	}
}

func TestVisibleWidth_IgnoresANSI(t *testing.T) {
	colored := "\x1b[31mhello\x1b[0m"
	if w := wrap.VisibleWidth(colored); w != 5 {
		t.Errorf("VisibleWidth ANSI: want 5, got %d", w)
	}
}

func TestVisibleWidth_WideChars(t *testing.T) {
	// Each CJK character is 2 columns wide.
	if w := wrap.VisibleWidth("日本"); w != 4 {
		t.Errorf("VisibleWidth CJK: want 4, got %d", w)
	}
}

// --- WrapText ---

func TestWrapText_EmptyString(t *testing.T) {
	got := wrap.WrapText("", 5, true)
	if len(got) != 1 || got[0] != "" {
		t.Errorf("WrapText empty: want [\"\"], got %v", got)
	}
}

func TestWrapText_WrapFalse_ReturnsOriginal(t *testing.T) {
	got := wrap.WrapText("one two three", 5, false)
	if len(got) != 1 || got[0] != "one two three" {
		t.Errorf("WrapText wrap=false: want original as single element, got %v", got)
	}
}

func TestWrapText_ZeroWidth_ReturnsOriginal(t *testing.T) {
	got := wrap.WrapText("abc", 0, true)
	if len(got) != 1 || got[0] != "abc" {
		t.Errorf("WrapText width=0: want [\"abc\"], got %v", got)
	}
}

func TestWrapText_WrapsAtWidth(t *testing.T) {
	// "one two" at width=6: "one" fits (3), " two" would make 7 → break.
	got := wrap.WrapText("one two", 6, true)
	for _, line := range got {
		if wrap.VisibleWidth(line) > 6 {
			t.Errorf("WrapText line exceeds width 6: %q (width %d)", line, wrap.VisibleWidth(line))
		}
	}
	if len(got) < 2 {
		t.Errorf("WrapText expected at least 2 lines, got %v", got)
	}
}

func TestWrapText_LongWordOverflows(t *testing.T) {
	url := "https://example.com/averylongpathwithoutspaces"
	got := wrap.WrapText(url, 10, true)
	if len(got) != 1 || got[0] != url {
		t.Errorf("WrapText long word: want single line with full URL, got %v", got)
	}
}

func TestWrapText_TrailingArticle_MovedDown(t *testing.T) {
	// "hello the" at width=11: "hello the" fits (9 < 11), then " world" would push to 15.
	// Orphan: "the" is pulled to next line with "world".
	got := wrap.WrapText("hello the world", 11, true)
	want := []string{"hello", "the world"}
	if !equalSlices(got, want) {
		t.Errorf("WrapText orphan article:\nwant %v\n got %v", want, got)
	}
}

func TestWrapText_TrailingPrepPhrase_MovedDown(t *testing.T) {
	// "walk in the" fits at width=11 (11 chars), " rain" overflows.
	// Orphan phrase "in the" pulled with "rain".
	got := wrap.WrapText("walk in the rain", 11, true)
	want := []string{"walk", "in the rain"}
	if !equalSlices(got, want) {
		t.Errorf("WrapText orphan prep+article:\nwant %v\n got %v", want, got)
	}
}

func TestWrapText_PunctuatedOrphan_NotMoved(t *testing.T) {
	// "the," is punctuated — not a bare orphan, so normal break.
	got := wrap.WrapText("hello the, world", 11, true)
	want := []string{"hello the,", "world"}
	if !equalSlices(got, want) {
		t.Errorf("WrapText punctuated orphan:\nwant %v\n got %v", want, got)
	}
}

func TestWrapText_NoTrailingSpaceOnLines(t *testing.T) {
	got := wrap.WrapText("one two three", 7, true)
	for _, line := range got {
		if strings.HasSuffix(line, " ") {
			t.Errorf("WrapText line has trailing space: %q", line)
		}
	}
}

func TestWrapText_SingleWord_ReturnedAsIs(t *testing.T) {
	got := wrap.WrapText("hello", 3, true)
	if len(got) != 1 || got[0] != "hello" {
		t.Errorf("WrapText single word overflow: want [\"hello\"], got %v", got)
	}
}

// --- WrapWithPrefix ---

func TestWrapWithPrefix_NoWrap_PrependPrefix(t *testing.T) {
	got := wrap.WrapWithPrefix("line1\nline2", 80, false, "│ ")
	want := []string{"│ line1", "│ line2"}
	if !equalSlices(got, want) {
		t.Errorf("WrapWithPrefix no-wrap:\nwant %v\n got %v", want, got)
	}
}

func TestWrapWithPrefix_Wrap_ReducesWidth(t *testing.T) {
	// prefix "│ " is 2 wide; total width 10 → content width 8.
	got := wrap.WrapWithPrefix("one two three four", 10, true, "│ ")
	for _, line := range got {
		if wrap.VisibleWidth(line) > 10 {
			t.Errorf("WrapWithPrefix line exceeds total width 10: %q", line)
		}
		if !strings.HasPrefix(line, "│ ") {
			t.Errorf("WrapWithPrefix line missing prefix: %q", line)
		}
	}
	if len(got) < 2 {
		t.Errorf("WrapWithPrefix expected wrapping, got %v", got)
	}
}

func TestWrapWithPrefix_EmptyPrefix(t *testing.T) {
	got := wrap.WrapWithPrefix("hello world", 7, true, "")
	for _, line := range got {
		if wrap.VisibleWidth(line) > 7 {
			t.Errorf("WrapWithPrefix empty prefix line too wide: %q", line)
		}
	}
}

// helpers

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
