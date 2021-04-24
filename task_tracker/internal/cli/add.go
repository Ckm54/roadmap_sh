package cli

import (
	"errors"
	"flag"
	"io"
	"strings"
)

func handleAdd(subArgs []string) error {
	addCmd := flag.NewFlagSet("add", flag.ContinueOnError)

	addCmd.SetOutput(io.Discard)

	if err := addCmd.Parse(subArgs); err != nil {
		return err
	}

	taskTitle := strings.Join(addCmd.Args(), " ")

	taskTitle = strings.TrimSpace(taskTitle)
	taskTitle = strings.Trim(taskTitle, `"'`)
	taskTitle = strings.TrimSpace(taskTitle)

	if taskTitle == "" {
		return errors.New("task cannot be empty; usage add [title]")
	}

	return svc.Add(taskTitle)
}
