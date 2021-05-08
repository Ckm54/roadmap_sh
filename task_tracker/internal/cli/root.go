package cli

import (
	"flag"
	"fmt"
	"strings"

	"github.com/ckm54/task_tracker/internal/task"
)

type TaskRunner interface {
	Add(title string) (int, error)
	Update(id int, title string) error
	UpdateStatus(id int, status string) error
	Delete(id int) error
	List(filter string) ([]task.TaskEntity, error)
}

var svc TaskRunner

func Init(taskService TaskRunner) {
	svc = taskService
}

func ExecuteLine(line string) error {
	args := strings.Fields(line)
	if len(args) == 0 {
		printGlobalHelp()
		return nil
	}

	subcommand := strings.ToLower(args[0])

	switch subcommand {
	case "add":
		return handleAdd(args[1:])
	case "update":
		return handleUpdate(args[1:])
	case "mark-in-progress", "mark-done":
		return handleUpdateStatus(args)
	case "delete":
		return handleDelete(args[1:])
	case "list":
		return handleList(args[1:])
	case "help":
		printGlobalHelp()
		return nil
	default:
		return fmt.Errorf("Unknown command: %q. Type 'help' for options\n", subcommand)
	}
}

func printGlobalHelp() {
	fmt.Println("Task tracker is a CLI tool to help you manage your tasks")
	fmt.Println("\nUsage:")
	fmt.Println("	task [command]")
	fmt.Println("\nAvailable commands:")
	fmt.Println("	add [title]			Add a new task")
	fmt.Println("	list [filter]			List tasks; filter by status: todo, in-progress, done")
	fmt.Println("	update [id] [title]		Update a task title by id")
	fmt.Println("	mark-in-progress [id]		Mark a task as in-progress")
	fmt.Println("	mark-done [id]			Mark a task as done")
	fmt.Println("	delete [id]			Delete a task by id")
}

func prepareTaskTitle(cmd *flag.FlagSet) string {
	taskTitle := strings.Join(cmd.Args(), " ")

	taskTitle = strings.TrimSpace(taskTitle)
	taskTitle = strings.Trim(taskTitle, `"'`)
	taskTitle = strings.TrimSpace(taskTitle)

	return taskTitle
}
