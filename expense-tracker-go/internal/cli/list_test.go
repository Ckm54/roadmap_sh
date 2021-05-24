package cli

import (
	"strings"
	"testing"
	"time"

	"github.com/ckm54/roadmap_sh/expense-tracker/internal/expense"
)

// day returns noon UTC on the given date, so the printed date is the same in
// every local time zone the tests might run in.
func day(year int, month time.Month, d int) time.Time {
	return time.Date(year, month, d, 12, 0, 0, 0, time.UTC)
}

func TestList(t *testing.T) {
	app, store, out := newTestApp()
	store.state = expense.State{NextID: 4, Expenses: []expense.ExpenseEntity{
		{ID: 1, Description: "Lunch", Amount: 2000, CreatedAt: day(2024, 8, 6)},
		{ID: 3, Description: "Coffee beans", Amount: 10, CreatedAt: day(2024, 8, 7)},
	}}

	if err := app.Run([]string{"list"}); err != nil {
		t.Fatalf("list: %v", err)
	}

	lines := nonEmptyLines(out.String())
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want a header and 2 rows:\n%s", len(lines), out)
	}
	assertInOrder(t, lines[0], "#", "ID", "Date", "Description", "Amount")
	assertInOrder(t, lines[1], "#", "1", "2024-08-06", "Lunch", "$20.00")
	assertInOrder(t, lines[2], "#", "3", "2024-08-07", "Coffee beans", "$0.10")
}

func TestListColumnsAreAligned(t *testing.T) {
	app, store, out := newTestApp()
	store.state = expense.State{NextID: 3, Expenses: []expense.ExpenseEntity{
		{ID: 1, Description: "A", Amount: 100, CreatedAt: day(2024, 8, 6)},
		{ID: 2, Description: "A much longer description", Amount: 100, CreatedAt: day(2024, 8, 6)},
	}}

	if err := app.Run([]string{"list"}); err != nil {
		t.Fatalf("list: %v", err)
	}

	// Every row's Amount column must start where the header's does.
	lines := nonEmptyLines(out.String())
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want a header and 2 rows:\n%s", len(lines), out)
	}
	want := strings.Index(lines[0], "Amount")
	for _, line := range lines[1:] {
		if got := strings.Index(line, "$"); got != want {
			t.Errorf("amount starts at column %d, header at %d:\n%s", got, want, out)
		}
	}
}

func TestListSortsByID(t *testing.T) {
	app, store, out := newTestApp()
	store.state = expense.State{NextID: 3, Expenses: []expense.ExpenseEntity{
		{ID: 2, Description: "Second", Amount: 100, CreatedAt: day(2024, 8, 6)},
		{ID: 1, Description: "First", Amount: 100, CreatedAt: day(2024, 8, 6)},
	}}

	if err := app.Run([]string{"list"}); err != nil {
		t.Fatalf("list: %v", err)
	}

	got := out.String()
	if strings.Index(got, "First") > strings.Index(got, "Second") {
		t.Errorf("expenses not listed by ID:\n%s", got)
	}
}

func TestListEmpty(t *testing.T) {
	app, _, out := newTestApp()

	if err := app.Run([]string{"list"}); err != nil {
		t.Fatalf("list: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "No expenses found") {
		t.Errorf("output = %q, want a 'No expenses found' message", got)
	}
	if strings.Contains(got, "Description") {
		t.Errorf("output = %q, should not print a table header when empty", got)
	}
}

func TestListRejectsArguments(t *testing.T) {
	app, _, _ := newTestApp()

	err := app.Run([]string{"list", "extra"})
	if err == nil || !strings.Contains(err.Error(), "unexpected argument") {
		t.Fatalf("err = %v, want an unexpected argument error", err)
	}
}

func TestListIsRegistered(t *testing.T) {
	app, _, out := newTestApp()

	if err := app.Run([]string{"help"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "list") {
		t.Errorf("help output does not mention list:\n%s", out)
	}

	out.Reset()
	if err := app.Run([]string{"list", "--help"}); err != nil {
		t.Fatalf("list --help: %v", err)
	}
	if !strings.Contains(out.String(), "Usage:") {
		t.Errorf("list --help output = %q, want usage text", out)
	}
}

func nonEmptyLines(s string) []string {
	var lines []string
	for line := range strings.Lines(s) {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, strings.TrimRight(line, "\n"))
		}
	}
	return lines
}

// assertInOrder checks that every part appears in line, left to right. It
// doesn't care about exact spacing, so the tests survive cosmetic changes.
func assertInOrder(t *testing.T, line string, parts ...string) {
	t.Helper()
	rest := line
	for _, p := range parts {
		i := strings.Index(rest, p)
		if i < 0 {
			t.Errorf("line %q: missing %q (or out of order); want %q", line, p, parts)
			return
		}
		rest = rest[i+len(p):]
	}
}
