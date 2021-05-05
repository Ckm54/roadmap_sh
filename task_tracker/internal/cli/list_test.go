package cli

import "testing"

func TestHandleList_CLIRouting(t *testing.T) {
	tests := []struct {
		name          string
		inputArgs     []string
		shouldFailSvc bool
		wantErr       bool
		wantFilter    string
	}{
		{
			name:       "Lists all tasks when no filter is provided",
			inputArgs:  []string{},
			wantErr:    false,
			wantFilter: "",
		},
		{
			name:       "Filters by done status",
			inputArgs:  []string{"done"},
			wantErr:    false,
			wantFilter: "done",
		},
		{
			name:       "Filters by todo status",
			inputArgs:  []string{"todo"},
			wantErr:    false,
			wantFilter: "todo",
		},
		{
			name:       "Filters by in-progress status",
			inputArgs:  []string{"in-progress"},
			wantErr:    false,
			wantFilter: "in-progress",
		},
		{
			name:      "Fails with an unknown status filter",
			inputArgs: []string{"invalid-status"},
			wantErr:   true,
		},
		{
			name:      "Fails when more than one argument is provided",
			inputArgs: []string{"done", "todo"},
			wantErr:   true,
		},
		{
			name:          "Forwards backend errors to the REPL",
			inputArgs:     []string{},
			shouldFailSvc: true,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockCliService{shouldFail: tt.shouldFailSvc}
			svc = mockSvc

			err := handleList(tt.inputArgs)
			if (err != nil) != tt.wantErr {
				t.Fatalf("HandleList() error=%v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && mockSvc.calledListFilter != tt.wantFilter {
				t.Errorf("HandleList() forwarded filter=%q, expected %q", mockSvc.calledListFilter, tt.wantFilter)
			}
		})
	}
}
