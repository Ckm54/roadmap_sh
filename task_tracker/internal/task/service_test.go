package task

import (
	"fmt"
	"testing"
)

type mockStore struct {
	tasks      []TaskEntity
	shouldFail bool
}

func (m *mockStore) Load() ([]TaskEntity, error) {
	if m.shouldFail {
		return nil, fmt.Errorf("simulated DB error")
	}

	return m.tasks, nil
}

func (m *mockStore) Save(tasks []TaskEntity) error {
	m.tasks = tasks
	return nil
}

func TestService_Add(t *testing.T) {
	tests := []struct {
		name         string
		initialTasks []TaskEntity
		inputTitle   string
		wantID       int
		wantStatus   string
		wantErr      bool
	}{
		{
			name:         "Add task to an empty tracker",
			initialTasks: []TaskEntity{},
			inputTitle:   "Buy groceries",
			wantID:       1,
			wantStatus:   "todo",
			wantErr:      false,
		},
		{
			name: "IDs increment sequentially",
			initialTasks: []TaskEntity{
				{ID: 1, Title: "Existing Task", Status: "todo"},
			},
			inputTitle: "New Task",
			wantID:     2,
			wantStatus: "todo",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memStore := &mockStore{tasks: tt.initialTasks}
			svc := NewService(memStore)

			err := svc.Add(tt.inputTitle)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Add() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				savedTasks := memStore.tasks
				newCreatedTask := savedTasks[len(savedTasks)-1]

				if newCreatedTask.ID != tt.wantID {
					t.Errorf("Expected ID %d, got %d", tt.wantID, newCreatedTask.ID)
				}
				if newCreatedTask.Title != tt.inputTitle {
					t.Errorf("Expected Title %q, got %q", tt.inputTitle, newCreatedTask.Title)
				}
				if newCreatedTask.Status != tt.wantStatus {
					t.Errorf("Expected Status %q, got %q", tt.wantStatus, newCreatedTask.Status)
				}
				if newCreatedTask.CreatedAt.IsZero() || newCreatedTask.UpdatedAt.IsZero() {
					t.Error("Expected timestamps to be populated, but they were empty")
				}
			}
		})
	}
}
