package api_test

import (
	"os"
	"testing"

	"example.com/hookline/internal/store"
	"example.com/hookline/internal/store/memory"
	"example.com/hookline/internal/store/postgres"
	"example.com/hookline/internal/testdb"
)

func TestMain(m *testing.M) { os.Exit(testdb.Main(m)) }

// newStore returns the store the API tests run against. Without DATABASE_URL it is the memory store
// (milestone M1); with it, the PostgreSQL store on a throw-away migrated database (milestone M2). The
// tests are the same either way: that the handlers pass unchanged is the point of the store interfaces.
func newStore(*testing.T) store.Store {
	if pool := testdb.Available(); pool != nil {
		return postgres.New(pool)
	}
	return memory.New()
}
