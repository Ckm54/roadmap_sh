package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strconv"
)

func handleDelete(subArgs []string) error {
	deleteCmd := flag.NewFlagSet("delete", flag.ContinueOnError)
	deleteCmd.SetOutput(io.Discard)

	if err := deleteCmd.Parse(subArgs); err != nil {
		return err
	}

	args := deleteCmd.Args()
	if len(args) < 1 {
		return errors.New("usage: delete [id]")
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid id %q: must be a number", args[0])
	}

	return svc.Delete(id)
}
