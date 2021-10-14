package storage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/ckm54/task_tracker/internal/task"
)

// EnvFile overrides the location of the tasks file.
const EnvFile = "TASK_CLI_FILE"

// DefaultPath returns $TASK_CLI_FILE if set, otherwise tasks.json inside the
// per-user data directory (e.g. ~/.local/share/task-cli on Linux).
func DefaultPath() (string, error) {
	if p := os.Getenv(EnvFile); p != "" {
		return p, nil
	}

	dir, err := dataDir()
	if err != nil {
		return "", fmt.Errorf("cannot locate data directory (set %s instead): %w", EnvFile, err)
	}
	return filepath.Join(dir, "task-cli", "tasks.json"), nil
}

func dataDir() (string, error) {
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return dir, nil
	}
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		return os.UserConfigDir()
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share"), nil
}

type JSONStore struct {
	path string
}

func NewJSONStore(path string) *JSONStore {
	return &JSONStore{path: path}
}

func (j *JSONStore) Path() string { return j.path }

func (j *JSONStore) Load() (*task.State, error) {
	data, err := os.ReadFile(j.path)
	if errors.Is(err, os.ErrNotExist) {
		return &task.State{NextID: 1}, nil
	}
	if err != nil {
		return nil, err
	}

	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return &task.State{NextID: 1}, nil
	}

	if data[0] == '[' {
		return loadLegacy(data)
	}

	var state task.State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("%s is not a valid tasks file: %w", j.path, err)
	}
	return &state, nil
}

// legacyTask accepts the original file format: a bare JSON array using
// "title", "created_at" and "updated_at". It is rewritten in the current
// format on the next save.
type legacyTask struct {
	task.TaskEntity
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func loadLegacy(data []byte) (*task.State, error) {
	var legacy []legacyTask
	if err := json.Unmarshal(data, &legacy); err != nil {
		return nil, fmt.Errorf("invalid legacy tasks file: %w", err)
	}

	state := &task.State{Tasks: make([]task.TaskEntity, 0, len(legacy))}
	for _, l := range legacy {
		t := l.TaskEntity
		if t.Description == "" {
			t.Description = l.Title
		}
		if t.CreatedAt.IsZero() {
			t.CreatedAt = l.CreatedAt
		}
		if t.UpdatedAt.IsZero() {
			t.UpdatedAt = l.UpdatedAt
		}
		state.Tasks = append(state.Tasks, t)
	}
	return state, nil
}

// Save writes to a temporary file and renames it over the real one, so a
// crash mid-write can never leave a truncated tasks file behind.
func (j *JSONStore) Save(state *task.State) error {
	if state.Tasks == nil {
		state.Tasks = []task.TaskEntity{}
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(j.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, filepath.Base(j.path)+".*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(append(data, '\n')); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	return os.Rename(tmp.Name(), j.path)
}
