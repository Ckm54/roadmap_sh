package task

import (
	"fmt"
	"time"
)

type Status string

const (
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in-progress"
	StatusDone       Status = "done"
)

var Statuses = []Status{StatusTodo, StatusInProgress, StatusDone}

func ParseStatus(s string) (Status, error) {
	for _, st := range Statuses {
		if string(st) == s {
			return st, nil
		}
	}
	return "", fmt.Errorf("unknown status %q (want todo, in-progress or done)", s)
}

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

var Priorities = []Priority{PriorityLow, PriorityMedium, PriorityHigh}

func ParsePriority(s string) (Priority, error) {
	for _, p := range Priorities {
		if string(p) == s {
			return p, nil
		}
	}
	return "", fmt.Errorf("unknown priority %q (want low, medium or high)", s)
}

// Rank orders priorities so that higher priorities sort first.
func (p Priority) Rank() int {
	switch p {
	case PriorityHigh:
		return 0
	case PriorityLow:
		return 2
	default:
		return 1
	}
}

const DateLayout = "2006-01-02"

type TaskEntity struct {
	ID          int        `json:"id"`
	Description string     `json:"description"`
	Status      Status     `json:"status"`
	Priority    Priority   `json:"priority,omitempty"`
	Due         *time.Time `json:"dueDate,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// Overdue reports whether an unfinished task's due date is before today.
func (t TaskEntity) Overdue(now time.Time) bool {
	if t.Due == nil || t.Status == StatusDone {
		return false
	}
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return t.Due.Before(today)
}

// State is everything a DataStore persists. NextID only ever grows, so IDs
// of deleted tasks are never handed out again.
type State struct {
	NextID int          `json:"nextId"`
	Tasks  []TaskEntity `json:"tasks"`
}

type DataStore interface {
	Load() (*State, error)
	Save(*State) error
}
