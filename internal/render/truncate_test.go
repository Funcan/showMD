package render

import (
	"testing"

	"github.com/funcan/showmd/internal/wrap"
)

func TestTruncateCell(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		width    int
		ellipsis string
		want     string
	}{
		{
			name:     "no truncation needed",
			text:     "hello",
			width:    10,
			ellipsis: "…",
			want:     "hello",
		},
		{
			name:     "simple ASCII truncation",
			text:     "hello world",
			width:    8,
			ellipsis: "…",
			want:     "hello w…",
		},
		{
			name:     "width equals ellipsis width",
			text:     "hello",
			width:    1,
			ellipsis: "…",
			want:     "…",
		},
		{
			name:     "wide-char ellipsis rune vs column mismatch",
			text:     "abcdefghij",
			width:    6,
			ellipsis: "🔥", // 1 rune but 2 columns wide
			want:     "abcd🔥",
		},
		{
			name:     "wide-char text truncation",
			text:     "日本語テスト",
			width:    8,
			ellipsis: "…",
			want:     "日本語…",
		},
		{
			name:     "width less than ellipsis width",
			text:     "hello",
			width:    1,
			ellipsis: "🔥", // 2 columns wide
			want:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateCell(tt.text, tt.width, tt.ellipsis)
			gotWidth := wrap.VisibleWidth(got)
			if got != tt.want {
				t.Errorf("truncateCell(%q, %d, %q) = %q (width %d), want %q (width %d)",
					tt.text, tt.width, tt.ellipsis, got, gotWidth, tt.want, wrap.VisibleWidth(tt.want))
			}
			if gotWidth > tt.width {
				t.Errorf("result %q has visible width %d, exceeds target %d",
					got, gotWidth, tt.width)
			}
		})
	}
}
