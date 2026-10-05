// Command ledger is a small CLI over package ledger, backed by a JSON file.
//
//	ledger [-file ledger.json] open <id> <owner>
//	ledger [-file ledger.json] deposit <id> <amount> [memo]
//	ledger [-file ledger.json] withdraw <id> <amount> [memo]
//	ledger [-file ledger.json] balance <id>
//	ledger [-file ledger.json] history [-cursor N] [-limit N] <id>
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	ledger "example.com/ch04-ledger"
)

// Exit codes. Scripts can tell "you asked wrong" from "the ledger said no".
const (
	exitOK           = 0
	exitError        = 1
	exitUsage        = 2
	exitInsufficient = 3
	exitNotFound     = 4
)

const usage = `usage: ledger [-file path] <command> [args]
commands:
  open <id> <owner>
  deposit <id> <amount> [memo]
  withdraw <id> <amount> [memo]
  balance <id>
  history [-cursor N] [-limit N] <id>`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, time.Now))
}

func run(args []string, stdout, stderr io.Writer, now func() time.Time) int {
	fs := flag.NewFlagSet("ledger", flag.ContinueOnError)
	fs.SetOutput(stderr)
	file := fs.String("file", "ledger.json", "ledger file")
	if err := fs.Parse(args); err != nil {
		return exitUsage
	}
	if fs.NArg() == 0 {
		fmt.Fprintln(stderr, usage)
		return exitUsage
	}
	store, err := ledger.OpenFileStore(*file)
	if err != nil {
		fmt.Fprintln(stderr, "ledger:", err)
		return exitError
	}
	l := ledger.New(store, now)

	err = dispatch(l, fs.Arg(0), fs.Args()[1:], stdout, stderr)
	return report(err, stderr)
}

// errUsage marks a mistake in the command line itself.
var errUsage = errors.New("usage")

func dispatch(l *ledger.Ledger, cmd string, args []string, stdout, stderr io.Writer) error {
	switch cmd {
	case "open":
		if len(args) != 2 {
			return errUsage
		}
		a, err := l.Open(ledger.AccountID(args[0]), args[1])
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "opened %s for %s\n", a.ID, a.Owner)
	case "deposit", "withdraw":
		if len(args) < 2 || len(args) > 3 {
			return errUsage
		}
		amount, err := ledger.ParseCents(args[1])
		if err != nil {
			return fmt.Errorf("%w: %w", errUsage, err)
		}
		memo := ""
		if len(args) == 3 {
			memo = args[2]
		}
		post := l.Deposit
		if cmd == "withdraw" {
			post = l.Withdraw // a method value: the method bound to l
		}
		a, err := post(ledger.AccountID(args[0]), amount, memo)
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "%s balance %s\n", a.ID, a.Balance)
	case "balance":
		if len(args) != 1 {
			return errUsage
		}
		b, err := l.Balance(ledger.AccountID(args[0]))
		if err != nil {
			return err
		}
		fmt.Fprintln(stdout, b)
	case "history":
		return history(l, args, stdout, stderr)
	default:
		return fmt.Errorf("%w: unknown command %q", errUsage, cmd)
	}
	return nil
}

func history(l *ledger.Ledger, args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("history", flag.ContinueOnError)
	fs.SetOutput(stderr)
	cursor := fs.Int("cursor", 0, "position to start from")
	limit := fs.Int("limit", 10, "transactions per page")
	if err := fs.Parse(args); err != nil || fs.NArg() != 1 {
		return errUsage
	}
	id := ledger.AccountID(fs.Arg(0))
	page, err := l.History(id, *cursor, *limit)
	if err != nil {
		return err
	}
	// TODO(day 7): turn the page into lines with page.Map, using this format:
	//   fmt.Sprintf("#%-3d %s  %-10v %10s  %s", tx.ID, tx.At.Format("2006-01-02 15:04"), tx.Kind, tx.Amount, tx.Memo)
	// print each line with trailing spaces trimmed (strings.TrimRight), and when
	// page.HasMore print: more: ledger history -cursor <Next> -limit <limit> <id>
	_ = page
	return errors.New("TODO: history output")
}

// report is the one place that turns errors into messages and exit codes.
func report(err error, stderr io.Writer) int {
	if err == nil {
		return exitOK
	}
	fmt.Fprintln(stderr, "ledger:", err)
	// TODO(day 7): map errors to exit codes, in this order:
	//   InsufficientFundsError (errors.AsType)  -> message with "(short by X)", exitInsufficient
	//   errUsage (errors.Is)                    -> also print usage, exitUsage
	//   ledger.ErrAccountNotFound (errors.Is)   -> exitNotFound
	//   anything else                           -> exitError
	return exitError
}
