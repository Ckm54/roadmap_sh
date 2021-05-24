package expense

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

var ErrNotFound = errors.New("expense not found")

// NotFoundErr lists every requested ID that does not exist
type NotFoundErr struct {
	IDs []int
}

func (e *NotFoundErr) Error() string {
	if len(e.IDs) == 1 {
		return fmt.Sprintf("expense with ID %d not found", e.IDs[0])
	}

	ids := make([]string, len(e.IDs))
	for i, id := range e.IDs {
		ids[i] = fmt.Sprint(id)
	}
	return fmt.Sprintf("expenses with IDs %s not found", strings.Join(ids, ", "))
}

func (e *NotFoundErr) Is(target error) bool { return target == ErrNotFound }

type AddOptions struct {
	Amount      Cents
	Description string
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
		return nil, fmt.Errorf("failed to load expenses: %w", err)
	}
	for _, e := range state.Expenses {
		state.NextID = max(state.NextID, e.ID+1)
	}
	state.NextID = max(state.NextID, 1)
	return state, nil
}

func (s *Service) save(state *State) error {
	if err := s.store.Save(state); err != nil {
		return fmt.Errorf("failed to save expenses: %w", err)
	}
	return nil
}

// Validate checks the rules every new expense must follow, whichever
// interface (CLI, tests, a future API) it comes from.
func (o *AddOptions) Validate() error {
	o.Description = strings.TrimSpace(o.Description)
	if o.Description == "" {
		return errors.New("description cannot be empty")
	}
	if o.Amount <= 0 {
		return errors.New("amount must be greater than 0")
	}
	return nil
}

func (s *Service) Add(opts AddOptions) (ExpenseEntity, error) {
	if err := opts.Validate(); err != nil {
		return ExpenseEntity{}, err
	}

	state, err := s.load()
	if err != nil {
		return ExpenseEntity{}, err
	}

	now := s.now()
	newExpense := ExpenseEntity{
		ID:          state.NextID,
		Description: opts.Description,
		Amount:      opts.Amount,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	state.NextID++
	state.Expenses = append(state.Expenses, newExpense)

	return newExpense, s.save(state)
}

func (s *Service) Get(id int) (ExpenseEntity, error) {
	state, err := s.load()
	if err != nil {
		return ExpenseEntity{}, err
	}

	i := indexOf(state.Expenses, id)
	if i < 0 {
		return ExpenseEntity{}, &NotFoundErr{IDs: []int{id}}
	}

	return state.Expenses[i], nil
}

func (s *Service) List() ([]ExpenseEntity, error) {
	state, err := s.load()
	if err != nil {
		return nil, err
	}

	slices.SortFunc(state.Expenses, func(a, b ExpenseEntity) int {
		return cmp.Compare(a.ID, b.ID)
	})

	return state.Expenses, nil
}

// SummaryOptions filters which expenses Summary adds up. Zero values mean
// "any", so SummaryOptions{} totals everything.
type SummaryOptions struct {
	Year  int
	Month time.Month
}

func (o SummaryOptions) Validate() error {
	if o.Month < 0 || o.Month > time.December {
		return fmt.Errorf("invalid month %d: must be 1-12", o.Month)
	}
	if o.Year < 0 {
		return fmt.Errorf("invalid year %d", o.Year)
	}
	return nil
}

func (s *Service) Summary(opts SummaryOptions) (Cents, error) {
	if err := opts.Validate(); err != nil {
		return 0, err
	}

	state, err := s.load()
	if err != nil {
		return 0, err
	}

	now := s.now()
	// A month on its own means that month of the current year, not that
	// month across every year.
	if opts.Month != 0 && opts.Year == 0 {
		opts.Year = now.Year()
	}

	var total Cents
	for _, e := range state.Expenses {
		// Compare in local time: an expense made late on 31 July locally
		// can be 1 August in UTC, and the user thinks of it as July.
		created := e.CreatedAt.In(now.Location())
		if opts.Year != 0 && created.Year() != opts.Year {
			continue
		}
		if opts.Month != 0 && created.Month() != opts.Month {
			continue
		}
		total += e.Amount
	}
	return total, nil
}

func ParseMonth(in string) (time.Month, error) {
	m, err := strconv.Atoi(strings.TrimSpace(in))
	if err != nil || m < 1 || m > 12 {
		return 0, fmt.Errorf("invalid month %q: use a number from 1 to 12", in)
	}
	return time.Month(m), nil
}

// ParseYear only accepts four-digit years, so a typo like "24" is an error
// instead of quietly meaning the year 24.
func ParseYear(in string) (int, error) {
	y, err := strconv.Atoi(strings.TrimSpace(in))
	if err != nil || y < 1000 || y > 9999 {
		return 0, fmt.Errorf("invalid year %q: use a four-digit year like 2024", in)
	}
	return y, nil
}

func indexOf(expenses []ExpenseEntity, id int) int {
	return slices.IndexFunc(expenses, func(e ExpenseEntity) bool { return e.ID == id })
}
