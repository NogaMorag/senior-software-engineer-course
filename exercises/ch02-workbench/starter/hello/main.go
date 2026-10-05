// Command hello prints a greeting. It is the smoke test for the Go toolchain (chapter 2).
package main

import (
	"fmt"
	"os"
	"strings"
)

// greeting returns "Hello, <name>!", or "Hello, world!" when name is empty or only spaces.
func greeting(name string) string {
	// TODO (day 3): trim the spaces around name, and use "world" when nothing is left.
	// The table in main_test.go lists every case.
	return "Hello, " + name + "!"
}

func main() {
	fmt.Println(greeting(strings.Join(os.Args[1:], " ")))
}
