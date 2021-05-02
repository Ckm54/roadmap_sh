package task

import "time"

type TaskEntity struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DataStore interface {
	Load() ([]TaskEntity, error)
	Save([]TaskEntity) error
}

func (t *TaskEntity) String() string {
	return t.Title
}
