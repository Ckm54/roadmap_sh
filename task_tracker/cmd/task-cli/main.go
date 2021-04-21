package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ckm54/task_tracker/internal/cli"
	"github.com/ckm54/task_tracker/internal/storage"
	"github.com/ckm54/task_tracker/internal/task"
)

var STORAGE_FILE = "tasks.json"
var PROMPT = "task-cli>> "

func main() {
	store := storage.NewJSONStore(STORAGE_FILE)
	taskService := task.NewService(store)

	cli.Init(taskService)

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print(PROMPT)

	for scanner.Scan() {
		input := scanner.Text()
		trimmed := strings.TrimSpace(input)

		if trimmed == "exit" || trimmed == "quit" {
			break
		}

		if trimmed == "" {
			fmt.Print(PROMPT)
			continue
		}

		err := cli.ExecuteLine(trimmed)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}

		fmt.Print(PROMPT)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Scanner error: %v\n", err)
		os.Exit(1)
	}

}
