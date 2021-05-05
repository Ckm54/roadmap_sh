package cli

import (
	"errors"
	"flag"
	"fmt"
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

	id, err := svc.Add(taskTitle)
	if err != nil {
		return err
	}

	fmt.Printf("Task added successfully (ID: %d)\n", id)
	return nil
}
