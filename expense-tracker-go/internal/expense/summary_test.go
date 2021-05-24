package expense

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// noon returns noon UTC on the given date, so the month and year are the
// same in every local time zone the tests might run in.
func noon(year int, month time.Month, d int) time.Time {
	return time.Date(year, month, d, 12, 0, 0, 0, time.UTC)
}

func TestServiceSummary(t *testing.T) {
	expenses := []ExpenseEntity{
		{ID: 1, Description: "Lunch", Amount: 2000, CreatedAt: noon(2024, time.August, 6)},
		{ID: 2, Description: "Dinner", Amount: 1000, CreatedAt: noon(2024, time.August, 20)},
		{ID: 3, Description: "Coffee", Amount: 10, CreatedAt: noon(2024, time.July, 31)},
		{ID: 4, Description: "Tea", Amount: 20, CreatedAt: noon(2024, time.July, 1)},
		{ID: 5, Description: "New year", Amount: 300, CreatedAt: noon(2024, time.January, 1)},
		{ID: 6, Description: "Old August", Amount: 7000, CreatedAt: noon(2023, time.August, 15)},
		{ID: 7, Description: "Old year end", Amount: 400, CreatedAt: noon(2023, time.December, 31)},
	}

	tests := []struct {
		name string
		opts SummaryOptions
		want Cents
	}{
		{"no filter totals everything", SummaryOptions{}, 2000 + 1000 + 10 + 20 + 300 + 7000 + 400},

		// Month only means that month of the current year (2024 here).
		{"month only uses the current year", SummaryOptions{Month: time.August}, 3000},
		{"month only with no expenses", SummaryOptions{Month: time.March}, 0},
		// 0.10 + 0.20 is exactly 0.30 in cents; as float64 it would not be.
		{"small amounts add up exactly", SummaryOptions{Month: time.July}, 30},

		// Year only means every month of that year.
		{"current year", SummaryOptions{Year: 2024}, 2000 + 1000 + 10 + 20 + 300},
		{"past year", SummaryOptions{Year: 2023}, 7000 + 400},
		{"year with no expenses", SummaryOptions{Year: 2019}, 0},

		// Month and year together pick one exact month.
		{"month of a past year", SummaryOptions{Year: 2023, Month: time.August}, 7000},
		{"month and year with no expenses", SummaryOptions{Year: 2023, Month: time.July}, 0},
		{"year boundary is not crossed", SummaryOptions{Year: 2024, Month: time.January}, 300},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(&memStore{state: State{NextID: 8, Expenses: expenses}})
			svc.now = func() time.Time { return noon(2024, time.October, 26) }

			got, err := svc.Summary(tt.opts)
			if err != nil {
				t.Fatalf("Summary(%+v): %v", tt.opts, err)
			}
			if got != tt.want {
				t.Errorf("Summary(%+v) = %s, want %s", tt.opts, got, tt.want)
			}
		})
	}
}

func TestServiceSummaryEmpty(t *testing.T) {
	got, err := NewService(&memStore{}).Summary(SummaryOptions{})
	if err != nil {
		t.Fatalf("Summary: %v", err)
	}
	if got != 0 {
		t.Errorf("Summary() = %s, want 0.00", got)
	}
}

func TestServiceSummaryInvalidOptions(t *testing.T) {
	tests := []struct {
		name string
		opts SummaryOptions
		want string
	}{
		{"month below range", SummaryOptions{Month: -1}, "invalid month"},
		{"month above range", SummaryOptions{Month: 13}, "invalid month"},
		{"negative year", SummaryOptions{Year: -2024}, "invalid year"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewService(&memStore{}).Summary(tt.opts)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("Summary(%+v) error = %v; want it to contain %q", tt.opts, err, tt.want)
			}
		})
	}
}

func TestServiceSummaryLoadError(t *testing.T) {
	boom := errors.New("disk on fire")

	_, err := NewService(&memStore{loadErr: boom}).Summary(SummaryOptions{})
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want it to wrap %v", err, boom)
	}
}

func TestParseMonth(t *testing.T) {
	valid := map[string]time.Month{"1": time.January, "8": time.August, "12": time.December, " 3 ": time.March}
	for in, want := range valid {
		got, err := ParseMonth(in)
		if err != nil || got != want {
			t.Errorf("ParseMonth(%q) = %v, %v; want %v", in, got, err, want)
		}
	}

	for _, in := range []string{"", "0", "13", "-1", "abc", "8.5"} {
		if _, err := ParseMonth(in); err == nil || !strings.Contains(err.Error(), "invalid month") {
			t.Errorf("ParseMonth(%q) error = %v; want an invalid month error", in, err)
		}
	}
}

func TestParseYear(t *testing.T) {
	valid := map[string]int{"2024": 2024, "1999": 1999, " 2026 ": 2026}
	for in, want := range valid {
		got, err := ParseYear(in)
		if err != nil || got != want {
			t.Errorf("ParseYear(%q) = %d, %v; want %d", in, got, err, want)
		}
	}

	// Short years like "24" are rejected rather than guessed at: they are
	// almost always a typo for 2024, not the year 24.
	for _, in := range []string{"", "0", "24", "-2024", "abc", "20245", "2024.5"} {
		if _, err := ParseYear(in); err == nil || !strings.Contains(err.Error(), "invalid year") {
			t.Errorf("ParseYear(%q) error = %v; want an invalid year error", in, err)
		}
	}
}
