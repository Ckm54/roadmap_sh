package task

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
)

var ErrNotFound = errors.New("task not found")

// NotFoundError lists every requested ID that does not exist. It matches
// ErrNotFound with errors.Is.
type NotFoundError struct {
	IDs []int
}

func (e *NotFoundError) Error() string {
	if len(e.IDs) == 1 {
		return fmt.Sprintf("task with ID %d not found", e.IDs[0])
	}
	ids := make([]string, len(e.IDs))
	for i, id := range e.IDs {
		ids[i] = fmt.Sprint(id)
	}
	return fmt.Sprintf("tasks with IDs %s not found", strings.Join(ids, ", "))
}

func (e *NotFoundError) Is(target error) bool { return target == ErrNotFound }

type AddOptions struct {
	Priority Priority
	Due      *time.Time
	Tags     []string
}

// Patch describes a partial update. Nil fields are left unchanged.
type Patch struct {
	Description *string
	Priority    *Priority
	Due         *time.Time
	ClearDue    bool
	Tags        *[]string
}

type SortField string

const (
	SortID       SortField = "id"
	SortPriority SortField = "priority"
	SortDue      SortField = "due"
	SortCreated  SortField = "created"
	SortUpdated  SortField = "updated"
)

var SortFields = []SortField{SortID, SortPriority, SortDue, SortCreated, SortUpdated}

func ParseSortField(s string) (SortField, error) {
	for _, f := range SortFields {
		if string(f) == s {
			return f, nil
		}
	}
	return "", fmt.Errorf("unknown sort field %q (want id, priority, due, created or updated)", s)
}

// ListOptions filters and orders List results. Zero values mean "no filter".
type ListOptions struct {
	Status   Status
	Priority Priority
	Tag      string
	Query    string
	SortBy   SortField
}

type Service struct {
	store DataStore
	now   func() time.Time
}

func NewService(store DataStore) *Service {
	return &Service{store: store, now: time.Now}
}

func (s *Service) load() (*State, error) {
	state, err := s.store.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load tasks: %w", err)
	}
	for i, t := range state.Tasks {
		state.NextID = max(state.NextID, t.ID+1)
		if t.Priority == "" {
			state.Tasks[i].Priority = PriorityMedium
		}
	}
	state.NextID = max(state.NextID, 1)
	return state, nil
}

func (s *Service) save(state *State) error {
	if err := s.store.Save(state); err != nil {
		return fmt.Errorf("failed to save tasks: %w", err)
	}
	return nil
}

