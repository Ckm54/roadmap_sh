package cli

import (
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/ckm54/task_tracker/internal/task"
)

func TestAdd(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		failSvc  bool
		wantErr  bool
		wantDesc string
		check    func(t *testing.T, opts task.AddOptions)
	}{
		{name: "single word", args: []string{"groceries"}, wantDesc: "groceries"},
		{name: "multiple words are joined", args: []string{"clean", "the", "car"}, wantDesc: "clean the car"},
		{name: "quotes inside text are kept", args: []string{`clean "the" car`}, wantDesc: `clean "the" car`},
		{
			name:     "flags after the description",
			args:     []string{"Buy", "milk", "-p", "high", "--due", "2026-10-20", "-t", "home,errands", "--tag", "food"},
			wantDesc: "Buy milk",
			check: func(t *testing.T, opts task.AddOptions) {
				if opts.Priority != task.PriorityHigh {
					t.Errorf("priority = %q", opts.Priority)
				}
				if opts.Due == nil || !opts.Due.Equal(*date(2026, 10, 20)) {
					t.Errorf("due = %v", opts.Due)
				}
				if !slices.Equal(opts.Tags, []string{"home", "errands", "food"}) {
					t.Errorf("tags = %q", opts.Tags)
				}
			},
		},
		{name: "flags before the description", args: []string{"--priority", "low", "Read", "book"}, wantDesc: "Read book"},
		{name: "double dash ends flags", args: []string{"--", "-p", "is", "text"}, wantDesc: "-p is text"},
		{name: "empty description", args: []string{}, wantErr: true},
		{name: "blank description", args: []string{"  "}, wantErr: true},
		{name: "unknown priority", args: []string{"x", "-p", "urgent"}, wantErr: true},
		{name: "bad due date", args: []string{"x", "--due", "soon"}, wantErr: true},
		{name: "unknown flag", args: []string{"x", "--colour"}, wantErr: true},
		{name: "backend errors are returned", args: []string{"x"}, failSvc: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockService{shouldFail: tt.failSvc}
			app, out := newTestApp(svc, "")

			err := app.Run(append([]string{"add"}, tt.args...))
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			if svc.addDescription != tt.wantDesc {
				t.Errorf("description = %q, want %q", svc.addDescription, tt.wantDesc)
			}
			if got := out.String(); got != "Task added successfully (ID: 7)\n" {
				t.Errorf("output = %q", got)
			}
			if tt.check != nil {
				tt.check(t, svc.addOpts)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		failSvc bool
		wantErr bool
		check   func(t *testing.T, p task.Patch)
	}{
		{
			name: "description only",
			args: []string{"1", "buy", "almond", "milk"},
			check: func(t *testing.T, p task.Patch) {
				if p.Description == nil || *p.Description != "buy almond milk" || p.Priority != nil || p.Tags != nil {
					t.Errorf("patch = %+v", p)
				}
			},
		},
		{
			name: "flags only",
			args: []string{"2", "-p", "high", "--due", "tomorrow", "-t", "work"},
			check: func(t *testing.T, p task.Patch) {
				if p.Description != nil || *p.Priority != task.PriorityHigh || !p.Due.Equal(*date(2026, 10, 18)) || !slices.Equal(*p.Tags, []string{"work"}) {
					t.Errorf("patch = %+v", p)
				}
			},
		},
		{
			name: "clear due date and tags",
			args: []string{"2", "--due", "none", "--clear-tags"},
			check: func(t *testing.T, p task.Patch) {
				if !p.ClearDue || p.Tags == nil || len(*p.Tags) != 0 {
					t.Errorf("patch = %+v", p)
				}
			},
		},
		{name: "no arguments", args: []string{}, wantErr: true},
		{name: "nothing to change", args: []string{"1"}, wantErr: true},
		{name: "blank description", args: []string{"1", "   "}, wantErr: true},
		{name: "non-numeric ID", args: []string{"abc", "title"}, wantErr: true},
		{name: "zero ID", args: []string{"0", "title"}, wantErr: true},
		{name: "tag and clear-tags together", args: []string{"1", "-t", "x", "--clear-tags"}, wantErr: true},
		{name: "backend errors are returned", args: []string{"1", "x"}, failSvc: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockService{shouldFail: tt.failSvc}
			app, _ := newTestApp(svc, "")

			err := app.Run(append([]string{"update"}, tt.args...))
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				tt.check(t, svc.updatePatch)
			}
		})
	}
}

