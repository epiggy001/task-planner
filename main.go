package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/satori/go.uuid"
)

const (
	maxTaskNum = 5
	maxDelay   = 3600 * 24 * 2 // 2 day
)

var (
	home string = os.Getenv("HOME")
	fp   string = filepath.Join(home, "/.task-planner")
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("You need a sub command like add, rm, start or ls")
		os.Exit(2)
	}

	var err error
	switch os.Args[1] {
	case "add":
		if len(os.Args) > 2 {
			err = add(strings.Join(os.Args[2:], " "))
		}
	case "pop":
		err = rm(0)
	case "rm":
		if len(os.Args) > 2 {
			var i int
			i, err = strconv.Atoi(os.Args[2])
			if err == nil {
				err = rm(i)
			}
		}
	case "ls":
		err = ls()
	default:
		fmt.Printf("%q is not valid command.\n", os.Args[1])
		os.Exit(2)
	}
	if err != nil {
		exit(err)
	}
}

type Task struct {
	UUID        string
	Description string
	CreateTime  time.Time
}

type Tasks []*Task

func (tasks Tasks) Len() int           { return len(tasks) }
func (tasks Tasks) Swap(i, j int)      { tasks[i], tasks[j] = tasks[j], tasks[i] }
func (tasks Tasks) Less(i, j int) bool { return tasks[i].CreateTime.Unix() > tasks[j].CreateTime.Unix() }

func exit(err error) {
	log.Fatalf("tp fail to run: %s\n", err)
}

func readTasks() ([]*Task, error) {
	var tasks []*Task
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		f, err := os.Create(fp)
		if err != nil {
			return nil, err
		}
		f.Close()
		return tasks, nil
	}

	bytes, err := ioutil.ReadFile(fp)
	if err != nil {
		return nil, err
	}

	if len(bytes) == 0 {
		return tasks, err
	}

	err = json.Unmarshal(bytes, &tasks)
	if err != nil {
		return nil, err
	}
	sort.Sort(Tasks(tasks))
	return tasks, nil
}

func add(desc string) error {
	tasks, err := readTasks()
	if err != nil {
		return err
	}

	if len(tasks) > maxDelay {
		color.Red("You alread have %d tasks. Clear them first !", maxDelay)
		return nil
	}

	tasks = append(tasks, &Task{
		UUID:        uuid.NewV4().String(),
		Description: desc,
		CreateTime:  time.Now(),
	})

	err = writeTasks(tasks)
	if err != nil {
		return err
	}
	return ls()
}

func writeTasks(tasks []*Task) error {
	bytes, err := json.Marshal(tasks)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fp, bytes, 0664)
}

func ls() error {
	tasks, err := readTasks()
	if err != nil {
		return err
	}
	if len(tasks) == 0 {
		fmt.Println("You do not have any tasks. You can take some coke now.")
		return nil
	}
	for i, task := range tasks {
		output(i, task)
	}
	return nil
}

func rm(i int) error {
	tasks, err := readTasks()
	if err != nil {
		return err
	}
	tasks = append(tasks[:i], tasks[i+1:]...)
	err = writeTasks(tasks)
	if err != nil {
		return err
	}
	return ls()
}

func output(index int, task *Task) {
	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	createdDuration := time.Since(task.CreateTime).Hours()
	durationString := fmt.Sprintf("%.2fh", createdDuration)
	if createdDuration >= maxDelay {
		durationString = red(durationString)
	} else {
		durationString = green(durationString)
	}

	fmt.Printf("%d. %s (%s)\n", index, task.Description, durationString)
}