func (s *Service) Add(description string, opts AddOptions) (TaskEntity, error) {
	state, err := s.load()
	if err != nil {
		return TaskEntity{}, err
	}

	priority := opts.Priority
	if priority == "" {
		priority = PriorityMedium
	}

	now := s.now()
	newTask := TaskEntity{
		ID:          state.NextID,
		Description: description,
		Status:      StatusTodo,
		Priority:    priority,
		Due:         opts.Due,
		Tags:        NormalizeTags(opts.Tags),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	state.NextID++
	state.Tasks = append(state.Tasks, newTask)

	return newTask, s.save(state)
}

func (s *Service) Get(id int) (TaskEntity, error) {
	state, err := s.load()
	if err != nil {
		return TaskEntity{}, err
	}

	i := indexOf(state.Tasks, id)
	if i < 0 {
		return TaskEntity{}, &NotFoundError{IDs: []int{id}}
	}
	return state.Tasks[i], nil
}

func (s *Service) Update(id int, patch Patch) error {
	state, err := s.load()
	if err != nil {
		return err
	}

	i := indexOf(state.Tasks, id)
	if i < 0 {
		return &NotFoundError{IDs: []int{id}}
	}

	t := &state.Tasks[i]
	if patch.Description != nil {
		t.Description = *patch.Description
	}
	if patch.Priority != nil {
		t.Priority = *patch.Priority
	}
	if patch.ClearDue {
		t.Due = nil
	} else if patch.Due != nil {
		t.Due = patch.Due
	}
	if patch.Tags != nil {
		t.Tags = NormalizeTags(*patch.Tags)
	}
	t.UpdatedAt = s.now()

	return s.save(state)
}

// SetStatus updates every given task, or none of them if any ID is missing.
func (s *Service) SetStatus(ids []int, status Status) error {
	state, err := s.load()
	if err != nil {
		return err
	}

	indexes, err := indexesOf(state.Tasks, ids)
	if err != nil {
		return err
	}

	now := s.now()
	for _, i := range indexes {
		state.Tasks[i].Status = status
		state.Tasks[i].UpdatedAt = now
	}

	return s.save(state)
}

// Delete removes every given task, or none of them if any ID is missing.
func (s *Service) Delete(ids []int) error {
	state, err := s.load()
	if err != nil {
		return err
	}

	if _, err := indexesOf(state.Tasks, ids); err != nil {
		return err
	}

	state.Tasks = slices.DeleteFunc(state.Tasks, func(t TaskEntity) bool {
		return slices.Contains(ids, t.ID)
	})

	return s.save(state)
}

// Clear deletes all tasks with the given status and returns how many were removed.
func (s *Service) Clear(status Status) (int, error) {
	state, err := s.load()
	if err != nil {
		return 0, err
	}

	before := len(state.Tasks)
	state.Tasks = slices.DeleteFunc(state.Tasks, func(t TaskEntity) bool {
		return t.Status == status
	})
	removed := before - len(state.Tasks)

	if removed == 0 {
		return 0, nil
	}
	return removed, s.save(state)
}

func (s *Service) List(opts ListOptions) ([]TaskEntity, error) {
	state, err := s.load()
	if err != nil {
		return nil, err
	}

	query := strings.ToLower(opts.Query)
	tag := strings.ToLower(opts.Tag)

	tasks := slices.DeleteFunc(slices.Clone(state.Tasks), func(t TaskEntity) bool {
		switch {
		case opts.Status != "" && t.Status != opts.Status:
			return true
		case opts.Priority != "" && t.Priority != opts.Priority:
			return true
		case tag != "" && !slices.Contains(t.Tags, tag):
			return true
		case query != "" && !strings.Contains(strings.ToLower(t.Description), query):
			return true
		}
		return false
	})

	slices.SortStableFunc(tasks, compareBy(opts.SortBy))
	return tasks, nil
}

func compareBy(field SortField) func(a, b TaskEntity) int {
	byID := func(a, b TaskEntity) int { return a.ID - b.ID }

	switch field {
	case SortPriority:
		return func(a, b TaskEntity) int {
			if c := a.Priority.Rank() - b.Priority.Rank(); c != 0 {
				return c
			}
			return byID(a, b)
		}
	case SortDue:
		// Tasks without a due date go last.
		return func(a, b TaskEntity) int {
			switch {
			case a.Due == nil && b.Due == nil:
				return byID(a, b)
			case a.Due == nil:
				return 1
			case b.Due == nil:
				return -1
			}
			if c := a.Due.Compare(*b.Due); c != 0 {
				return c
			}
			return byID(a, b)
		}
	case SortCreated:
		return func(a, b TaskEntity) int {
			if c := a.CreatedAt.Compare(b.CreatedAt); c != 0 {
				return c
			}
			return byID(a, b)
		}
	case SortUpdated:
		return func(a, b TaskEntity) int {
			if c := a.UpdatedAt.Compare(b.UpdatedAt); c != 0 {
				return c
			}
			return byID(a, b)
		}
	default:
		return byID
	}
}

// NormalizeTags lowercases, trims and de-duplicates tags, dropping empty ones.
func NormalizeTags(tags []string) []string {
	var out []string
	for _, tag := range tags {
		tag = strings.ToLower(strings.TrimSpace(tag))
		if tag != "" && !slices.Contains(out, tag) {
			out = append(out, tag)
		}
	}
	return out
}

func indexOf(tasks []TaskEntity, id int) int {
	return slices.IndexFunc(tasks, func(t TaskEntity) bool { return t.ID == id })
}

func indexesOf(tasks []TaskEntity, ids []int) ([]int, error) {
	var indexes, missing []int
	for _, id := range ids {
		if i := indexOf(tasks, id); i >= 0 {
			indexes = append(indexes, i)
		} else {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		return nil, &NotFoundError{IDs: missing}
	}
	return indexes, nil
}
