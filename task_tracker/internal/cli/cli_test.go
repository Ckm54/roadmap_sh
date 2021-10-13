package cli

import (
	"bytes"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/ckm54/task_tracker/internal/task"
)

var errBackend = errors.New("backend failure")

// mockService records the last call to each method.
type mockService struct {
	tasks      []task.TaskEntity
	shouldFail bool

	addDescription string
	addOpts        task.AddOptions
	updateID       int
	updatePatch    task.Patch
	statusIDs      []int
	status         task.Status
	deleteIDs      []int
	cleared        task.Status
	listOpts       task.ListOptions
}

func (m *mockService) Add(description string, opts task.AddOptions) (task.TaskEntity, error) {
	m.addDescription, m.addOpts = description, opts
	if m.shouldFail {
		return task.TaskEntity{}, errBackend
	}
	return task.TaskEntity{ID: 7, Description: description}, nil
}

func (m *mockService) Get(id int) (task.TaskEntity, error) {
	for _, t := range m.tasks {
		if t.ID == id {
			return t, nil
		}
	}
	return task.TaskEntity{}, &task.NotFoundError{IDs: []int{id}}
}

func (m *mockService) Update(id int, patch task.Patch) error {
	m.updateID, m.updatePatch = id, patch
	if m.shouldFail {
		return errBackend
	}
	return nil
}

func (m *mockService) SetStatus(ids []int, status task.Status) error {
	m.statusIDs, m.status = ids, status
	if m.shouldFail {
		return errBackend
	}
	return nil
}

func (m *mockService) Delete(ids []int) error {
	m.deleteIDs = ids
	if m.shouldFail {
		return errBackend
	}
	return nil
}

func (m *mockService) Clear(status task.Status) (int, error) {
	m.cleared = status
	if m.shouldFail {
		return 0, errBackend
	}
	return 2, nil
}

func (m *mockService) List(opts task.ListOptions) ([]task.TaskEntity, error) {
	m.listOpts = opts
	if m.shouldFail {
		return nil, errBackend
	}
	return m.tasks, nil
}

var fixedNow = time.Date(2026, 10, 17, 12, 0, 0, 0, time.Local)

// newTestApp returns an app writing to a buffer, with optional stdin.
func newTestApp(svc *mockService, stdin string) (*App, *bytes.Buffer) {
	out := &bytes.Buffer{}
	opts := Options{Out: out, Now: func() time.Time { return fixedNow }}
	if stdin != "" {
		opts.In = strings.NewReader(stdin)
	}
	return New(svc, opts), out
}

func date(y int, m time.Month, d int) *time.Time {
	v := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	return &v
}

func TestRun_UnknownCommand(t *testing.T) {
	app, _ := newTestApp(&mockService{}, "")

	err := app.Run([]string{"lsit"})
	if err == nil || !strings.Contains(err.Error(), `did you mean "list"`) {
		t.Errorf("error = %v, want a suggestion", err)
	}

	err = app.Run([]string{"frobnicate"})
	if err == nil || strings.Contains(err.Error(), "did you mean") {
		t.Errorf("error = %v, want no suggestion", err)
	}
}

func TestRun_Help(t *testing.T) {
	for _, args := range [][]string{nil, {"help"}, {"--help"}} {
		app, out := newTestApp(&mockService{}, "")
		if err := app.Run(args); err != nil {
			t.Fatalf("Run(%q): %v", args, err)
		}
		for _, c := range app.commands {
			if !strings.Contains(out.String(), c.name) {
				t.Errorf("Run(%q) help is missing %q", args, c.name)
			}
		}
	}

	for _, args := range [][]string{{"help", "add"}, {"add", "--help"}, {"add", "-h"}} {
		svc := &mockService{}
		app, out := newTestApp(svc, "")
		if err := app.Run(args); err != nil {
			t.Fatalf("Run(%q): %v", args, err)
		}
		if !strings.Contains(out.String(), "-p, --priority level") {
			t.Errorf("Run(%q) output missing merged flag help:\n%s", args, out)
		}
		if svc.addDescription != "" {
			t.Errorf("Run(%q) should not add a task", args)
		}
	}
}

