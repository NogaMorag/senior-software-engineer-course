// Command linkcheck reads URLs (one per line) from a file or stdin, checks them
// concurrently and prints a report in input order.
//
//	linkcheck [-workers 4] [-timeout 5s] [-fail-fast] [file]
//
// Exit status: 0 all ok, 1 some links not ok, 2 usage or input error,
// 130 interrupted.
package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	linkcheck "example.com/ch05-linkcheck"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("linkcheck", flag.ContinueOnError)
	fs.SetOutput(stderr)
	workers := fs.Int("workers", 4, "maximum requests in flight")
	timeout := fs.Duration("timeout", 5*time.Second, "timeout for each request")
	failFast := fs.Bool("fail-fast", false, "stop at the first link that is not ok")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	in := stdin
	if fs.NArg() == 1 {
		f, err := os.Open(fs.Arg(0))
		if err != nil {
			fmt.Fprintln(stderr, "linkcheck:", err)
			return 2
		}
		defer f.Close()
		in = f
	}
	urls, err := readURLs(in)
	if err != nil {
		fmt.Fprintln(stderr, "linkcheck:", err)
		return 2
	}

	// Ctrl-C or SIGTERM cancels ctx; the checker stops every request in flight.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	c := &linkcheck.Checker{Workers: *workers, Timeout: *timeout, FailFast: *failFast}
	results, runErr := c.Check(ctx, urls)
	notOK, err := linkcheck.WriteReport(stdout, results)
	if err != nil {
		fmt.Fprintln(stderr, "linkcheck:", err)
		return 2
	}
	switch {
	case errors.Is(runErr, context.Canceled):
		fmt.Fprintln(stderr, "linkcheck: interrupted")
		return 130
	case notOK > 0:
		return 1
	}
	return 0
}

// readURLs returns the non-empty lines of r that do not start with '#'.
func readURLs(r io.Reader) ([]string, error) {
	var urls []string
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		urls = append(urls, line)
	}
	return urls, sc.Err()
}
