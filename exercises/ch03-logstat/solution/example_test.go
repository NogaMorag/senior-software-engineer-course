package main

import (
	"fmt"
	"slices"
	"unicode/utf8"
)

// The aliasing bug: head shares its backing array with paths, so append writes into paths.
func Example_appendAliasing() {
	paths := []string{"/a", "/b", "/c", "/d"}
	head := paths[:2]           // len 2, cap 4: same backing array
	head = append(head, "/new") // fits in cap, so it overwrites paths[2]
	fmt.Println(paths, len(head), cap(head))

	paths = []string{"/a", "/b", "/c", "/d"}
	safe := slices.Clone(paths[:2]) // own backing array
	safe = append(safe, "/new")
	limited := paths[:2:2] // full slice expression: cap 2, so append must copy
	limited = append(limited, "/new")
	fmt.Println(paths, safe, limited)
	// Output:
	// [/a /b /new /d] 3 4
	// [/a /b /c /d] [/a /b /new] [/a /b /new]
}

// Arrays are values, maps and slices are references to shared storage.
func Example_valuesAndReferences() {
	s := NewStats()
	classes := s.Classes // copies all six ints
	classes[2]++
	paths := s.Paths // copies the map header only: same map
	paths["/x"]++
	fmt.Println(s.Classes[2], s.Paths["/x"])
	// Output: 0 1
}

// A string is bytes. Hebrew letters are two bytes each in UTF-8.
func Example_hebrewBytesAndRunes() {
	word := "שלום"
	fmt.Println(len(word), utf8.RuneCountInString(word))
	fmt.Println(utf8.ValidString(word[:3])) // a byte slice cut through a letter
	for i, r := range "אב!" {
		fmt.Println(i, string(r)) // i is the byte offset, r is the rune
	}
	fmt.Println(Truncate("/מוצרים/נעליים", 8))
	// Output:
	// 8 4
	// false
	// 0 א
	// 2 ב
	// 4 !
	// /מוצרים…
}
