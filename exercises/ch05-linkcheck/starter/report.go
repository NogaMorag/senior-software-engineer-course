package linkcheck

import (
	"fmt"
	"io"
	"strconv"
)

// WriteReport writes one line per result, in the order given, followed by a
// summary line. It returns how many results were not ok.
func WriteReport(w io.Writer, results []Result) (notOK int, err error) {
	for _, r := range results {
		status := "-"
		if r.Status != 0 {
			status = strconv.Itoa(r.Status)
		}
		if r.Err != nil {
			notOK++
		}
		if _, err := fmt.Fprintf(w, "%-9s %3s  %s\n", r.Verdict(), status, r.URL); err != nil {
			return notOK, err
		}
	}
	_, err = fmt.Fprintf(w, "checked %d, ok %d, not ok %d\n", len(results), len(results)-notOK, notOK)
	return notOK, err
}
