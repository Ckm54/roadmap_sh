package task

import (
	"errors"
	"slices"
	"testing"
	"time"
)

// memStore behaves like a file: Load returns a copy, so the service can
// only change what is stored by calling Save.
type memStore struct {
	state      State
	shouldFail bool
	saves      int
}

func (m *memStore) Load() (*State, error) {
	if m.shouldFail {
		return nil, errors.New("simulated storage error")
	}
	return &State{NextID: m.state.NextID, Tasks: slices.Clone(m.state.Tasks)}, nil
}

func (m *memStore) Save(s *State) error {
	m.saves++
	m.state = State{NextID: s.NextID, Tasks: slices.Clone(s.Tasks)}
	return nil
}

var (
	baseTime = time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	nowTime  = baseTime.Add(48 * time.Hour)
)

func newTestService(tasks ...TaskEntity) (*Service, *memStore) {
	store := &memStore{state: State{Tasks: tasks}}
	svc := NewService(store)
	svc.now = func() time.Time { return nowTime }
	return svc, store
}

func taskWith(id int, status Status) TaskEntity {
	return TaskEntity{ID: id, Description: "task", Status: status, Priority: PriorityMedium, CreatedAt: baseTime, UpdatedAt: baseTime}
}

func ids(tasks []TaskEntity) []int {
	out := make([]int, len(tasks))
	for i, t := range tasks {
		out[i] = t.ID
	}
	return out
}

func TestService_Add(t *testing.T) {
	t.Run("first task gets ID 1 and defaults", func(t *testing.T) {
		svc, store := newTestService()

		got, err := svc.Add("Buy groceries", AddOptions{Tags: []string{" Home ", "home", "", "Errands"}})
		if err != nil {
			t.Fatal(err)
		}

		if got.ID != 1 || got.Status != StatusTodo || got.Priority != PriorityMedium {
			t.Errorf("unexpected task: %+v", got)
		}
		if !slices.Equal(got.Tags, []string{"home", "errands"}) {
			t.Errorf("tags not normalised: %q", got.Tags)
		}
		if !got.CreatedAt.Equal(nowTime) || !got.UpdatedAt.Equal(nowTime) {
			t.Errorf("timestamps not set: %+v", got)
		}
		if store.state.NextID != 2 || len(store.state.Tasks) != 1 {
			t.Errorf("unexpected stored state: %+v", store.state)
		}
	})

	t.Run("IDs are never reused after deleting the newest task", func(t *testing.T) {
		svc, _ := newTestService()
		for range 3 {
			if _, err := svc.Add("task", AddOptions{}); err != nil {
				t.Fatal(err)
			}
		}
		if err := svc.Delete([]int{3}); err != nil {
			t.Fatal(err)
		}

		got, err := svc.Add("next", AddOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 4 {
			t.Errorf("got ID %d, want 4", got.ID)
		}
	})

	t.Run("recovers NextID from tasks when missing", func(t *testing.T) {
		svc, _ := newTestService(taskWith(7, StatusTodo))
		got, err := svc.Add("next", AddOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != 8 {
			t.Errorf("got ID %d, want 8", got.ID)
		}
	})

	t.Run("keeps given priority and due date", func(t *testing.T) {
		svc, _ := newTestService()
		due := baseTime
		got, err := svc.Add("x", AddOptions{Priority: PriorityHigh, Due: &due})
		if err != nil {
			t.Fatal(err)
		}
		if got.Priority != PriorityHigh || got.Due == nil || !got.Due.Equal(due) {
			t.Errorf("unexpected task: %+v", got)
		}
	})

	t.Run("fails when store is unavailable", func(t *testing.T) {
		svc, store := newTestService()
		store.shouldFail = true
		if _, err := svc.Add("x", AddOptions{}); err == nil {
			t.Fatal("expected an error")
		}
	})
}

func TestService_Get(t *testing.T) {
	svc, _ := newTestService(taskWith(1, StatusTodo), taskWith(2, StatusDone))

	got, err := svc.Get(2)
	if err != nil || got.ID != 2 {
		t.Fatalf("Get(2) = %+v, %v", got, err)
	}

	_, err = svc.Get(9)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get(9) error = %v, want ErrNotFound", err)
	}
}

func TestService_Update(t *testing.T) {
	description := "New description"
	high := PriorityHigh
	due := baseTime.AddDate(0, 0, 5)
	tags := []string{"Work"}
	noTags := []string{}

	withExtras := taskWith(1, StatusTodo)
	oldDue := baseTime
	withExtras.Due = &oldDue
	withExtras.Tags = []string{"old"}

	tests := []struct {
		name    string
		patch   Patch
		id      int
		wantErr error
		check   func(t *testing.T, got TaskEntity)
	}{
		{
			name:  "changes only the description",
			id:    1,
			patch: Patch{Description: &description},
			check: func(t *testing.T, got TaskEntity) {
				if got.Description != description || got.Priority != PriorityMedium || got.Due == nil {
					t.Errorf("unexpected task: %+v", got)
				}
			},
		},
		{
			name:  "changes priority, due date and tags",
			id:    1,
			patch: Patch{Priority: &high, Due: &due, Tags: &tags},
			check: func(t *testing.T, got TaskEntity) {
				if got.Priority != high || !got.Due.Equal(due) || !slices.Equal(got.Tags, []string{"work"}) {
					t.Errorf("unexpected task: %+v", got)
				}
			},
		},
		{
			name:  "clears due date and tags",
			id:    1,
			patch: Patch{ClearDue: true, Tags: &noTags},
			check: func(t *testing.T, got TaskEntity) {
				if got.Due != nil || len(got.Tags) != 0 {
					t.Errorf("unexpected task: %+v", got)
				}
			},
		},
		{
			name:    "fails for a missing ID",
			id:      42,
			patch:   Patch{Description: &description},
			wantErr: ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, store := newTestService(withExtras)

			err := svc.Update(tt.id, tt.patch)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Update() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return
			}

			got := store.state.Tasks[0]
			if !got.UpdatedAt.Equal(nowTime) || !got.CreatedAt.Equal(baseTime) {
				t.Errorf("timestamps wrong: %+v", got)
			}
			tt.check(t, got)
		})
	}
}

