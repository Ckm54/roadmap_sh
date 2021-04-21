package storage

import (
	"encoding/json"
	"os"

	"github.com/ckm54/task_tracker/internal/task"
)

type JSONStore struct {
	filepath string
}

func NewJSONStore(filepath string) *JSONStore {
	store := &JSONStore{filepath: filepath}

	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		emptyData := []byte("[]")
		_ = os.WriteFile(filepath, emptyData, 0644)
	}

	return store
}

func (j *JSONStore) Load() ([]task.TaskEntity, error) {
	if _, err := os.Stat(j.filepath); os.IsNotExist(err) {
		return []task.TaskEntity{}, nil
	}

	data, err := os.ReadFile(j.filepath)
	if err != nil {
		return nil, err
	}

	var tasks []task.TaskEntity
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (j *JSONStore) Save(tasks []task.TaskEntity) error {
	data, err := json.MarshalIndent(tasks, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(j.filepath, data, 0644)
}