func TestMarkStatus(t *testing.T) {
	tests := []struct {
		args       []string
		failSvc    bool
		wantErr    bool
		wantIDs    []int
		wantStatus task.Status
		wantOut    string
	}{
		{args: []string{"mark-in-progress", "1"}, wantIDs: []int{1}, wantStatus: task.StatusInProgress, wantOut: "Task 1 marked as in-progress\n"},
		{args: []string{"mark-done", "3", "1", "3"}, wantIDs: []int{3, 1}, wantStatus: task.StatusDone, wantOut: "Tasks 3, 1 marked as done\n"},
		{args: []string{"mark-todo", "2"}, wantIDs: []int{2}, wantStatus: task.StatusTodo, wantOut: "Task 2 marked as todo\n"},
		{args: []string{"mark-done"}, wantErr: true},
		{args: []string{"mark-done", "abc"}, wantErr: true},
		{args: []string{"mark-done", "1", "-2"}, wantErr: true},
		{args: []string{"mark-done", "1"}, failSvc: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			svc := &mockService{shouldFail: tt.failSvc}
			app, out := newTestApp(svc, "")

			err := app.Run(tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !slices.Equal(svc.statusIDs, tt.wantIDs) || svc.status != tt.wantStatus {
				t.Errorf("SetStatus(%v, %q), want (%v, %q)", svc.statusIDs, svc.status, tt.wantIDs, tt.wantStatus)
			}
			if out.String() != tt.wantOut {
				t.Errorf("output = %q, want %q", out, tt.wantOut)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		stdin       string
		failSvc     bool
		wantErr     bool
		wantDeleted []int
		wantOut     string
	}{
		{name: "force skips the prompt", args: []string{"1", "-f"}, wantDeleted: []int{1}, wantOut: "Task 1 deleted"},
		{name: "several IDs", args: []string{"--force", "1", "2"}, wantDeleted: []int{1, 2}, wantOut: "Tasks 1, 2 deleted"},
		{name: "confirmed with y", args: []string{"4"}, stdin: "y\n", wantDeleted: []int{4}, wantOut: "Delete task 4? [y/N]: Task 4 deleted"},
		{name: "declined", args: []string{"4"}, stdin: "n\n", wantOut: "Delete task 4? [y/N]: Aborted."},
		{name: "empty answer declines", args: []string{"4"}, stdin: "\n", wantOut: "Delete task 4? [y/N]: Aborted."},
		{name: "no input asks for --force", args: []string{"4"}, wantErr: true},
		{name: "no IDs", args: []string{"-f"}, wantErr: true},
		{name: "non-numeric ID", args: []string{"abc", "-f"}, wantErr: true},
		{name: "backend errors are returned", args: []string{"1", "-f"}, failSvc: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockService{shouldFail: tt.failSvc}
			app, out := newTestApp(svc, tt.stdin)

			err := app.Run(append([]string{"delete"}, tt.args...))
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !slices.Equal(svc.deleteIDs, tt.wantDeleted) {
				t.Errorf("deleted %v, want %v", svc.deleteIDs, tt.wantDeleted)
			}
			if got := strings.TrimSpace(out.String()); got != tt.wantOut {
				t.Errorf("output = %q, want %q", got, tt.wantOut)
			}
		})
	}
}

func TestClear(t *testing.T) {
	done := []task.TaskEntity{{ID: 1, Status: task.StatusDone}, {ID: 2, Status: task.StatusDone}}

	t.Run("clears done tasks after confirmation", func(t *testing.T) {
		svc := &mockService{tasks: done}
		app, out := newTestApp(svc, "yes\n")
		if err := app.Run([]string{"clear"}); err != nil {
			t.Fatal(err)
		}
		if svc.cleared != task.StatusDone || svc.listOpts.Status != task.StatusDone {
			t.Errorf("cleared %q after listing %+v", svc.cleared, svc.listOpts)
		}
		if !strings.Contains(out.String(), "Delete 2 done tasks? [y/N]: Cleared 2 done tasks") {
			t.Errorf("output = %q", out)
		}
	})

	t.Run("nothing to clear", func(t *testing.T) {
		svc := &mockService{}
		app, out := newTestApp(svc, "")
		if err := app.Run([]string{"clear"}); err != nil {
			t.Fatal(err)
		}
		if svc.cleared != "" || out.String() != "No done tasks to clear\n" {
			t.Errorf("cleared %q, output %q", svc.cleared, out)
		}
	})

	t.Run("rejects arguments", func(t *testing.T) {
		app, _ := newTestApp(&mockService{tasks: done}, "")
		if err := app.Run([]string{"clear", "todo", "-f"}); err == nil {
			t.Error("expected an error")
		}
	})
}

func TestList(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		failSvc  bool
		wantErr  bool
		wantOpts task.ListOptions
	}{
		{name: "no filter", args: nil, wantOpts: task.ListOptions{SortBy: task.SortID}},
		{name: "status filter", args: []string{"done"}, wantOpts: task.ListOptions{Status: task.StatusDone, SortBy: task.SortID}},
		{name: "all filters", args: []string{"in-progress", "-p", "high", "-t", "work", "--sort", "due"},
			wantOpts: task.ListOptions{Status: task.StatusInProgress, Priority: task.PriorityHigh, Tag: "work", SortBy: task.SortDue}},
		{name: "unknown status", args: []string{"doing"}, wantErr: true},
		{name: "two statuses", args: []string{"done", "todo"}, wantErr: true},
		{name: "unknown sort field", args: []string{"--sort", "colour"}, wantErr: true},
		{name: "unknown priority", args: []string{"-p", "urgent"}, wantErr: true},
		{name: "backend errors are returned", failSvc: true, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockService{shouldFail: tt.failSvc}
			app, _ := newTestApp(svc, "")

			err := app.Run(append([]string{"list"}, tt.args...))
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && svc.listOpts != tt.wantOpts {
				t.Errorf("opts = %+v, want %+v", svc.listOpts, tt.wantOpts)
			}
		})
	}
}

