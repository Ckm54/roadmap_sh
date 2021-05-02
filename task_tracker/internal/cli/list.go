package cli

import (
	"flag"
	"fmt"
	"io"

	"github.com/ckm54/task_tracker/internal/constants"
)

var filterOptions map[string]string = map[string]string{
	"todo":        constants.StatusTodo,
	"done":        constants.StatusDone,
	"in-progress": constants.StatusInProgress,
}

func HandleList(subArgs []string) error {
	listCmd := flag.NewFlagSet("list", flag.ContinueOnError)
	listCmd.SetOutput(io.Discard)

	if err := listCmd.Parse(subArgs); err != nil {
		return err
	}

	args := listCmd.Args()
	var filter string

	if len(args) > 1 {
		return fmt.Errorf("too many filter options. Just need one")
	}

	if len(args) == 1 {
		option := args[0]
		if opt, ok := filterOptions[option]; ok {
			filter = opt
		} else {
			return fmt.Errorf("unknown status filter %s", option)
		}
	}
	tasks, err := svc.List(filter)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		fmt.Println(task.String())
	}

	return nil
}
