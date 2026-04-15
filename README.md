A CLI to view well rendered markdown in the terminal
====================================================

A port of https://www.github.com/steipete/markdansi from typescript to golang,
so that I can have a single binary. Most of the heavy lifting done by copilot.

This should absolutely be considered a derivative work of the above, despite
efforts being made to generate intermediate human readable representations of
the code as you would in a clean-room reimplementation, because considering it
otherwise would be a dick move when somebody else did all the hard work.

Continuing the MIT License from the source, and portions (c) Peter Steinberger
though it is technically difficult to identify which specific portions

---

## Install

### Pre-built binaries

Download the latest release for your platform from the
[releases page](https://github.com/funcan/showmd/releases).

### go install

```sh
go install github.com/funcan/showmd/cmd/showmd@latest
```

### Build from source

```sh
git clone https://github.com/funcan/showmd
cd showmd
make build
# binary is at ./showmd
```

---

## Usage

```
showmd [options] [file]
```

Reads Markdown from `file` (or stdin if omitted) and writes ANSI-rendered
output to stdout.

```sh
# Render a file
showmd README.md

# Pipe from stdin
cat README.md | showmd

# Pipe to a pager
showmd README.md | less -R

# Write to a file (strips ANSI)
showmd --no-color README.md --out plain.txt
```

---

## Options

| Flag | Default | Description |
| --- | --- | --- |
| `file` / `--in FILE` | stdin | Input Markdown file (`-` for explicit stdin) |
| `--out FILE` | stdout | Output file |
| `--width N` | TTY width or 80 | Column wrap width |
| `--no-wrap` | — | Disable word-wrapping |
| `--no-color` | — | Disable all ANSI/OSC output |
| `--no-links` | — | Disable OSC-8 hyperlinks |
| `--theme NAME` | `default` | Color theme: `default` `dim` `bright` `solarized` `monochrome` `contrast` |
| `--list-indent N` | `2` | Spaces per list nesting level |
| `--quote-prefix STR` | `│ ` | Blockquote line prefix |
| `--table-border STR` | `unicode` | Table border style: `unicode` `ascii` `none` |
| `--table-padding N` | `1` | Spaces of padding inside table cells |
| `--table-dense` | — | Omit the mid-table separator row |
| `--table-truncate=BOOL` | `true` | Truncate cells that exceed column width |
| `--table-ellipsis STR` | `…` | String used when truncating |
| `--code-wrap=BOOL` | `true` | Wrap long code lines at `--width` |
| `--code-box=BOOL` | `true` | Draw a Unicode box around multi-line code blocks |
| `--code-gutter=BOOL` | `false` | Show line numbers in code blocks |
| `--help`, `-h` | — | Print help and exit |

---

## Library API

```go
import "github.com/funcan/showmd/internal/render"
import "github.com/funcan/showmd/internal/types"

// Render Markdown to an ANSI string.
out := render.Render(markdown, types.RenderOptions{})

// Strip all ANSI codes from the result (plain text).
plain := render.Strip(markdown, types.RenderOptions{})

// Create a reusable renderer function.
renderer := render.CreateRenderer(types.RenderOptions{})
out = renderer(markdown)
```

---

## License

MIT — see [LICENSE](LICENSE).