func TestList_Output(t *testing.T) {
	tasks := []task.TaskEntity{
		{ID: 1, Description: "Buy groceries", Status: task.StatusTodo, Priority: task.PriorityHigh, Due: date(2026, 10, 1), Tags: []string{"home"}},
		{ID: 12, Description: "Write tests", Status: task.StatusInProgress, Priority: task.PriorityMedium},
	}

	t.Run("table with aligned columns and summary", func(t *testing.T) {
		app, out := newTestApp(&mockService{tasks: tasks}, "")
		if err := app.Run([]string{"list"}); err != nil {
			t.Fatal(err)
		}

		want := "" +
			"ID  STATUS       PRIORITY  DUE         DESCRIPTION    TAGS\n" +
			"1   todo         high      2026-10-01  Buy groceries  home\n" +
			"12  in-progress  medium    -           Write tests    -\n" +
			"\n" +
			"2 tasks: 1 todo, 1 in-progress, 0 done\n"
		if out.String() != want {
			t.Errorf("output:\n%s\nwant:\n%s", out, want)
		}
	})

	t.Run("colour keeps alignment and marks overdue dates", func(t *testing.T) {
		svc := &mockService{tasks: tasks}
		app, out := newTestApp(svc, "")
		app.color = true
		if err := app.Run([]string{"list"}); err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(out.String(), "\x1b[31m2026-10-01\x1b[0m") {
			t.Errorf("overdue date not red:\n%q", out)
		}
		plain := stripANSI(out.String())
		if !strings.Contains(plain, "12  in-progress  medium    -           Write tests    -\n") {
			t.Errorf("colour broke alignment:\n%s", plain)
		}
	})

	t.Run("json", func(t *testing.T) {
		app, out := newTestApp(&mockService{tasks: tasks}, "")
		if err := app.Run([]string{"list", "--json"}); err != nil {
			t.Fatal(err)
		}
		var got []task.TaskEntity
		if err := json.Unmarshal(out.Bytes(), &got); err != nil || len(got) != 2 {
			t.Errorf("invalid JSON (%v):\n%s", err, out)
		}
	})

	t.Run("empty json is an array", func(t *testing.T) {
		app, out := newTestApp(&mockService{}, "")
		if err := app.Run([]string{"list", "--json"}); err != nil {
			t.Fatal(err)
		}
		if strings.TrimSpace(out.String()) != "[]" {
			t.Errorf("output = %q", out)
		}
	})

	t.Run("empty messages", func(t *testing.T) {
		app, out := newTestApp(&mockService{}, "")
		_ = app.Run([]string{"list"})
		_ = app.Run([]string{"list", "done"})
		if !strings.Contains(out.String(), "No tasks yet") || !strings.Contains(out.String(), "No tasks match the given filters") {
			t.Errorf("output = %q", out)
		}
	})
}

func TestSearch(t *testing.T) {
	svc := &mockService{}
	app, out := newTestApp(svc, "")

	if err := app.Run([]string{"search", "buy", "milk"}); err != nil {
		t.Fatal(err)
	}
	if svc.listOpts.Query != "buy milk" {
		t.Errorf("query = %q", svc.listOpts.Query)
	}
	if out.String() != "No tasks match \"buy milk\"\n" {
		t.Errorf("output = %q", out)
	}

	if err := app.Run([]string{"search"}); err == nil {
		t.Error("expected an error without a query")
	}
}

func TestShow(t *testing.T) {
	svc := &mockService{tasks: []task.TaskEntity{
		{ID: 3, Description: "Buy groceries", Status: task.StatusTodo, Priority: task.PriorityLow, Due: date(2026, 10, 1), Tags: []string{"home", "food"}},
	}}

	app, out := newTestApp(svc, "")
	if err := app.Run([]string{"show", "3"}); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"ID:          3", "Description: Buy groceries", "Due:         2026-10-01 (overdue)", "Tags:        home, food"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}

	out.Reset()
	if err := app.Run([]string{"show", "3", "--json"}); err != nil {
		t.Fatal(err)
	}
	var got task.TaskEntity
	if err := json.Unmarshal(out.Bytes(), &got); err != nil || got.ID != 3 {
		t.Errorf("invalid JSON (%v): %s", err, out)
	}

	if err := app.Run([]string{"show", "99"}); !errors.Is(err, task.ErrNotFound) {
		t.Errorf("error = %v, want ErrNotFound", err)
	}
	if err := app.Run([]string{"show"}); err == nil {
		t.Error("expected an error without an ID")
	}
}

func stripANSI(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' {
			for i < len(s) && s[i] != 'm' {
				i++
			}
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}
