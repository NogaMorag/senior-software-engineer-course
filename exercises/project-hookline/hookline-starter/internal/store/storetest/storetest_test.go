package storetest_test

import (
	"testing"

	"example.com/hookline/internal/store"
	"example.com/hookline/internal/store/memory"
	"example.com/hookline/internal/store/storetest"
)

// The suite itself must pass against the reference memory store, or it tests nothing.
func TestSuiteAgainstMemory(t *testing.T) {
	storetest.Run(t, func(*testing.T) store.Store { return memory.New() })
}
