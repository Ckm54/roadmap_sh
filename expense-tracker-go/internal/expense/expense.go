package expense

import "time"

const DateLayout = "2006-01-02"

type ExpenseEntity struct {
	ID          int       `json:"id"`
	Description string    `json:"description"`
	Amount      Cents     `json:"amountCents"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// State is everything a datastore persists. NextID only grows
type State struct {
	NextID   int             `json:"nextId"`
	Expenses []ExpenseEntity `json:"expenses"`
}

type DataStore interface {
	Load() (*State, error)
	Save(*State) error
}
