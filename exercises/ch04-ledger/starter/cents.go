package ledger

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Cents is an amount of money in hundredths. Money is never a float.
type Cents int64

// String implements fmt.Stringer: Cents(1250) prints as "12.50".
func (c Cents) String() string {
	sign := ""
	if c < 0 {
		sign, c = "-", -c
	}
	return fmt.Sprintf("%s%d.%02d", sign, c/100, c%100)
}

// ParseCents parses a non-negative amount: "12", "12.5" or "12.50" all work.
func ParseCents(s string) (Cents, error) {
	whole, frac, hasDot := strings.Cut(s, ".")
	if !digits(whole) || len(frac) > 2 || (hasDot && !digits(frac)) {
		return 0, fmt.Errorf("parse amount %q: want digits with at most two decimals", s)
	}
	frac = (frac + "00")[:2] // "" -> "00", "5" -> "50"
	w, err := strconv.ParseInt(whole, 10, 64)
	if err != nil || w > math.MaxInt64/100-1 {
		return 0, fmt.Errorf("parse amount %q: too large", s)
	}
	f, _ := strconv.ParseInt(frac, 10, 64) // two ASCII digits always parse
	return Cents(w*100 + f), nil
}

func digits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// MustParseCents is for constants in code and tests. It panics on bad input,
// because bad input here is a programmer error, not a runtime condition.
func MustParseCents(s string) Cents {
	c, err := ParseCents(s)
	if err != nil {
		panic(err)
	}
	return c
}
