package ledger

import (
	"errors"
	"fmt"
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
	if limit < 1 {
		return Page[T]{}, fmt.Errorf("limit %d: must be at least 1", limit)
	}
	if cursor < 0 || cursor > len(items) {
		return Page[T]{}, fmt.Errorf("cursor %d of %d items: %w", cursor, len(items), ErrBadCursor)
	}
	end := min(cursor+limit, len(items))
	return Page[T]{
		Items:   items[cursor:end:end], // cap == len: appending to Items cannot touch items
		Next:    end,
		HasMore: end < len(items),
	}, nil
}

// Map converts every item and keeps the paging fields. It is a generic method,
// with its own type parameter U; Go supports these since 1.27.
func (p Page[T]) Map[U any](f func(T) U) Page[U] {
	out := Page[U]{Items: make([]U, 0, len(p.Items)), Next: p.Next, HasMore: p.HasMore}
	for _, item := range p.Items {
		out.Items = append(out.Items, f(item))
	}
	return out
}

// Integer is a constraint: any type whose underlying type is one of these.
// The ~ is what lets Cents (underlying int64) satisfy it.
type Integer interface {
	~int | ~int32 | ~int64
}

// Sum adds up values of any Integer type and returns the same type.
func Sum[N Integer](values []N) N {
	var total N // the zero value of N
	for _, v := range values {
		total += v
	}
	return total
}
