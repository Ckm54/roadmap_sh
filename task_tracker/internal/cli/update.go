package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ckm54/task_tracker/internal/task"
)

func (a *App) runUpdate(args []string) error {
	fs := a.flagSet("update")
	var priority, due string
	var tags tagsFlag
	var clearTags bool
	fs.StringVar(&priority, "priority", "", "new priority `level`: low, medium or high")
	fs.StringVar(&priority, "p", "", "shorthand for --priority")
	fs.StringVar(&due, "due", "", "new due `date`: YYYY-MM-DD, today, tomorrow, or none to remove it")
	fs.Var(&tags, "tag", "replace the tags with `tag`; repeat or comma-separate for several")
	fs.Var(&tags, "t", "shorthand for --tag")
	fs.BoolVar(&clearTags, "clear-tags", false, "remove all tags")

	positional, err := a.parse(fs, args)
	if err != nil {
		return err
	}
	if len(positional) == 0 {
		return errors.New("usage: update <id> [description] [flags]")
	}

	id, err := parseID(positional[0])
	if err != nil {
		return err
	}

	var patch task.Patch
	changed := false

	if len(positional) > 1 {
		description := joinText(positional[1:])
		if description == "" {
			return errors.New("new description cannot be empty")
		}
		patch.Description = &description
		changed = true
	}
	if priority != "" {
		p, err := task.ParsePriority(priority)
		if err != nil {
			return err
		}
		patch.Priority = &p
		changed = true
	}
	if due != "" {
		if strings.EqualFold(due, "none") {
			patch.ClearDue = true
		} else if patch.Due, err = parseDate(due, a.now()); err != nil {
			return err
		}
		changed = true
	}
	if clearTags && len(tags) > 0 {
		return errors.New("use either --tag or --clear-tags, not both")
	}
	if clearTags || len(tags) > 0 {
		t := []string(tags)
		patch.Tags = &t
		changed = true
	}

	if !changed {
		return errors.New("nothing to update; give a new description or a flag (see 'help update')")
	}

	if err := a.svc.Update(id, patch); err != nil {
		return err
	}

	fmt.Fprintf(a.out, "Task %d updated successfully\n", id)
	return nil
}

func (a *App) statusRunner(status task.Status) func(args []string) error {
	return func(args []string) error {
		positional, err := a.parse(a.flagSet("mark-"+string(status)), args)
		if err != nil {
			return err
		}

		ids, err := parseIDs(positional)
		if err != nil {
			return err
		}

		if err := a.svc.SetStatus(ids, status); err != nil {
			return err
		}

		if len(ids) == 1 {
			fmt.Fprintf(a.out, "Task %d marked as %s\n", ids[0], status)
		} else {
			fmt.Fprintf(a.out, "Tasks %s marked as %s\n", joinIDs(ids), status)
		}
		return nil
	}
}
