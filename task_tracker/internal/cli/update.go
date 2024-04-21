package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"
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
