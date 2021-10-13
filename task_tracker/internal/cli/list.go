package cli

import (
	"errors"
	"fmt"

	"github.com/ckm54/task_tracker/internal/task"
)

func (a *App) runList(args []string) error {
	fs := a.flagSet("list")
	var priority, tag, sortBy string
	var asJSON bool
	fs.StringVar(&priority, "priority", "", "only show tasks with this priority `level`")
	fs.StringVar(&priority, "p", "", "shorthand for --priority")
	fs.StringVar(&tag, "tag", "", "only show tasks with this `tag`")
	fs.StringVar(&tag, "t", "", "shorthand for --tag")
	fs.StringVar(&sortBy, "sort", "id", "sort `field`: id, priority, due, created or updated")
	fs.BoolVar(&asJSON, "json", false, "print tasks as JSON")

	positional, err := a.parse(fs, args)
	if err != nil {
		return err
	}
	if len(positional) > 1 {
		return errors.New("too many arguments; list takes at most one status filter")
	}

	var opts task.ListOptions
	if len(positional) == 1 {
		if opts.Status, err = task.ParseStatus(positional[0]); err != nil {
			return err
		}
	}
	if priority != "" {
		if opts.Priority, err = task.ParsePriority(priority); err != nil {
			return err
		}
	}
	opts.Tag = tag
	if opts.SortBy, err = task.ParseSortField(sortBy); err != nil {
		return err
	}

	tasks, err := a.svc.List(opts)
	if err != nil {
		return err
	}

	if asJSON {
		return a.printJSON(tasks)
	}

	if len(tasks) == 0 {
		if opts == (task.ListOptions{SortBy: opts.SortBy}) {
			fmt.Fprintln(a.out, "No tasks yet. To add one: add <description>")
		} else {
			fmt.Fprintln(a.out, "No tasks match the given filters")
		}
		return nil
	}

	a.printTable(tasks)
	return nil
}

func (a *App) runSearch(args []string) error {
	fs := a.flagSet("search")
	var sortBy string
	var asJSON bool
	fs.StringVar(&sortBy, "sort", "id", "sort `field`: id, priority, due, created or updated")
	fs.BoolVar(&asJSON, "json", false, "print tasks as JSON")

	positional, err := a.parse(fs, args)
	if err != nil {
		return err
	}

	query := joinText(positional)
	if query == "" {
		return errors.New("usage: search <text>")
	}

	opts := task.ListOptions{Query: query}
	if opts.SortBy, err = task.ParseSortField(sortBy); err != nil {
		return err
	}

	tasks, err := a.svc.List(opts)
	if err != nil {
		return err
	}

	if asJSON {
		return a.printJSON(tasks)
	}

	if len(tasks) == 0 {
		fmt.Fprintf(a.out, "No tasks match %q\n", query)
		return nil
	}

	a.printTable(tasks)
	return nil
}

func (a *App) runShow(args []string) error {
	fs := a.flagSet("show")
	var asJSON bool
	fs.BoolVar(&asJSON, "json", false, "print the task as JSON")

	positional, err := a.parse(fs, args)
	if err != nil {
		return err
	}
	if len(positional) != 1 {
		return errors.New("usage: show <id>")
	}

	id, err := parseID(positional[0])
	if err != nil {
		return err
	}

	t, err := a.svc.Get(id)
	if err != nil {
		return err
	}

	if asJSON {
		return a.printJSON(t)
	}

	a.printDetails(t)
	return nil
}
