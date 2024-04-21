package cli

import "testing"

func TestHandleAdd_CliParsing(t *testing.T) {
	tests := []struct {
		name      string
		inputArgs []string
		wantErr   bool
		wantTitle string
	}{

		{
			name:      "Valid single word",
			inputArgs: []string{"groceries"},
			wantErr:   false,
			wantTitle: "groceries",
		},
		{
			name:      "Valid multi-word string",
			inputArgs: []string{"clean", "the", "car"},
			wantErr:   false,
			wantTitle: "clean the car",
		},
		{
			name:      "Strips out sorrounding quotation marks",
			inputArgs: []string{`"clean`, "the", `car"`},
			wantErr:   false,
			wantTitle: "clean the car",
		},
		{
			name:      "Keeps quotes within the text",
			inputArgs: []string{`"clean`, `"the"`, `car"`},
			wantErr:   false,
			wantTitle: `clean "the" car`,
		},
		{
			name:      "Empty input returns validation error",
			inputArgs: []string{},
			wantErr:   true,
			wantTitle: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockCliService{}
			svc = mockSvc

			err := handleAdd(tt.inputArgs)
			if (err != nil) != tt.wantErr {
				t.Errorf("HandleAdd() unexpected error status = %v, want %v", err, tt.wantErr)
			}

			if !tt.wantErr && mockSvc.calledAddTitle != tt.wantTitle {
				t.Errorf("HandleAdd() forwaded title = %q, expected %q", mockSvc.calledAddTitle, tt.wantTitle)
			}
		})
	}
}
