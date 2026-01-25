package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
)

type Formatter struct {
	json    bool
	quiet   bool
	noColor bool
	writer  io.Writer
}

func New(jsonOutput, quiet, noColor bool) *Formatter {
	return &Formatter{
		json:    jsonOutput,
		quiet:   quiet,
		noColor: noColor || os.Getenv("NO_COLOR") != "",
		writer:  os.Stdout,
	}
}

func (f *Formatter) JSON(data interface{}) error {
	enc := json.NewEncoder(f.writer)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

func (f *Formatter) Print(format string, args ...interface{}) {
	if !f.quiet {
		_, _ = fmt.Fprintf(f.writer, format, args...)
	}
}

func (f *Formatter) Println(args ...interface{}) {
	if !f.quiet {
		_, _ = fmt.Fprintln(f.writer, args...)
	}
}

func (f *Formatter) Error(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(os.Stderr, format, args...)
}

func (f *Formatter) IsJSON() bool {
	return f.json
}

func (f *Formatter) Table(headers []string, rows [][]string) {
	w := tabwriter.NewWriter(f.writer, 0, 0, 2, ' ', 0)

	// Print headers
	for i, h := range headers {
		if i > 0 {
			_, _ = fmt.Fprint(w, "\t")
		}
		_, _ = fmt.Fprint(w, h)
	}
	_, _ = fmt.Fprintln(w)

	// Print rows
	for _, row := range rows {
		for i, cell := range row {
			if i > 0 {
				_, _ = fmt.Fprint(w, "\t")
			}
			_, _ = fmt.Fprint(w, cell)
		}
		_, _ = fmt.Fprintln(w)
	}

	_ = w.Flush()
}

func (f *Formatter) Success(msg string) {
	if f.noColor {
		f.Println("✓", msg)
	} else {
		f.Println("\033[32m✓\033[0m", msg)
	}
}

func (f *Formatter) Warn(msg string) {
	if f.noColor {
		_, _ = fmt.Fprintln(os.Stderr, "⚠", msg)
	} else {
		_, _ = fmt.Fprintln(os.Stderr, "\033[33m⚠\033[0m", msg)
	}
}
