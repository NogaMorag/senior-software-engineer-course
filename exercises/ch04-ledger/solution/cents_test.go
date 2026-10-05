package ledger

import "testing"

func TestParseCents(t *testing.T) {
	tests := []struct {
		in      string
		want    Cents
		wantErr bool
	}{
		{"12", 12_00, false},
		{"12.5", 12_50, false},
		{"12.50", 12_50, false},
		{"0.07", 7, false},
		{"0", 0, false},
		{"", 0, true},
		{"12.", 0, true},
		{".5", 0, true},
		{"12.345", 0, true},
		{"-3", 0, true},
		{"+3", 0, true},
		{"1,000", 0, true},
		{"12.a", 0, true},
		{"99999999999999999999", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := ParseCents(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseCents(%q) error = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ParseCents(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestCentsString(t *testing.T) {
	for c, want := range map[Cents]string{0: "0.00", 7: "0.07", 12_50: "12.50", -3_05: "-3.05"} {
		if got := c.String(); got != want {
			t.Errorf("Cents(%d).String() = %q, want %q", int64(c), got, want)
		}
	}
}

func TestMustParseCentsPanicsOnBadInput(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("MustParseCents(\"oops\") did not panic")
		}
	}()
	MustParseCents("oops")
}
