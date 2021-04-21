package task

import (
	"fmt"
	"time"
)

type Service struct {
	store DataStore
}

func NewService(store DataStore) *Service {
	return &Service{store: store}
}

func (s *Service) Add(title string) error {
	tasks, err := s.store.Load()
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}

	nextID := 1
	if len(tasks) > 0 {
		nextID = tasks[len(tasks)-1].ID + 1
	}

	now := time.Now()
	newTask := TaskEntity{
		ID:        nextID,
		Title:     title,
		Status:    StatusTodo,
		CreatedAt: now,
		UpdatedAt: now,
	}

	tasks = append(tasks, newTask)

	return s.store.Save(tasks)
}
