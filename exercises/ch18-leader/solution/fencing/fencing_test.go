package fencing

import (
	"context"
	"errors"
	"os"
	"testing"

	"example.com/ch18-leader/internal/testdb"
)

func TestMain(m *testing.M) { os.Exit(testdb.Main(m)) }

func TestStaleTokenIsRejected(t *testing.T) {
	pool := testdb.Pool(t)
	ctx := context.Background()
	if _, err := pool.Exec(ctx, Schema); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `DELETE FROM fenced_resources WHERE name = 'config'`); err != nil {
		t.Fatal(err) // a fresh start for every run, also with -count=N
	}
	s := Store{DB: pool}
	steps := []struct {
		value   string
		token   int64
		wantErr error
	}{
		{"from term 33", 33, nil},
		{"term 33 again", 33, nil},          // the same leader writes twice
		{"from term 34", 34, nil},           // a new leader
		{"late term 33", 33, ErrStaleToken}, // the old leader wakes up
		{"from term 36", 36, nil},           // terms may skip numbers
	}
	for _, st := range steps {
		err := s.Write(ctx, "config", st.value, st.token)
		if !errors.Is(err, st.wantErr) {
			t.Fatalf("Write(%q, token %d) = %v, want %v", st.value, st.token, err, st.wantErr)
		}
	}
	value, token, err := s.Read(ctx, "config")
	if err != nil || value != "from term 36" || token != 36 {
		t.Fatalf("Read = %q, %d, %v; want the term 36 write", value, token, err)
	}
}
