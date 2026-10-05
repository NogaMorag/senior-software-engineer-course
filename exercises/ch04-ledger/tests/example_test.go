package ledger

import (
	"errors"
	"fmt"
)

// lookup returns a nil *InsufficientFundsError inside a non-nil error interface.
// This is the bug: the interface holds a type, so it is not nil.
func lookup() error {
	var e *InsufficientFundsError // nil pointer
	return e                      // wrapped in an interface: (type=*InsufficientFundsError, value=nil)
}

func Example_nilInterfaceTrap() {
	err := lookup()
	fmt.Println(err == nil)
	fmt.Printf("%T\n", err)
	// Output:
	// false
	// *ledger.InsufficientFundsError
}

// Embedding promotes methods, but it is not inheritance: FileStore's methods
// are found first, and MemStore's are still reachable by their full name.
func Example_embeddingPromotes() {
	fs := &FileStore{path: "unused.json"}
	_, err := fs.Account("x") // promoted from MemStore
	fmt.Println(errors.Is(err, ErrAccountNotFound))
	_ = fs.MemStore.CreateAccount(Account{ID: "x"}) // the embedded method: no save
	a, _ := fs.Account("x")
	fmt.Println(a.ID)
	// Output:
	// true
	// x
}

func ExampleSum() {
	fmt.Println(Sum([]Cents{12_50, 7_25}))
	// Output: 19.75
}