func TestService_SetStatus(t *testing.T) {
	t.Run("updates several tasks", func(t *testing.T) {
		svc, store := newTestService(taskWith(1, StatusTodo), taskWith(2, StatusTodo), taskWith(3, StatusTodo))

		if err := svc.SetStatus([]int{1, 3}, StatusDone); err != nil {
			t.Fatal(err)
		}

		want := []Status{StatusDone, StatusTodo, StatusDone}
		for i, task := range store.state.Tasks {
			if task.Status != want[i] {
				t.Errorf("task %d status = %q, want %q", task.ID, task.Status, want[i])
			}
		}
		if !store.state.Tasks[0].UpdatedAt.Equal(nowTime) || !store.state.Tasks[1].UpdatedAt.Equal(baseTime) {
			t.Error("UpdatedAt should change only on updated tasks")
		}
	})

	t.Run("changes nothing if any ID is missing", func(t *testing.T) {
		svc, store := newTestService(taskWith(1, StatusTodo))

		err := svc.SetStatus([]int{1, 5, 6}, StatusDone)
		var nf *NotFoundError
		if !errors.As(err, &nf) || !slices.Equal(nf.IDs, []int{5, 6}) {
			t.Fatalf("error = %v, want NotFoundError for 5 and 6", err)
		}
		if store.saves != 0 || store.state.Tasks[0].Status != StatusTodo {
			t.Error("store should be untouched")
		}
	})

	t.Run("can move a task back to todo", func(t *testing.T) {
		svc, store := newTestService(taskWith(1, StatusDone))
		if err := svc.SetStatus([]int{1}, StatusTodo); err != nil {
			t.Fatal(err)
		}
		if store.state.Tasks[0].Status != StatusTodo {
			t.Error("status not changed")
		}
	})
}

func TestService_Delete(t *testing.T) {
	t.Run("deletes several tasks", func(t *testing.T) {
		svc, store := newTestService(taskWith(1, StatusTodo), taskWith(2, StatusTodo), taskWith(3, StatusTodo))

		if err := svc.Delete([]int{1, 3}); err != nil {
			t.Fatal(err)
		}
		if got := ids(store.state.Tasks); !slices.Equal(got, []int{2}) {
			t.Errorf("remaining IDs = %v, want [2]", got)
		}
	})

	t.Run("deletes nothing if any ID is missing", func(t *testing.T) {
		svc, store := newTestService(taskWith(1, StatusTodo))

		if err := svc.Delete([]int{1, 9}); !errors.Is(err, ErrNotFound) {
			t.Fatalf("error = %v, want ErrNotFound", err)
		}
		if store.saves != 0 {
			t.Error("store should be untouched")
		}
	})
}

func TestService_Clear(t *testing.T) {
	svc, store := newTestService(taskWith(1, StatusDone), taskWith(2, StatusTodo), taskWith(3, StatusDone))

	n, err := svc.Clear(StatusDone)
	if err != nil || n != 2 {
		t.Fatalf("Clear() = %d, %v; want 2, nil", n, err)
	}
	if got := ids(store.state.Tasks); !slices.Equal(got, []int{2}) {
		t.Errorf("remaining IDs = %v, want [2]", got)
	}

	saves := store.saves
	if n, err := svc.Clear(StatusDone); err != nil || n != 0 || store.saves != saves {
		t.Errorf("second Clear() = %d, %v and saved=%v; want 0, nil, no save", n, err, store.saves != saves)
	}
}

