package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ckm54/roadmap_sh/expense-tracker/internal/expense"
)

// memStore is an in-memory expense.DataStore so tests never touch disk.
type memStore struct{ state expense.State }

func (m *memStore) Load() (*expense.State, error) {
	s := m.state
	s.Expenses = append([]expense.ExpenseEntity(nil), m.state.Expenses...)
	return &s, nil
}

func (m *memStore) Save(s *expense.State) error { m.state = *s; return nil }

func newTestApp() (*App, *memStore, *bytes.Buffer) {
	store := &memStore{}
	var out bytes.Buffer
	return New(expense.NewService(store), &out), store, &out
}

func TestAdd(t *testing.T) {
	app, store, out := newTestApp()

	if err := app.Run([]string{"add", "--description", "  Lunch ", "--amount", "20"}); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := app.Run([]string{"add", "--description=Dinner", "--amount=10.5"}); err != nil {
		t.Fatalf("add: %v", err)
	}

	if got := out.String(); !strings.Contains(got, "(ID: 1)") || !strings.Contains(got, "(ID: 2)") {
		t.Errorf("output = %q, want IDs 1 and 2", got)
	}
	if len(store.state.Expenses) != 2 {
		t.Fatalf("stored %d expenses, want 2", len(store.state.Expenses))
	}
	if e := store.state.Expenses[0]; e.Description != "Lunch" || e.Amount != 2000 {
		t.Errorf("first expense = %+v", e)
	}
	if e := store.state.Expenses[1]; e.Amount != 1050 {
		t.Errorf("second expense amount = %d cents, want 1050", e.Amount)
	}
}

func TestAddErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"missing description", []string{"add", "--amount", "5"}, "description cannot be empty"},
		{"missing amount", []string{"add", "--description", "Lunch"}, "amount must be"},
		{"negative amount", []string{"add", "--description", "Lunch", "--amount", "-3"}, "invalid amount"},
		{"zero amount", []string{"add", "--description", "Lunch", "--amount", "0.00"}, "amount must be greater than 0"},
		{"too many decimals", []string{"add", "--description", "Lunch", "--amount", "1.005"}, "at most 2 decimal places"},
		{"amount not a number", []string{"add", "--description", "Lunch", "--amount", "abc"}, "invalid amount"},
		{"unknown flag", []string{"add", "--colour", "red"}, "flag provided but not defined"},
		{"stray argument", []string{"add", "--description", "Lunch", "--amount", "5", "extra"}, "unexpected argument"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app, store, _ := newTestApp()

			err := app.Run(tt.args)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("err = %v, want it to contain %q", err, tt.want)
			}
			if len(store.state.Expenses) != 0 {
				t.Error("a failed add must not save anything")
			}
		})
	}
}
