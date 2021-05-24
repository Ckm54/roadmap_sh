package cli

import "fmt"

// runList handles: expense-tracker list
func (a *App) runList(args []string) error {
	fs := a.flagSet("list")

	if err := a.parse(fs, args); err != nil {
		return err
	}

	expenses, err := a.svc.List()
	if err != nil {
		return err
	}

	if len(expenses) == 0 {
		fmt.Fprintln(a.out, "No expenses found")
		return nil
	}

	return a.printTable(expenses)
}
