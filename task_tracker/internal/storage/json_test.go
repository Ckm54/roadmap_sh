package storage

import (
	"path/filepath"
	"testing"

	"github.com/ckm54/task_tracker/internal/task"
)

func TestJSONStore_SaveAndLoad(t *testing.T) {
	tests := []struct {
		name         string
		tasksToSave  []task.TaskEntity
		expectLength int
	}{
		{
			name:         "Empty slice saves and loads array successfully",
			tasksToSave:  []task.TaskEntity{},
			expectLength: 0,
		},
		{
			name: "Populated data marshals correctly",
			tasksToSave: []task.TaskEntity{
				{ID: 1, Title: "Write Tests", Status: "in-progress"},
			},
			expectLength: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create an isolated sandboxed folder on your OS for this test run
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test_tasks.json")

			store := NewJSONStore(tmpFile)

			// Step 1: Write data to disk
			err := store.Save(tt.tasksToSave)
			if err != nil {
				t.Fatalf("Failed to save: %v", err)
			}

			// Step 2: Read it back into execution memory
			loadedTasks, err := store.Load()
			if err != nil {
				t.Fatalf("Failed to load: %v", err)
			}

			if len(loadedTasks) != tt.expectLength {
				t.Errorf("Expected %d tasks, loaded %d", tt.expectLength, len(loadedTasks))
			}
		})
	}
}
