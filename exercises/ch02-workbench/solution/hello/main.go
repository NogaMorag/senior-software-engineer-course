// Command hello prints a greeting. It is the smoke test for the Go toolchain (chapter 2).
package main

import (
	"fmt"
	"os"
	"strings"
)

// greeting returns "Hello, <name>!", or "Hello, world!" when name is empty or only spaces.
func greeting(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "world"
	}
	return "Hello, " + name + "!"
}

func main() {
	fmt.Println(greeting(strings.Join(os.Args[1:], " ")))
}
