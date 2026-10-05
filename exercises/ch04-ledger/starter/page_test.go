package ledger

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestPaginate(t *testing.T) {
	items := []string{"a", "b", "c", "d", "e"}
	tests := []struct {
		name          string
		cursor, limit int
		want          Page[string]
		wantErr       bool
	}{
		{"first page", 0, 2, Page[string]{Items: []string{"a", "b"}, Next: 2, HasMore: true}, false},
		{"middle page", 2, 2, Page[string]{Items: []string{"c", "d"}, Next: 4, HasMore: true}, false},
		{"last partial page", 4, 2, Page[string]{Items: []string{"e"}, Next: 5, HasMore: false}, false},
		{"limit larger than list", 0, 10, Page[string]{Items: items, Next: 5, HasMore: false}, false},
		{"cursor at the end is an empty last page", 5, 2, Page[string]{Items: []string{}, Next: 5, HasMore: false}, false},
		{"cursor past the end", 6, 2, Page[string]{}, true},
		{"negative cursor", -1, 2, Page[string]{}, true},
		{"zero limit", 0, 0, Page[string]{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Paginate(items, tt.cursor, tt.limit)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !slices.Equal(got.Items, tt.want.Items) || got.Next != tt.want.Next || got.HasMore != tt.want.HasMore {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestPaginateBadCursorIsSentinel(t *testing.T) {
	_, err := Paginate([]int{1, 2}, 9, 1)
	if !errors.Is(err, ErrBadCursor) {
		t.Fatalf("error = %v, want ErrBadCursor", err)
	}
}

func TestPageItemsDoNotAliasTheSource(t *testing.T) {
	items := []int{1, 2, 3, 4}
	page, err := Paginate(items, 0, 2)
	if err != nil {
		t.Fatal(err)
	}
	_ = append(page.Items, 99) // must copy, not overwrite items[2]
	if items[2] != 3 {
		t.Fatalf("append to a page overwrote the source: %v", items)
	}
}

func TestPageMap(t *testing.T) {
	p := Page[int]{Items: []int{1, 2}, Next: 2, HasMore: true}
	got := p.Map(func(n int) string { return strings.Repeat("*", n) })
	if !slices.Equal(got.Items, []string{"*", "**"}) || got.Next != 2 || !got.HasMore {
		t.Fatalf("Map = %+v", got)
	}
}

func TestSumWorksForNamedIntegerTypes(t *testing.T) {
	if got := Sum([]Cents{1_50, 2_25}); got != 3_75 {
		t.Errorf("Sum(Cents) = %v, want 3.75", got)
	}
	if got := Sum([]int{1, 2, 3}); got != 6 {
		t.Errorf("Sum(int) = %d, want 6", got)
	}
	if got := Sum[int](nil); got != 0 {
		t.Errorf("Sum(nil) = %d, want 0", got)
	}
}
