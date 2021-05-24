package cli

import (
	"fmt"

	"github.com/ckm54/roadmap_sh/expense-tracker/internal/expense"
)

// runAdd handles: expense-tracker add --description "Lunch" --amount 20
//
// The CLI only turns text into typed values and prints the result; the
// rules (non-empty description, positive amount) live in the service.
func (a *App) runAdd(args []string) error {
	fs := a.flagSet("add")
	var opts expense.AddOptions
	fs.StringVar(&opts.Description, "description", "", "what the money was spent on")
	// fs.Func lets us run our own parser on the raw text, so the amount goes
	// straight from string to Cents without ever being a float.
	fs.Func("amount", "`amount` spent, e.g. 12.50", func(s string) (err error) {
		opts.Amount, err = expense.ParseCents(s)
		return err
	})

	if err := a.parse(fs, args); err != nil {
		return err
	}

	e, err := a.svc.Add(opts)
	if err != nil {
		return err
	}

	fmt.Fprintf(a.out, "Expense added successfully (ID: %d)\n", e.ID)
	return nil
}