func TestRun_Version(t *testing.T) {
	out := &bytes.Buffer{}
	app := New(&mockService{}, Options{Out: out, Version: "v1.2.3"})
	if err := app.Run([]string{"--version"}); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); got != "task-cli v1.2.3\n" {
		t.Errorf("output = %q", got)
	}
}

func TestExecuteLine_Quoting(t *testing.T) {
	svc := &mockService{}
	app, _ := newTestApp(svc, "")

	if err := app.ExecuteLine(`add "Buy  milk" -t 'a b'`); err != nil {
		t.Fatal(err)
	}
	if svc.addDescription != "Buy  milk" || !slices.Equal(svc.addOpts.Tags, []string{"a b"}) {
		t.Errorf("got description %q, tags %q", svc.addDescription, svc.addOpts.Tags)
	}

	if err := app.ExecuteLine(`add "unterminated`); err == nil {
		t.Error("expected an error for an unterminated quote")
	}
}

func TestREPL(t *testing.T) {
	svc := &mockService{}
	app, out := newTestApp(svc, "\nadd first task\nbogus\nexit\nadd never runs\n")

	if err := app.REPL(); err != nil {
		t.Fatal(err)
	}
	if svc.addDescription != "first task" {
		t.Errorf("addDescription = %q", svc.addDescription)
	}
	if !strings.Contains(out.String(), `Error: unknown command "bogus"`) {
		t.Errorf("REPL did not report the bad command:\n%s", out)
	}

	// EOF ends the REPL cleanly too.
	app, _ = newTestApp(&mockService{}, "list")
	if err := app.REPL(); err != nil {
		t.Fatal(err)
	}
}

func TestSplitArgs(t *testing.T) {
	tests := []struct {
		line    string
		want    []string
		wantErr bool
	}{
		{"add buy milk", []string{"add", "buy", "milk"}, false},
		{"  add   spaced  ", []string{"add", "spaced"}, false},
		{`add "buy  milk"`, []string{"add", "buy  milk"}, false},
		{`add 'it''s'`, []string{"add", "its"}, false},
		{`add "it's \"quoted\""`, []string{"add", `it's "quoted"`}, false},
		{`add it\'s`, []string{"add", "it's"}, false},
		{`update 1 ""`, []string{"update", "1", ""}, false},
		{`add "oops`, nil, true},
	}

	for _, tt := range tests {
		got, err := splitArgs(tt.line)
		if (err != nil) != tt.wantErr {
			t.Errorf("splitArgs(%q) error = %v, wantErr %v", tt.line, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && !slices.Equal(got, tt.want) {
			t.Errorf("splitArgs(%q) = %q, want %q", tt.line, got, tt.want)
		}
	}
}

func TestParseDate(t *testing.T) {
	tests := []struct {
		in      string
		want    *time.Time
		wantErr bool
	}{
		{"2026-12-25", date(2026, 12, 25), false},
		{"today", date(2026, 10, 17), false},
		{"Tomorrow", date(2026, 10, 18), false},
		{"25/12/2026", nil, true},
	}

	for _, tt := range tests {
		got, err := parseDate(tt.in, fixedNow)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseDate(%q) error = %v", tt.in, err)
			continue
		}
		if !tt.wantErr && !got.Equal(*tt.want) {
			t.Errorf("parseDate(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestCompletion(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish"} {
		app, out := newTestApp(&mockService{}, "")
		if err := app.Run([]string{"completion", shell}); err != nil {
			t.Fatalf("completion %s: %v", shell, err)
		}
		if !strings.Contains(out.String(), "mark-in-progress") {
			t.Errorf("completion %s does not mention commands", shell)
		}
	}

	app, _ := newTestApp(&mockService{}, "")
	if err := app.Run([]string{"completion", "powershell"}); err == nil {
		t.Error("expected an error for an unsupported shell")
	}
}
