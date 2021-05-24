package expense

import (
	"errors"
	"testing"
)

// memStore is an in-memory DataStore so tests never touch disk.
type memStore struct {
	state   State
	loadErr error
}

func (m *memStore) Load() (*State, error) {
	if m.loadErr != nil {
		return nil, m.loadErr
	}
	s := m.state
	s.Expenses = append([]ExpenseEntity(nil), m.state.Expenses...)
	return &s, nil
}

func (m *memStore) Save(s *State) error { m.state = *s; return nil }

func TestServiceList(t *testing.T) {
	store := &memStore{state: State{NextID: 4, Expenses: []ExpenseEntity{
		{ID: 3, Description: "Dinner", Amount: 1000},
		{ID: 1, Description: "Lunch", Amount: 2000},
	}}}

	got, err := NewService(store).List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(got) != 2 || got[0].ID != 1 || got[1].ID != 3 {
		t.Fatalf("List() = %+v, want IDs [1 3]", got)
	}
}

func TestServiceListEmpty(t *testing.T) {
	got, err := NewService(&memStore{}).List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("List() = %+v, want no expenses", got)
	}
}

func TestServiceListDoesNotChangeTheStore(t *testing.T) {
	store := &memStore{state: State{NextID: 2, Expenses: []ExpenseEntity{
		{ID: 1, Description: "Lunch", Amount: 2000},
	}}}

	got, err := NewService(store).List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	got[0].Description = "changed by caller"

	if store.state.Expenses[0].Description != "Lunch" {
		t.Error("editing the returned slice must not change stored expenses")
	}
}

func TestServiceListLoadError(t *testing.T) {
	boom := errors.New("disk on fire")

	_, err := NewService(&memStore{loadErr: boom}).List()
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want it to wrap %v", err, boom)
	}
}
