package showmd_test

import (
	"fmt"

	"github.com/funcan/showmd"
)

func ExampleRender() {
	out := showmd.Render("# Hello\n\nSome *Markdown*.\n", showmd.Terminal{Width: 40}, showmd.Options{})
	fmt.Print(out)
	// Output:
	// Hello
	// Some Markdown.
}
