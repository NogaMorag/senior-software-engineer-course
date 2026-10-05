package memory_test

import (
	"testing"

	"example.com/hookline/internal/store"
	"example.com/hookline/internal/store/memory"
	"example.com/hookline/internal/store/storetest"
)

func TestConformance(t *testing.T) {
	storetest.Run(t, func(*testing.T) store.Store { return memory.New() })
}

func TestActiveSubscriptions(t *testing.T) { storetest.ActiveSubscriptions(t, memory.New()) }
