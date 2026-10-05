// Command logstat summarizes an access log: status classes, top paths, malformed lines.
//
//	logstat [-top N] [-width W] [file]
//
// With no file, or with "-", it reads standard input.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

// run is main without the process: tests call it with fake streams and check the exit code.
func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("logstat", flag.ContinueOnError)
	fs.SetOutput(stderr)
	top := fs.Int("top", 5, "how many paths to show")
	width := fs.Int("width", 40, "maximum path width in characters")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() > 1 {
		fmt.Fprintln(stderr, "usage: logstat [-top N] [-width W] [file]")
		return 2
	}

	in := stdin
	if name := fs.Arg(0); name != "" && name != "-" {
		f, err := os.Open(name)
		if err != nil {
			fmt.Fprintln(stderr, "logstat:", err)
			return 1
		}
		defer f.Close()
		in = f
	}

	s, err := Collect(in)
	if err != nil {
		fmt.Fprintln(stderr, "logstat: read:", err)
		return 1
	}
	report(stdout, s, *top, *width)
	return 0
}

func report(w io.Writer, s *Stats, top, width int) {
	fmt.Fprintf(w, "lines      %d\n", s.Lines)
	fmt.Fprintf(w, "malformed  %d", s.Malformed)
	if len(s.BadLines) > 0 {
		fmt.Fprintf(w, " (first at lines %v)", s.BadLines)
	}
	fmt.Fprintln(w)
	for class := 1; class < len(s.Classes); class++ {
		fmt.Fprintf(w, "%-10s %d\n", StatusClass(class*100), s.Classes[class])
	}
	fmt.Fprintln(w, "top paths:")
	for _, row := range TopPaths(s.Paths, top) {
		fmt.Fprintf(w, "%5d  %s\n", row.Count, Truncate(DisplayPath(row.Path), width))
	}
}
