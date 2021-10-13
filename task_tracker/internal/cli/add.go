package cli

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ckm54/task_tracker/internal/task"
)

func (a *App) runAdd(args []string) error {
	fs := a.flagSet("add")
	var priority, due string
	var tags tagsFlag
	fs.StringVar(&priority, "priority", "", "priority `level`: low, medium or high (default medium)")
	fs.StringVar(&priority, "p", "", "shorthand for --priority")
	fs.StringVar(&due, "due", "", "due `date`: YYYY-MM-DD, today or tomorrow")
	fs.Var(&tags, "tag", "`tag` to attach; repeat or comma-separate for several")
	fs.Var(&tags, "t", "shorthand for --tag")

	positional, err := a.parse(fs, args)
	if err != nil {
		return err
	}

	description := joinText(positional)
	if description == "" {
		return errors.New("task description cannot be empty; usage: add <description>")
	}

	opts := task.AddOptions{Tags: tags}
	if priority != "" {
		if opts.Priority, err = task.ParsePriority(priority); err != nil {
			return err
		}
	}
	if due != "" {
		if opts.Due, err = parseDate(due, a.now()); err != nil {
			return err
		}
	}

	t, err := a.svc.Add(description, opts)
	if err != nil {
		return err
	}

	fmt.Fprintf(a.out, "Task added successfully (ID: %d)\n", t.ID)
	return nil
}

// tagsFlag collects repeated and/or comma-separated --tag values.
type tagsFlag []string

func (t *tagsFlag) String() string { return strings.Join(*t, ",") }

func (t *tagsFlag) Set(v string) error {
	*t = append(*t, strings.Split(v, ",")...)
	return nil
}

// parseDate accepts YYYY-MM-DD, "today" or "tomorrow" in local time.
func parseDate(s string, now time.Time) (*time.Time, error) {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var d time.Time
	switch strings.ToLower(s) {
	case "today":
		d = today
	case "tomorrow":
		d = today.AddDate(0, 0, 1)
	default:
		var err error
		d, err = time.ParseInLocation(task.DateLayout, s, now.Location())
		if err != nil {
			return nil, fmt.Errorf("invalid date %q: use YYYY-MM-DD, today or tomorrow", s)
		}
	}
	return &d, nil
}
