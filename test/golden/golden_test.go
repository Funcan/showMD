package golden_test

import (
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/funcan/showmd/internal/render"
	"github.com/funcan/showmd/internal/types"
)

var update = flag.Bool("update", false, "update golden files")

// renderGolden renders sample.md with deterministic options (no color, wrap at 80).
func renderGolden(t *testing.T) string {
	t.Helper()
	src, err := os.ReadFile(filepath.Join("testdata", "sample.md"))
	if err != nil {
		t.Fatalf("read sample.md: %v", err)
	}
	f := false
	tr := true
	w := 80
	return render.Render(string(src), types.RenderOptions{
		Color:      &f,
		Hyperlinks: &f,
		Wrap:       &tr,
		Width:      &w,
	})
}

func TestGolden_SampleDocument(t *testing.T) {
	got := renderGolden(t)
	golden := filepath.Join("testdata", "sample.golden")

	if *update {
		if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		t.Logf("updated %s", golden)
		return
	}

	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("read golden file %s: %v\n(run with -update to create it)", golden, err)
	}

	if got != string(want) {
		diff := diffLines(string(want), got)
		t.Errorf("golden mismatch (-want +got):\n%s", diff)
	}
}

// diffLines produces a simple line-by-line diff for test output.
func diffLines(want, got string) string {
	wLines := strings.Split(want, "\n")
	gLines := strings.Split(got, "\n")
	var b strings.Builder
	max := len(wLines)
	if len(gLines) > max {
		max = len(gLines)
	}
	for i := 0; i < max; i++ {
		var w, g string
		if i < len(wLines) {
			w = wLines[i]
		}
		if i < len(gLines) {
			g = gLines[i]
		}
		if w != g {
			b.WriteString("line ")
			b.WriteString(strconv.Itoa(i + 1))
			b.WriteString(":\n  want: ")
			b.WriteString(w)
			b.WriteString("\n  got:  ")
			b.WriteString(g)
			b.WriteString("\n")
		}
	}
	return b.String()
}
