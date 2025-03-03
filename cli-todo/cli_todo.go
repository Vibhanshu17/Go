package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofrs/flock"
)

/*
1. creat a new task
2. list all tasks
3. mark a task complete
4. delete a task
*/

const tasksFile = "tasks.csv"

func generateID() int {
	file, err := os.OpenFile(tasksFile, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return 1
	}
	defer file.Close()

	size, err := file.Seek(0, io.SeekEnd)
	if err != nil || size == 0 {
		return 1
	}

	buf := make([]byte, min(size, 1024))
	start := max(size-int64(len(buf)), 0)

	_, err = file.Seek(start, io.SeekStart)
	if err != nil {
		return 1
	}

	n, err := file.Read(buf)
	if err != nil || err != io.EOF {
		return 1
	}
	buf = buf[:n]

	lastNewLine := bytes.LastIndex(buf, []byte("\n"))
	if lastNewLine == -1 {
		return 1
	}

	lastLine := string(buf[lastNewLine+1:])
	reader := csv.NewReader(strings.NewReader(lastLine))
	row, err := reader.Read()
	if err != nil || len(row) == 0 {
		return 1
	}

	lastID, _ := strconv.Atoi(row[0])
	return lastID + 1
}

type Task struct {
	ID           int
	task_name    string ""
	completed    bool
	created_at   time.Time
	completed_at time.Time
}

func (t *Task) String() string {
	return fmt.Sprintf("%+v", *t)
}

func (t *Task) Set(value string) error {
	t.ID = generateID()
	t.task_name = value
	t.completed = false
	t.created_at = time.Now()
	t.completed_at = time.Time{}
	return nil
}

func readFile(filename string, mode int, perm os.FileMode) (*os.File, error) {
	file, err := os.OpenFile(filename, mode, perm)
	if err != nil {
		fmt.Printf("Failed to open file: %v\n", err)
		return nil, err
	}
	return file, nil
}

func createTask(task Task) error {
	lock := flock.New(tasksFile + ".lock")
	err := lock.Lock()
	if err != nil {
		return fmt.Errorf("failed to acquire lock.\nERROR: %v", err)
	}
	defer lock.Unlock()

	fileStats, err := os.Stat(tasksFile)
	isNewFile := os.IsNotExist(err) || fileStats.Size() == 0

	file, err := readFile(tasksFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file.\nERROR: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// write header if file is empty
	if isNewFile {
		if err := writer.Write([]string{"ID", "Task", "Completed", "Created At", "Completed At"}); err != nil {
			return fmt.Errorf("failed to write headers to file.\nERROR: %v", err)
		}
	}
	// write task to file
	row := []string{
		strconv.Itoa(task.ID),
		task.task_name,
		strconv.FormatBool(task.completed),
		task.created_at.Format(time.RFC3339),
		task.completed_at.Format(time.RFC3339),
	}
	if err := writer.Write(row); err != nil {
		return fmt.Errorf("failed to write to file.\nERROR: %v", err)
	}
	return nil
}

func listAllTasks() error {
	lock := flock.New(tasksFile + ".lock")
	err := lock.Lock()
	if err != nil {
		return fmt.Errorf("failed to acquire lock.\nERROR: %v", err)
	}
	defer lock.Unlock()
	file, err := readFile(tasksFile, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file.\nERROR: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read file.\nERROR: %v", err)
	}

	fmt.Printf("\n%-5s %-30s %-10s %-25s %-25s\n", "ID", "Task", "Completed", "Created At", "Completed At")
	fmt.Println(strings.Repeat("-", 100))
	for i, row := range rows {
		if i == 0 {
			continue
		}
		fmt.Printf("\n%-5s %-30s %-10s %-25s %-25s\n", row[0], row[1], row[2], row[3], row[4])
	}
	return nil
}

func markComplete(ID int) error {
	file, err := readFile(tasksFile, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to read file.\nERROR: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read file.\nERROR: %v", err)
	}
	for _, row := range rows {
		if row[0] == strconv.Itoa(ID) {
			row[2] = "true"
			row[4] = time.Now().Format(time.RFC3339)
			break
		}
	}
	file.Truncate(0)
	file.Seek(0, 0)
	writer := csv.NewWriter(file)
	if err = writer.WriteAll(rows); err != nil {
		return fmt.Errorf("failed to write to file.\nERROR: %v", err)
	}
	writer.Flush()
	return nil
}

func deleteTask(ID int) error {
	file, err := readFile(tasksFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to read file.\nERROR: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read file.\nERROR: %v", err)
	}

	newRows := [][]string{}
	for _, row := range rows {
		if row[0] != strconv.Itoa(ID) {
			newRows = append(newRows, row)
		}
	}

	file.Truncate(0)
	file.Seek(0, 0)
	writer := csv.NewWriter(file)
	if err = writer.WriteAll(newRows); err != nil {
		return fmt.Errorf("failed to write to file.\nERROR: %v", err)
	}
	writer.Flush()
	return nil
}

func main() {
	option := flag.String("option", "list-all", "option to perform on tasks\nChoose from: 'create', 'list-all', 'mark-complete', 'delete'")
	task := &Task{}
	var ID int

	flag.Var(task, "task", "task to create")
	flag.IntVar(&ID, "id", 0, "task id to mark complete/delete")
	flag.Parse()

	switch *option {
	case "create":
		if task.task_name == "" {
			fmt.Println("task name cannot be empty")
			return
		}
		createTask(*task)

	case "list-all":
		listAllTasks()

	case "mark-complete":
		markComplete(ID)

	case "delete":
		deleteTask(ID)

	default:
		fmt.Println("Invalid option\nChoose from: 'create', 'list-all', 'mark-complete', 'delete'")
	}

}
