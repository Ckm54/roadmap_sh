package cli

import (
	"testing"

	"github.com/ckm54/task_tracker/internal/constants"
)

func TestHandleUpdateStatus_CLIRouting(t *testing.T) {
	tests := []struct {
		name          string
		inputArgs     []string
		shouldFailSvc bool
		wantErr       bool
		wantID        int
		wantStatus    string
	}{
		{
			name:       "Successfully updates an existing task",
			inputArgs:  []string{"mark-in-progress", "1"},
			wantErr:    false,
			wantID:     1,
			wantStatus: constants.StatusInProgress,
		},
		{
			name:       "Successfully updates a task with a different ID",
			inputArgs:  []string{"mark-done", "234"},
			wantErr:    false,
			wantID:     234,
			wantStatus: constants.StatusDone,
		},
		{
			name:      "Fails when no arguments are passed",
			inputArgs: []string{},
			wantErr:   true,
		},
		{
			name:      "Fails when invalid command is provided",
			inputArgs: []string{"update", "23"},
			wantErr:   true,
		},
		{
			name:      "Fails to mark in progress when id is missing",
			inputArgs: []string{"mark-in-progress"},
			wantErr:   true,
		},
		{
			name:      "Fails to mark in progress if ID is non-numeric",
			inputArgs: []string{"mark-in-progress", "abcds"},
			wantErr:   true,
		},
		{
			name:      "Fails to mark done when id is missing",
			inputArgs: []string{"mark-done"},
			wantErr:   true,
		},
		{
			name:      "Fails to mark done if ID is non-numeric",
			inputArgs: []string{"mark-done", "abcds"},
			wantErr:   true,
		},
		{
			name:          "Forwards backend errors to the REPL",
			inputArgs:     []string{"mark-in-progress", "1"},
			shouldFailSvc: true,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockCliService{shouldFail: tt.shouldFailSvc}
			svc = mockSvc

			err := handleUpdateStatus(tt.inputArgs)

			if (err != nil) != tt.wantErr {
				t.Fatalf("handleUpdateStatus() error=%v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if mockSvc.calledUpdateStatusID != tt.wantID {
					t.Errorf("handleUpdateStatus() forwaded ID=%d, expected %d", mockSvc.calledUpdateStatusID, tt.wantID)
				}
				if mockSvc.calledUpdateStatus != tt.wantStatus {
					t.Errorf("handleUpdateStatus() forwaded Status=%s, expected %s", mockSvc.calledUpdateStatus, tt.wantStatus)
				}
			}
		})
	}
}
