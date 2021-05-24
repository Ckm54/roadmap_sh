package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/ckm54/roadmap_sh/expense-tracker/internal/expense"
)

func (a *App) printTable(expenses []expense.ExpenseEntity) error {
	w := tabwriter.NewWriter(a.out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "# ID\tDate\tDescription\tAmount")
	for _, e := range expenses {
		fmt.Fprintf(w, "# %d\t%s\t%s\t$%s\n", e.ID, e.CreatedAt.Local().Format(expense.DateLayout), e.Description, e.Amount)
	}
	return w.Flush()
}
