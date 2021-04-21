package cli

import (
	"fmt"
	"strings"
)

type TaskRunner interface {
	Add(title string) error
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
	case "list":
		return HandleList()
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
	fmt.Println("	add			Add a new task")
	fmt.Println("	list		List all your pending tasks")
}
