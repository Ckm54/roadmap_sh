package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ckm54/task_tracker/internal/task"
)

func TestJSONStore_SaveAndLoad(t *testing.T) {
	due := time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name  string
		state task.State
	}{
		{
			name:  "empty state",
			state: task.State{NextID: 1, Tasks: []task.TaskEntity{}},
		},
		{
			name: "populated state keeps every field",
			state: task.State{NextID: 5, Tasks: []task.TaskEntity{
				{ID: 4, Description: "Write tests", Status: task.StatusInProgress, Priority: task.PriorityHigh, Due: &due, Tags: []string{"work"}},
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// A nested path also checks that missing directories are created.
			store := NewJSONStore(filepath.Join(t.TempDir(), "nested", "tasks.json"))

			if err := store.Save(&tt.state); err != nil {
				t.Fatalf("Save: %v", err)
			}

			loaded, err := store.Load()
			if err != nil {
				t.Fatalf("Load: %v", err)
			}

			if loaded.NextID != tt.state.NextID || len(loaded.Tasks) != len(tt.state.Tasks) {
				t.Fatalf("loaded %+v, want %+v", loaded, tt.state)
			}
			for i, want := range tt.state.Tasks {
				got := loaded.Tasks[i]
				if got.ID != want.ID || got.Description != want.Description || got.Status != want.Status ||
					got.Priority != want.Priority || !got.Due.Equal(*want.Due) || strings.Join(got.Tags, ",") != strings.Join(want.Tags, ",") {
					t.Errorf("task %d: got %+v, want %+v", i, got, want)
				}
			}
		})
	}
}

func TestJSONStore_LoadMissingOrEmptyFile(t *testing.T) {
	dir := t.TempDir()

	empty := filepath.Join(dir, "empty.json")
	if err := os.WriteFile(empty, []byte("  \n"), 0o644); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{filepath.Join(dir, "missing.json"), empty} {
		state, err := NewJSONStore(path).Load()
		if err != nil {
			t.Fatalf("Load(%s): %v", filepath.Base(path), err)
		}
		if state.NextID != 1 || len(state.Tasks) != 0 {
			t.Errorf("Load(%s) = %+v, want empty state", filepath.Base(path), state)
		}
	}

	if _, err := os.Stat(filepath.Join(dir, "missing.json")); !os.IsNotExist(err) {
		t.Error("Load must not create the file")
	}
}

func TestJSONStore_LoadInvalidFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := NewJSONStore(path).Load(); err == nil {
		t.Fatal("expected an error for invalid JSON")
	}
}

func TestJSONStore_LoadLegacyFormat(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	legacy := `[
 {"id": 1, "title": "old task", "status": "done", "created_at": "2026-10-07T00:27:02Z", "updated_at": "2026-10-08T00:00:00Z"},
 {"id": 3, "title": "another", "status": "todo", "created_at": "2026-10-07T00:27:15Z", "updated_at": "2026-10-07T00:27:15Z"}
]`
	if err := os.WriteFile(path, []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}

	state, err := NewJSONStore(path).Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if len(state.Tasks) != 2 {
		t.Fatalf("got %d tasks, want 2", len(state.Tasks))
	}
	first := state.Tasks[0]
	if first.Description != "old task" || first.Status != task.StatusDone {
		t.Errorf("unexpected task: %+v", first)
	}
	if want := time.Date(2026, 10, 7, 0, 27, 2, 0, time.UTC); !first.CreatedAt.Equal(want) {
		t.Errorf("CreatedAt = %v, want %v", first.CreatedAt, want)
	}
	if want := time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC); !first.UpdatedAt.Equal(want) {
		t.Errorf("UpdatedAt = %v, want %v", first.UpdatedAt, want)
	}
}

func TestJSONStore_SaveLeavesNoTempFiles(t *testing.T) {
	dir := t.TempDir()
	store := NewJSONStore(filepath.Join(dir, "tasks.json"))

	for range 3 {
		if err := store.Save(&task.State{NextID: 1}); err != nil {
			t.Fatal(err)
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "tasks.json" {
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf("directory contains %v, want only tasks.json", names)
	}
}

func TestDefaultPath(t *testing.T) {
	t.Run("env var wins", func(t *testing.T) {
		t.Setenv(EnvFile, "/custom/tasks.json")
		got, err := DefaultPath()
		if err != nil || got != "/custom/tasks.json" {
			t.Errorf("DefaultPath() = %q, %v", got, err)
		}
	})

	t.Run("uses XDG_DATA_HOME", func(t *testing.T) {
		t.Setenv(EnvFile, "")
		t.Setenv("XDG_DATA_HOME", "/xdg")
		got, err := DefaultPath()
		if err != nil || got != filepath.Join("/xdg", "task-cli", "tasks.json") {
			t.Errorf("DefaultPath() = %q, %v", got, err)
		}
	})
}
