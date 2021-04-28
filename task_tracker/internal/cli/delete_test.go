package cli

import "testing"

func TestHandleDelete_CLIRouting(t *testing.T) {
	tests := []struct {
		name          string
		inputArgs     []string
		shouldFailSvc bool
		wantErr       bool
		wantID        int
	}{
		{
			name:      "Successfully deletes an existing task",
			inputArgs: []string{"1"},
			wantErr:   false,
			wantID:    1,
		},
		{
			name:      "Successfully deletes with a different ID",
			inputArgs: []string{"42"},
			wantErr:   false,
			wantID:    42,
		},
		{
			name:      "Fails if no arguments are passed",
			inputArgs: []string{},
			wantErr:   true,
		},
		{
			name:      "Fails if ID is non-numeric",
			inputArgs: []string{"abc"},
			wantErr:   true,
		},
		{
			name:          "Forwards backend errors to the REPL",
			inputArgs:     []string{"1"},
			shouldFailSvc: true,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockCliService{shouldFail: tt.shouldFailSvc}
			svc = mockSvc

			err := handleDelete(tt.inputArgs)

			if (err != nil) != tt.wantErr {
				t.Fatalf("handleDelete() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && mockSvc.calledDeleteID != tt.wantID {
				t.Errorf("handleDelete() forwarded ID = %d, expected %d", mockSvc.calledDeleteID, tt.wantID)
			}
		})
	}
}
