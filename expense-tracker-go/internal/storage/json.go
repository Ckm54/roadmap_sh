package storage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/ckm54/roadmap_sh/expense-tracker/internal/expense"
)

// EnvFile overrides the location of the expenses file.
const EnvFile = "EXPENSE_CLI_FILE"

func DefaultPath() (string, error) {
	if p := os.Getenv(EnvFile); p != "" {
		return p, nil
	}

	dir, err := dataDir()
	if err != nil {
		return "", fmt.Errorf("cannot locate data directory (set %s instead): %w", EnvFile, err)
	}
	return filepath.Join(dir, "expense-cli", "expenses.json"), nil
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

// JSONStore keeps all expenses in a single JSON file. It satisfies
// expense.DataStore, so the service never knows how data is persisted.
type JSONStore struct {
	path string
}

func NewJSONStore(path string) *JSONStore {
	return &JSONStore{path: path}
}

func (j *JSONStore) Load() (*expense.State, error) {
	data, err := os.ReadFile(j.path)
	if errors.Is(err, os.ErrNotExist) {
		return &expense.State{NextID: 1}, nil
	}
	if err != nil {
		return nil, err
	}

	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return &expense.State{NextID: 1}, nil
	}

	var state expense.State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("%s is not a valid expenses file: %w", j.path, err)
	}
	return &state, nil
}

// Save writes to a temporary file and renames it over the real one, so a
// crash mid-write can never leave a truncated expenses file behind.
func (j *JSONStore) Save(state *expense.State) error {
	if state.Expenses == nil {
		state.Expenses = []expense.ExpenseEntity{}
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
