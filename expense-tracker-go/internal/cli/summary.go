package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ckm54/roadmap_sh/expense-tracker/internal/expense"
)

func (a *App) runSummary(args []string) error {
	fs := a.flagSet("summary")
	var opts expense.SummaryOptions
	fs.Func("month", "`month` to total, 1-12 (current year unless --year is set)", func(s string) (err error) {
		opts.Month, err = expense.ParseMonth(s)
		return err
	})
	fs.Func("year", "`year` to total, e.g. 2024", func(s string) (err error) {
		opts.Year, err = expense.ParseYear(s)
		return err
	})

	if err := a.parse(fs, args); err != nil {
		return err
	}

	total, err := a.svc.Summary(opts)
	if err != nil {
		return err
	}

	fmt.Fprintf(a.out, "Total expenses%s: $%s\n", periodLabel(opts), total)
	return nil
}

// periodLabel names only the filters the user passed: "", " for 2024",
// " for August" or " for August 2024".
func periodLabel(opts expense.SummaryOptions) string {
	var parts []string
	if opts.Month != 0 {
		parts = append(parts, opts.Month.String())
	}
	if opts.Year != 0 {
		parts = append(parts, strconv.Itoa(opts.Year))
	}
	if len(parts) == 0 {
		return ""
	}
	return " for " + strings.Join(parts, " ")
}