func TestService_List(t *testing.T) {
	day := func(d int) *time.Time {
		v := baseTime.AddDate(0, 0, d)
		return &v
	}

	tasks := []TaskEntity{
		{ID: 1, Description: "Buy groceries", Status: StatusTodo, Priority: PriorityLow, Tags: []string{"home"}, Due: day(3), CreatedAt: baseTime},
		{ID: 2, Description: "Write tests", Status: StatusInProgress, Priority: PriorityHigh, Tags: []string{"work"}, CreatedAt: baseTime.Add(time.Hour)},
		{ID: 3, Description: "Deploy app", Status: StatusDone, Priority: PriorityMedium, Tags: []string{"work"}, Due: day(1), CreatedAt: baseTime.Add(-time.Hour)},
		{ID: 4, Description: "Review tests PR", Status: StatusDone, Priority: PriorityHigh, CreatedAt: baseTime},
		{ID: 5, Description: "legacy task with no priority", Status: StatusTodo, CreatedAt: baseTime},
	}

	tests := []struct {
		name    string
		opts    ListOptions
		wantIDs []int
	}{
		{"no filter returns everything by ID", ListOptions{}, []int{1, 2, 3, 4, 5}},
		{"filter by status", ListOptions{Status: StatusDone}, []int{3, 4}},
		{"filter by priority", ListOptions{Priority: PriorityHigh}, []int{2, 4}},
		{"missing priority counts as medium", ListOptions{Priority: PriorityMedium}, []int{3, 5}},
		{"filter by tag is case-insensitive", ListOptions{Tag: "WORK"}, []int{2, 3}},
		{"search is case-insensitive", ListOptions{Query: "TESTS"}, []int{2, 4}},
		{"filters combine", ListOptions{Status: StatusDone, Tag: "work"}, []int{3}},
		{"no matches", ListOptions{Status: StatusInProgress, Tag: "home"}, []int{}},
		{"sort by priority, then ID", ListOptions{SortBy: SortPriority}, []int{2, 4, 3, 5, 1}},
		{"sort by due puts undated tasks last", ListOptions{SortBy: SortDue}, []int{3, 1, 2, 4, 5}},
		{"sort by created", ListOptions{SortBy: SortCreated}, []int{3, 1, 4, 5, 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, store := newTestService(tasks...)

			got, err := svc.List(tt.opts)
			if err != nil {
				t.Fatal(err)
			}
			if gotIDs := ids(got); !slices.Equal(gotIDs, tt.wantIDs) {
				t.Errorf("IDs = %v, want %v", gotIDs, tt.wantIDs)
			}
			if store.saves != 0 {
				t.Error("List must not save")
			}
		})
	}

	t.Run("fails when store is unavailable", func(t *testing.T) {
		svc, store := newTestService()
		store.shouldFail = true
		if _, err := svc.List(ListOptions{}); err == nil {
			t.Fatal("expected an error")
		}
	})
}

func TestTaskEntity_Overdue(t *testing.T) {
	now := time.Date(2026, 10, 10, 15, 0, 0, 0, time.UTC)
	date := func(d int) *time.Time {
		v := time.Date(2026, 10, d, 0, 0, 0, 0, time.UTC)
		return &v
	}

	tests := []struct {
		name string
		task TaskEntity
		want bool
	}{
		{"no due date", TaskEntity{Status: StatusTodo}, false},
		{"due yesterday", TaskEntity{Status: StatusTodo, Due: date(9)}, true},
		{"due today is not overdue", TaskEntity{Status: StatusTodo, Due: date(10)}, false},
		{"done tasks are never overdue", TaskEntity{Status: StatusDone, Due: date(1)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.task.Overdue(now); got != tt.want {
				t.Errorf("Overdue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParsers(t *testing.T) {
	if _, err := ParseStatus("in-progress"); err != nil {
		t.Error(err)
	}
	if _, err := ParseStatus("doing"); err == nil {
		t.Error("ParseStatus accepted an unknown status")
	}
	if _, err := ParsePriority("high"); err != nil {
		t.Error(err)
	}
	if _, err := ParsePriority("urgent"); err == nil {
		t.Error("ParsePriority accepted an unknown priority")
	}
	if _, err := ParseSortField("due"); err != nil {
		t.Error(err)
	}
	if _, err := ParseSortField("colour"); err == nil {
		t.Error("ParseSortField accepted an unknown field")
	}
}
