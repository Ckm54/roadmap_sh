package cli

import (
	"fmt"

	"github.com/ckm54/task_tracker/internal/task"
)

func (a *App) runDelete(args []string) error {
	fs := a.flagSet("delete")
	var force bool
	fs.BoolVar(&force, "force", false, "delete without asking for confirmation")
	fs.BoolVar(&force, "f", false, "shorthand for --force")

	positional, err := a.parse(fs, args)
	if err != nil {
		return err
	}

	ids, err := parseIDs(positional)
	if err != nil {
		return err
	}

	if !force {
		question := fmt.Sprintf("Delete task %d?", ids[0])
		if len(ids) > 1 {
			question = fmt.Sprintf("Delete tasks %s?", joinIDs(ids))
		}
		ok, err := a.confirm(question)
		if err != nil {
			return err
		}
		if !ok {
			fmt.Fprintln(a.out, "Aborted.")
			return nil
		}
	}

	if err := a.svc.Delete(ids); err != nil {
		return err
	}

	if len(ids) == 1 {
		fmt.Fprintf(a.out, "Task %d deleted\n", ids[0])
	} else {
		fmt.Fprintf(a.out, "Tasks %s deleted\n", joinIDs(ids))
	}
	return nil
}

func (a *App) runClear(args []string) error {
	fs := a.flagSet("clear")
	var force bool
	fs.BoolVar(&force, "force", false, "clear without asking for confirmation")
	fs.BoolVar(&force, "f", false, "shorthand for --force")

	positional, err := a.parse(fs, args)
	if err != nil {
		return err
	}
	if len(positional) > 0 {
		return fmt.Errorf("clear takes no arguments; usage: clear [-f]")
	}

	done, err := a.svc.List(task.ListOptions{Status: task.StatusDone})
	if err != nil {
		return err
	}
	if len(done) == 0 {
		fmt.Fprintln(a.out, "No done tasks to clear")
		return nil
	}

	if !force {
		ok, err := a.confirm(fmt.Sprintf("Delete %s?", plural(len(done), "done task")))
		if err != nil {
			return err
		}
		if !ok {
			fmt.Fprintln(a.out, "Aborted.")
			return nil
		}
	}

	removed, err := a.svc.Clear(task.StatusDone)
	if err != nil {
		return err
	}

	fmt.Fprintf(a.out, "Cleared %s\n", plural(removed, "done task"))
	return nil
}
