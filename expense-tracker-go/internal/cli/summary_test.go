package cli

import (
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ckm54/roadmap_sh/expense-tracker/internal/expense"
)

// summaryExpenses spans two years, so year and month filters both matter.
func summaryExpenses() expense.State {
	return expense.State{NextID: 5, Expenses: []expense.ExpenseEntity{
		{ID: 1, Description: "Lunch", Amount: 2000, CreatedAt: day(2024, 8, 6)},
		{ID: 2, Description: "Dinner", Amount: 1000, CreatedAt: day(2024, 7, 6)},
		{ID: 3, Description: "Old lunch", Amount: 500, CreatedAt: day(2023, 8, 6)},
		{ID: 4, Description: "Old gift", Amount: 250, CreatedAt: day(2023, 12, 24)},
	}}
}

func TestSummary(t *testing.T) {
	// Every case passes --year when it filters, so none depend on today's date.
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"everything", []string{"summary"}, "Total expenses: $37.50"},
		{"one year", []string{"summary", "--year", "2024"}, "Total expenses for 2024: $30.00"},
		{"another year", []string{"summary", "--year", "2023"}, "Total expenses for 2023: $7.50"},
		{"month of a year", []string{"summary", "--month", "8", "--year", "2023"}, "Total expenses for August 2023: $5.00"},
		{"flag order does not matter", []string{"summary", "--year", "2024", "--month", "8"}, "Total expenses for August 2024: $20.00"},
		{"year with nothing spent", []string{"summary", "--year", "2019"}, "Total expenses for 2019: $0.00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, store, out := newTestApp()
			store.state = summaryExpenses()

			if err := app.Run(tt.args); err != nil {
				t.Fatalf("%v: %v", tt.args, err)
			}
			if got := out.String(); !strings.Contains(got, tt.want) {
				t.Errorf("output = %q, want it to contain %q", got, tt.want)
			}
		})
	}
}

func TestSummaryEmpty(t *testing.T) {
	app, _, out := newTestApp()

	if err := app.Run([]string{"summary"}); err != nil {
		t.Fatalf("summary: %v", err)
	}

	if got := out.String(); !strings.Contains(got, "Total expenses: $0.00") {
		t.Errorf("output = %q, want a zero total", got)
	}
}

// --month on its own means "this month of the current year", so this test
// uses the real clock and places expenses relative to today: one this month,
// one last month, one this month last year. Starting from the 1st means
// AddDate(0, -1, 0) can never roll over into the same month.
func TestSummaryMonthUsesCurrentYear(t *testing.T) {
	now := time.Now()
	thisMonth := time.Date(now.Year(), now.Month(), 1, 12, 0, 0, 0, time.Local)

	app, store, out := newTestApp()
	store.state = expense.State{NextID: 4, Expenses: []expense.ExpenseEntity{
		{ID: 1, Description: "This month", Amount: 2000, CreatedAt: thisMonth},
		{ID: 2, Description: "Last month", Amount: 500, CreatedAt: thisMonth.AddDate(0, -1, 0)},
		{ID: 3, Description: "A year ago", Amount: 700, CreatedAt: thisMonth.AddDate(-1, 0, 0)},
	}}

	month := now.Month()
	if err := app.Run([]string{"summary", "--month", strconv.Itoa(int(month))}); err != nil {
		t.Fatalf("summary --month: %v", err)
	}

	want := "Total expenses for " + month.String() + ": $20.00"
	if got := out.String(); !strings.Contains(got, want) {
		t.Errorf("output = %q, want it to contain %q", got, want)
	}
}

func TestSummaryErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"month zero", []string{"summary", "--month", "0"}, "invalid month"},
		{"month too large", []string{"summary", "--month", "13"}, "invalid month"},
		{"month not a number", []string{"summary", "--month", "august"}, "invalid month"},
		{"missing month value", []string{"summary", "--month"}, "flag needs an argument"},
		{"year not a number", []string{"summary", "--year", "last"}, "invalid year"},
		{"short year", []string{"summary", "--year", "24"}, "invalid year"},
		{"missing year value", []string{"summary", "--year"}, "flag needs an argument"},
		{"stray argument", []string{"summary", "extra"}, "unexpected argument"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, _, _ := newTestApp()

			err := app.Run(tt.args)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("err = %v, want it to contain %q", err, tt.want)
			}
		})
	}
}

func TestSummaryIsRegistered(t *testing.T) {
	app, _, out := newTestApp()

	if err := app.Run([]string{"help"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "summary") {
		t.Errorf("help output does not mention summary:\n%s", out)
	}

	out.Reset()
	if err := app.Run([]string{"summary", "--help"}); err != nil {
		t.Fatalf("summary --help: %v", err)
	}
	for _, flag := range []string{"-month", "-year"} {
		if !strings.Contains(out.String(), flag) {
			t.Errorf("summary --help output = %q, want the %s flag listed", out, flag)
		}
	}
}
