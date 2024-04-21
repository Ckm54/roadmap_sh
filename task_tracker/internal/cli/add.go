package cli

import (
	"errors"
	"flag"
	"io"
)

func handleAdd(subArgs []string) error {
	addCmd := flag.NewFlagSet("add", flag.ContinueOnError)

	addCmd.SetOutput(io.Discard)

	if err := addCmd.Parse(subArgs); err != nil {
		return err
	}

	taskTitle := prepareTaskTitle(addCmd)

	if taskTitle == "" {
		return errors.New("task cannot be empty; usage add [title]")
	}

	return svc.Add(taskTitle)
}
