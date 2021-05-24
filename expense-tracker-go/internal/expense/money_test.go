package expense

import (
	"strings"
	"testing"
)

func TestParseCents(t *testing.T) {
	valid := []struct {
		in   string
		want Cents
	}{
		{"20", 2000},
		{"12.5", 1250},
		{"12.50", 1250},
		{"0.01", 1},
		{".5", 50},
		{"7.", 700},
		{" 3.10 ", 310},
		{"0", 0}, // allowed here; the service rejects non-positive amounts
	}
	for _, tt := range valid {
		got, err := ParseCents(tt.in)
		if err != nil || got != tt.want {
			t.Errorf("ParseCents(%q) = %d, %v; want %d", tt.in, got, err, tt.want)
		}
	}

	invalid := []struct{ in, want string }{
		{"", "invalid amount"},
		{".", "invalid amount"},
		{"abc", "invalid amount"},
		{"-3", "invalid amount"},
		{"1e3", "invalid amount"},
		{"1,000", "invalid amount"},
		{"1.2.3", "invalid amount"},
		{"1.005", "at most 2 decimal places"},
		{"99999999999999999999", "too large"},
		{"92233720368547758.08", "too large"},
	}
	for _, tt := range invalid {
		if _, err := ParseCents(tt.in); err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Errorf("ParseCents(%q) error = %v; want it to contain %q", tt.in, err, tt.want)
		}
	}
}

func TestCentsString(t *testing.T) {
	for c, want := range map[Cents]string{0: "0.00", 5: "0.05", 1250: "12.50", 2000: "20.00", -150: "-1.50"} {
		if got := c.String(); got != want {
			t.Errorf("Cents(%d).String() = %q, want %q", int64(c), got, want)
		}
	}
}
