package cli

import (
	"testing"
	"time"

	"github.com/ckm54/task_tracker/internal/constants"
	"github.com/ckm54/task_tracker/internal/task"
)

func TestHandleUpdate_CliRouting(t *testing.T) {
	tests := []struct {
		name          string
		inputArgs     []string
		shouldFailSvc bool
		wantErr       bool
		wantID        int
		wantTitle     string
	}{
		{
			name:      "Valid args with ID and multi-word title",
			inputArgs: []string{"1", "buy", "almond", "milk"},
			wantErr:   false,
			wantID:    1,
			wantTitle: "buy almond milk",
		},
		{
			name:      "Valid args with ID and single word title",
			inputArgs: []string{"2", "coding"},
			wantErr:   false,
			wantID:    2,
			wantTitle: "coding",
		},
		{
			name:      "Fails validation if no arguments are passed",
			inputArgs: []string{},
			wantErr:   true,
		},
		{
			name:      "Fails if only ID is provided but title is missing",
			inputArgs: []string{"1"},
			wantErr:   true,
		},
		{
			name:      "Fails if ID is provided but title is empty",
			inputArgs: []string{"1", "'   '"},
			wantErr:   true,
		},
		{
			name:      "Fails if ID is non-numeric",
			inputArgs: []string{"abc", "New", "Title"},
			wantErr:   true,
		},
		{
			name:          "Forwards backend errors to the REPL",
			inputArgs:     []string{"1", "new", "title"},
			shouldFailSvc: true,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockCliService{shouldFail: tt.shouldFailSvc}
			svc = mockSvc

			err := handleUpdate(tt.inputArgs)

			if (err != nil) != tt.wantErr {
				t.Fatalf("handleUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if mockSvc.calledUpdateID != tt.wantID {
					t.Errorf("handleUpdate() forwarded ID = %d, expected %d", mockSvc.calledUpdateID, tt.wantID)
				}
				if mockSvc.calledUpdateTitle != tt.wantTitle {
					t.Errorf("handleUpdate() forwarded title = %q, expected %q", mockSvc.calledUpdateTitle, tt.wantTitle)
				}
			}
		})
	}
}

func includeTask(id int, title string, t time.Time) task.TaskEntity {
	return task.TaskEntity{ID: id, Title: title, Status: constants.StatusTodo, CreatedAt: t, UpdatedAt: t}
}
