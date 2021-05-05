package cli

import (
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/ckm54/task_tracker/internal/constants"
)

var filterOptions map[string]string = map[string]string{
	"todo":        constants.StatusTodo,
	"done":        constants.StatusDone,
	"in-progress": constants.StatusInProgress,
}

func handleList(subArgs []string) error {
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

	if len(tasks) == 0 {
		if filter != "" {
			fmt.Printf("No tasks found matching %q\n", filter)
		} else {
			fmt.Printf("It seems no tasks have been added.\nTo add one:\t add [description]\n")
		}
	} else {
		fmt.Printf("ID\tDescription\tStatus\tCreated\n")
		fmt.Println(strings.Repeat("-", 40))
		for _, task := range tasks {
			fmt.Printf("%d:\t%s\t%s\t%s\n", task.ID, task.Title, task.Status, task.CreatedAt.Format("2006-01-02"))
		}
	}

	return nil
}
