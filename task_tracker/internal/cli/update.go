package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/ckm54/task_tracker/internal/task"
)

func handleUpdate(subArgs []string) error {
	updateCmd := flag.NewFlagSet("update", flag.ContinueOnError)
	updateCmd.SetOutput(io.Discard)

	if err := updateCmd.Parse(subArgs); err != nil {
		return err
	}

	args := updateCmd.Args()
	if len(args) < 2 {
		return errors.New("usage: update [id] [new title]")
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid ID %q: must be a number", args[0])
	}

	newTitle := strings.TrimSpace(strings.Join(args[1:], " "))
	newTitle = strings.Trim(newTitle, `"'`)
	newTitle = strings.TrimSpace(newTitle)

	if newTitle == "" {
		return errors.New("new title cannot be empty")
	}

	return svc.Update(id, newTitle)
}

func handleUpdateStatus(subArgs []string) error {
	if len(subArgs) < 1 {
		msg := fmt.Sprintf("expected [mark-in-progress] [id] or [mark-done] [id] subcommands")
		return errors.New(msg)
	}

	command := subArgs[0]

	// id, err := strconv.Atoi(subArgs[1])
	// if err != nil {
	// 	return fmt.Errorf("invalid ID %q: must be a number", subArgs[1])
	// }

	markInProgressCmd := flag.NewFlagSet("mark-in-progress", flag.ContinueOnError)
	markInProgressCmd.SetOutput(io.Discard)

	markDoneCmd := flag.NewFlagSet("mark-done", flag.ContinueOnError)
	markDoneCmd.SetOutput(io.Discard)

	switch command {
	case "mark-in-progress":
		if err := markInProgressCmd.Parse(subArgs[1:]); err != nil {
			return err
		}

		remaining := markInProgressCmd.Args()
		if len(remaining) < 1 {
			return errors.New("missing ID: usage: mark-in-progress [id]")
		}

		id, err := strconv.Atoi(remaining[0])
		if err != nil {
			return fmt.Errorf("invalid ID %q: must be a number", subArgs[1])
		}

		return svc.UpdateStatus(id, task.StatusInProgress)
	case "mark-done":
		if err := markDoneCmd.Parse(subArgs[1:]); err != nil {
			return err
		}

		remaining := markDoneCmd.Args()
		if len(remaining) < 1 {
			return errors.New("missing ID: usage: mark-done [id]")
		}

		id, err := strconv.Atoi(remaining[0])
		if err != nil {
			return fmt.Errorf("invalid ID %q: must be a number", subArgs[1])
		}

		return svc.UpdateStatus(id, task.StatusDone)
	default:
		return fmt.Errorf("unknown command %v", subArgs)
	}
}
