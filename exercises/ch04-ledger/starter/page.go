package ledger

import (
	"errors"
)

// ErrBadCursor is returned for a cursor that does not point into the list.
var ErrBadCursor = errors.New("bad cursor")

// Page is one page of any list. Next is the cursor of the following page;
// it is meaningful only when HasMore is true.
type Page[T any] struct {
	Items   []T
	Next    int
	HasMore bool
}

// Paginate returns items[cursor : cursor+limit] as a Page. The cursor here is a
// plain offset; chapter 9 explains why real APIs prefer opaque cursors.
func Paginate[T any](items []T, cursor, limit int) (Page[T], error) {
	// TODO(day 7): limit < 1 is an error; a cursor outside 0..len(items) wraps
	// ErrBadCursor with %w; otherwise return items[cursor:end:end] (why the third
	// index? see chapter 3), Next = end, HasMore = end < len(items).
	return Page[T]{}, errors.New("TODO: Paginate")
}

// Map converts every item and keeps the paging fields. It is a generic method,
// with its own type parameter U; Go supports these since 1.27.
func (p Page[T]) Map[U any](f func(T) U) Page[U] {
	// TODO(day 7): apply f to every item; copy Next and HasMore.
	return Page[U]{}
}

// Integer is a constraint: any type whose underlying type is one of these.
// The ~ is what lets Cents (underlying int64) satisfy it.
type Integer interface {
	~int | ~int32 | ~int64
}

// Sum adds up values of any Integer type and returns the same type.
func Sum[N Integer](values []N) N {
	var total N // the zero value of N
	// TODO(day 7): add up values.
	return total
}
