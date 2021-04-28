package task

import (
	"fmt"
	"testing"
	"time"
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

func TestService_Update(t *testing.T) {
	oldTime := time.Now().Add(-1 * time.Hour)

	tests := []struct {
		name          string
		initialTasks  []TaskEntity
		inputID       int
		inputTitle    string
		wantErr       bool
		expectedTitle string
	}{
		{
			name:          "Successfully updates an existing task",
			initialTasks:  []TaskEntity{{ID: 1, Title: "Old Title", Status: StatusTodo, CreatedAt: oldTime, UpdatedAt: oldTime}},
			inputID:       1,
			inputTitle:    "New Title",
			wantErr:       false,
			expectedTitle: "New Title",
		},
		{
			name:         "Fails if task ID does not exist",
			initialTasks: []TaskEntity{{ID: 1, Title: "Old Title", Status: StatusTodo, CreatedAt: oldTime, UpdatedAt: oldTime}},
			inputID:      42,
			inputTitle:   "New Title",
			wantErr:      true,
		},
		{
			name:         "Fails if store is unavailable",
			initialTasks: nil,
			inputID:      1,
			inputTitle:   "New Title",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memStore := &mockStore{tasks: tt.initialTasks, shouldFail: tt.initialTasks == nil}
			svc := NewService(memStore)

			err := svc.Update(tt.inputID, tt.inputTitle)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				updated := memStore.tasks[0]
				if updated.Title != tt.expectedTitle {
					t.Errorf("Expected title %q, got %q", tt.expectedTitle, updated.Title)
				}
				if !updated.UpdatedAt.After(oldTime) {
					t.Error("Expected UpdatedAt to be refreshed, but it was not")
				}
				if updated.CreatedAt != oldTime {
					t.Error("Expected CreatedAt to be unchanged")
				}
			}
		})
	}
}

func TestService_Delete(t *testing.T) {
	tests := []struct {
		name           string
		initialTasks   []TaskEntity
		inputID        int
		wantErr        bool
		wantTasksCount int
	}{
		{
			name: "Successfully deletes an existing task",
			initialTasks: []TaskEntity{
				{ID: 1, Title: "Buy groceries", Status: StatusTodo},
			},
			inputID:        1,
			wantErr:        false,
			wantTasksCount: 0,
		},
		{
			name: "Deletes the correct task when multiple exist",
			initialTasks: []TaskEntity{
				{ID: 1, Title: "Task one", Status: StatusTodo},
				{ID: 2, Title: "Task two", Status: StatusTodo},
				{ID: 3, Title: "Task three", Status: StatusTodo},
			},
			inputID:        2,
			wantErr:        false,
			wantTasksCount: 2,
		},
		{
			name: "Fails if task ID does not exist",
			initialTasks: []TaskEntity{
				{ID: 1, Title: "Buy groceries", Status: StatusTodo},
			},
			inputID: 42,
			wantErr: true,
		},
		{
			name:         "Fails if store is unavailable",
			initialTasks: nil,
			inputID:      1,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memStore := &mockStore{tasks: tt.initialTasks, shouldFail: tt.initialTasks == nil}
			svc := NewService(memStore)

			err := svc.Delete(tt.inputID)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if len(memStore.tasks) != tt.wantTasksCount {
					t.Errorf("Expected %d tasks remaining, got %d", tt.wantTasksCount, len(memStore.tasks))
				}

				for _, task := range memStore.tasks {
					if task.ID == tt.inputID {
						t.Errorf("Task with ID %d still present after deletion", tt.inputID)
					}
				}
			}
		})
	}
}
