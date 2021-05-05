package task

import (
	"fmt"
	"iter"
	"slices"
	"time"

	"github.com/ckm54/task_tracker/internal/constants"
)

type Service struct {
	store DataStore
}

func NewService(store DataStore) *Service {
	return &Service{store: store}
}

func (s *Service) Add(title string) (int, error) {
	tasks, err := s.store.Load()
	if err != nil {
		return 0, fmt.Errorf("failed to load tasks: %w", err)
	}

	nextID := 1
	if len(tasks) > 0 {
		nextID = tasks[len(tasks)-1].ID + 1
	}

	now := time.Now()
	newTask := TaskEntity{
		ID:        nextID,
		Title:     title,
		Status:    constants.StatusTodo,
		CreatedAt: now,
		UpdatedAt: now,
	}

	tasks = append(tasks, newTask)

	return nextID, s.store.Save(tasks)
}

func (s *Service) Update(id int, title string) error {
	tasks, err := s.store.Load()
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}

	for i, t := range tasks {
		if t.ID == id {
			tasks[i].Title = title
			tasks[i].UpdatedAt = time.Now()
			return s.store.Save(tasks)
		}
	}

	return fmt.Errorf("task with ID %d not found", id)
}

func (s *Service) UpdateStatus(id int, status string) error {
	tasks, err := s.store.Load()
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}

	for i, t := range tasks {
		if t.ID == id {
			tasks[i].Status = status
			tasks[i].UpdatedAt = time.Now()
			return s.store.Save(tasks)
		}
	}

	return fmt.Errorf("task with ID %d not found", id)
}

func (s *Service) List(filter string) ([]TaskEntity, error) {
	tasks, err := s.store.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}

	if filter != "" {
		filteredTasks := filterTasks(tasks, func(t TaskEntity) bool { return t.Status == filter })
		tasks = slices.Collect(filteredTasks)
	}

	return tasks, nil
}

func (s *Service) Delete(id int) error {
	tasks, err := s.store.Load()
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}

	for i, t := range tasks {
		if t.ID == id {
			tasks = append(tasks[:i], tasks[i+1:]...)
			return s.store.Save(tasks)
		}
	}

	return fmt.Errorf("task with ID %d not found", id)
}

func filterTasks[T any](slice []T, predicate func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, item := range slice {
			if predicate(item) {
				if !yield(item) {
					return
				}
			}
		}
	}
}
