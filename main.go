package main

import (
	"encoding/json"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"strconv"
	"time"
)

var userDir string

type Tasks struct {
	Task []Task `json:"tasks"`
}

type Task struct {
	ID        int       `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	DoneAt    time.Time `json:"done_at"`
	Name      string    `json:"name"`
	Label     string    `json:"label"`
	Completed bool      `json:"completed"`
}

func getCountOfTasks() int {
	file, err := os.ReadFile(userDir + "/todoCLI/tasks.json")

	data := Tasks{}

	err = json.Unmarshal(file, &data)

	if err != nil {
		return 0
	}

	return len(data.Task)
}

func taskExists(id int) bool {
	for _, task := range getAllTasks().Task {
		if task.ID == id {
			return true
		}
	}
	return false
}

func deleteAllTasks() error {
	_ = os.WriteFile(userDir+"/todoCLI/tasks.json", []byte(`{"tasks": []}`), 0644)
	return nil
}

func deleteTask(id int) bool {
	allTasks := getAllTasks()

	for i, task := range allTasks.Task {
		if task.ID == id {
			allTasks.Task = append(allTasks.Task[:i], allTasks.Task[i+1:]...)
		}
	}

	file, err := json.MarshalIndent(allTasks, "", " ")

	if err != nil {
		return false
	}

	_ = os.WriteFile(userDir+"/todoCLI/tasks.json", file, 0644)

	reassignIDs()

	return true
}

func getAllTasks() Tasks {
	file, err := os.ReadFile(userDir + "/todoCLI/tasks.json")

	allTasks := Tasks{}

	err = json.Unmarshal(file, &allTasks)

	if err != nil {
		return allTasks
	}

	return allTasks
}

func writeTaskToFile(task Task) error {
	data := getAllTasks()

	data.Task = append(data.Task, task)

	file, err := json.MarshalIndent(data, "", " ")

	if err != nil {
		return err
	}

	_ = os.WriteFile(userDir+"/todoCLI/tasks.json", file, 0644)

	return nil
}

func reassignIDs() error {
	allTasks := getAllTasks()

	for i, task := range allTasks.Task {
		task.ID = i + 1
		allTasks.Task[i] = task
	}

	file, err := json.MarshalIndent(allTasks, "", " ")

	if err != nil {
		return err
	}

	_ = os.WriteFile(userDir+"/todoCLI/tasks.json", file, 0644)

	return nil
}

func printTasks() error {
	for _, task := range getAllTasks().Task {
		var checkedASCII string
		if task.Completed {
			checkedASCII = "[✓]"
		} else {
			checkedASCII = "[ ]"
		}

		log.Printf("%d\t%s - %s", task.ID, checkedASCII, task.Name)
	}
	return nil
}

func main() {
	// Remove the timestamp from the log output
	log.SetFlags(0)

	// Get the user's home directory
	userDir, _ = os.UserHomeDir()

	// Create the directory for the tasks.json file
	_ = os.Mkdir(userDir+"/todoCLI", 0755)

	// Create the tasks.json file if it doesn't exist
	_, err := os.Stat(userDir + "/todoCLI/tasks.json")

	if os.IsNotExist(err) {
		_ = os.WriteFile(userDir+"/todoCLI/tasks.json", []byte(`{"tasks":[]}`), 0644)
	}

	// Create the CLI app
	app := &cli.App{
		Name:  "todoCLI",
		Usage: "This is a Todo list CLI application",
		Commands: []*cli.Command{
			{
				Name:    "add",
				Aliases: []string{"a"},
				Usage:   "Add a new task to the list",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Aliases: []string{"n"}, Usage: "Name of the task", Required: false},
					&cli.StringFlag{Name: "label", Aliases: []string{"l"}, Usage: "Label of the task", Required: false},
				},
				Action: func(c *cli.Context) error {
					taskName := c.String("name")

					if taskName == "" {
						// Get the task name as argument
						taskName = c.Args().Get(0)

						if taskName == "" {
							return cli.Exit("No task provided", 1)
						}
					}

					// Create Task
					newTask := Task{
						ID:        getCountOfTasks() + 1,
						CreatedAt: time.Now(),
						DoneAt:    time.Time{},
						Name:      taskName,
						Label:     c.String("label"),
						Completed: false,
					}

					writeTaskToFile(newTask)

					return nil
				},
			},
			{
				Name:    "remove",
				Aliases: []string{"r"},
				Usage:   "Remove a task from the list by its ID",
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "all", Aliases: []string{"a"}, Usage: "Remove all tasks", Required: false},
				},
				Action: func(c *cli.Context) error {
					taskIDStr := c.Args().Get(0)

					if taskIDStr == "" {
						return cli.Exit("No task ID provided", 1)
					}

					if taskIDStr == "all" {
						err := deleteAllTasks()
						if err != nil {
							return err
						}
						return nil
					}

					taskID, err := strconv.Atoi(taskIDStr)

					if err != nil {
						return cli.Exit("Invalid task ID provided", 1)
					}

					if !taskExists(taskID) {
						return cli.Exit("Task does not exist", 1)
					}

					if deleteTask(taskID) {
						log.Printf("Task %d has been deleted", taskID)
					} else {
						return cli.Exit("Error deleting task", 1)
					}

					return nil
				},
			},
			{
				Name:    "list",
				Aliases: []string{"l"},
				Usage:   "List all tasks",
				Action: func(c *cli.Context) error {
					if getCountOfTasks() == 0 {
						return cli.Exit("No tasks found, use 'add' command to start creating tasks", 1)
					}
					// print header
					err := printTasks()

					if err != nil {
						return err
					}
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
