package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/funcan/showmd/internal/render"
	"github.com/funcan/showmd/internal/types"
)

// Set by goreleaser via -ldflags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

const helpText = `showmd – render Markdown to ANSI in your terminal

Usage: showmd [options] [file]

  file                  Input file (default: stdin; use - for stdin explicitly)

Options:
  --in FILE             Input file (alternative to positional arg)
  --out FILE            Output file (default: stdout)
  --width N             Wrap width (default: TTY width or 80)
  --no-wrap             Disable hard word-wrapping
  --no-color            Disable ANSI/OSC output
  --no-links            Disable OSC-8 hyperlinks
  --theme=NAME          Theme: default|dim|bright|solarized|monochrome|contrast
  --list-indent N       Spaces per list nesting level (default: 2)
  --quote-prefix STR    Blockquote prefix (default: "│ ")
  --table-border=STR    unicode|ascii|none (default: unicode)
  --table-padding N     Spaces around cell content (default: 1)
  --table-dense         Omit mid-table separator row
  --table-truncate=BOOL Truncate cells to fit column width (default: true)
  --table-ellipsis STR  Truncation marker (default: "…")
  --code-wrap=BOOL      Wrap long code lines (default: true)
  --code-box=BOOL       Draw box around code blocks (default: true)
  --code-gutter=BOOL    Show line numbers in code blocks (default: false)
  --help, -h            Show this help
  --version             Print version and exit
`

// cliArgs holds parsed CLI arguments.
type cliArgs struct {
	in      string // input file path ("" or "-" = stdin)
	out     string // output file path ("" = stdout)
	help    bool
	version bool
	opts    types.RenderOptions
}

// parseArgs parses os.Args-style argument slice (starting from index 1).
func parseArgs(argv []string) (*cliArgs, error) {
	a := &cliArgs{}
	args := argv[1:] // skip program name

	boolFlag := func(dst **bool, v bool) {
		b := v
		*dst = &b
	}
	intFlag := func(dst **int, s string) error {
		n, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("expected integer, got %q", s)
		}
		*dst = &n
		return nil
	}
	strFlag := func(dst **string, s string) {
		*dst = &s
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		next := func() (string, bool) {
			if i+1 < len(args) {
				i++
				return args[i], true
			}
			return "", false
		}

		switch {
		case arg == "--help" || arg == "-h":
			a.help = true

		case arg == "--version":
			a.version = true

		case arg == "--no-wrap":
			boolFlag(&a.opts.Wrap, false)

		case arg == "--no-color":
			boolFlag(&a.opts.Color, false)

		case arg == "--no-links":
			boolFlag(&a.opts.Hyperlinks, false)

		case arg == "--in":
			if v, ok := next(); ok {
				a.in = v
			}

		case arg == "--out":
			if v, ok := next(); ok {
				a.out = v
			}

		case arg == "--width":
			if v, ok := next(); ok {
				if err := intFlag(&a.opts.Width, v); err != nil {
					return nil, fmt.Errorf("--width: %w", err)
				}
			}

		case arg == "--list-indent":
			if v, ok := next(); ok {
				if err := intFlag(&a.opts.ListIndent, v); err != nil {
					return nil, fmt.Errorf("--list-indent: %w", err)
				}
			}

		case arg == "--table-padding":
			if v, ok := next(); ok {
				if err := intFlag(&a.opts.TablePadding, v); err != nil {
					return nil, fmt.Errorf("--table-padding: %w", err)
				}
			}

		case arg == "--quote-prefix":
			if v, ok := next(); ok {
				strFlag(&a.opts.QuotePrefix, v)
			}

		case arg == "--table-ellipsis":
			if v, ok := next(); ok {
				strFlag(&a.opts.TableEllipsis, v)
			}

		case arg == "--table-dense":
			a.opts.TableDense = true

		case strings.HasPrefix(arg, "--theme="):
			a.opts.ThemeName = types.ThemeName(strings.TrimPrefix(arg, "--theme="))

		case strings.HasPrefix(arg, "--table-border="):
			val := strings.TrimPrefix(arg, "--table-border=")
			switch types.TableBorder(val) {
			case types.TableBorderUnicode, types.TableBorderASCII, types.TableBorderNone:
				a.opts.TableBorder = types.TableBorder(val)
			default:
				return nil, fmt.Errorf("--table-border: unknown value %q (use unicode|ascii|none)", val)
			}

		case arg == "--table-truncate=true":
			boolFlag(&a.opts.TableTruncate, true)
		case arg == "--table-truncate=false":
			boolFlag(&a.opts.TableTruncate, false)

		case arg == "--code-wrap=true":
			boolFlag(&a.opts.CodeWrap, true)
		case arg == "--code-wrap=false":
			boolFlag(&a.opts.CodeWrap, false)

		case arg == "--code-box=true":
			boolFlag(&a.opts.CodeBox, true)
		case arg == "--code-box=false":
			boolFlag(&a.opts.CodeBox, false)

		case arg == "--code-gutter=true":
			a.opts.CodeGutter = true
		case arg == "--code-gutter=false":
			a.opts.CodeGutter = false

		case arg == "-":
			// Explicit stdin positional arg — leave a.in empty.

		case !strings.HasPrefix(arg, "-"):
			// Positional argument: treat as input file.
			a.in = arg

		default:
			return nil, fmt.Errorf("unknown flag: %s", arg)
		}
	}
	return a, nil
}

// run executes the CLI with the given arguments, reading from r and writing to w.
// Returns an exit code (0 = success, 1 = error).
func run(a *cliArgs, r io.Reader, w io.Writer, stderr io.Writer) int {
	if a.help {
		fmt.Fprint(w, helpText)
		return 0
	}
	if a.version {
		fmt.Fprintf(w, "showmd %s (%s) built %s\n", version, commit, date)
		return 0
	}

	// Read input.
	var input []byte
	var err error
	if a.in != "" && a.in != "-" {
		input, err = os.ReadFile(a.in)
		if err != nil {
			fmt.Fprintf(stderr, "showmd: %v\n", err)
			return 1
		}
	} else {
		input, err = io.ReadAll(r)
		if err != nil {
			fmt.Fprintf(stderr, "showmd: reading stdin: %v\n", err)
			return 1
		}
	}

	output := render.Render(string(input), a.opts)

	// Write output.
	if a.out != "" {
		if err := os.WriteFile(a.out, []byte(output), 0o644); err != nil {
			fmt.Fprintf(stderr, "showmd: %v\n", err)
			return 1
		}
	} else {
		if _, err := fmt.Fprint(w, output); err != nil {
			// Ignore EPIPE (downstream closed the pipe, e.g. `head`).
			if isEPIPE(err) {
				return 0
			}
			fmt.Fprintf(stderr, "showmd: writing output: %v\n", err)
			return 1
		}
	}
	return 0
}

func isEPIPE(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "broken pipe") || strings.Contains(s, "EPIPE")
}

func main() {
	// Ignore SIGPIPE so we don't crash when a downstream consumer (e.g. head)
	// closes the pipe early. We also check the write error in run().
	signal.Ignore(syscall.SIGPIPE)

	a, err := parseArgs(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "showmd: %v\n", err)
		os.Exit(1)
	}

	os.Exit(run(a, os.Stdin, os.Stdout, os.Stderr))
}
