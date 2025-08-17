// Shared task logic

package tasks
import (
	"fmt"
	"encoding/json"
	"os"
	"time"
)

// structure of the task storage file
type Task struct {
	ID		  		int    `json:"id"`
	Description	    string `json:"description"`
	Status          string `json:"status"`
	CreatedAt	    string `json:"created_at"`
	UpdatedAt	    string `json:"updated_at"`
}

// there will be a file to store tasks
const storageFile = "tasks.json"

// LoadTasks loads tasks from the storage file
func LoadTasks() ([]Task, error) {
	data, err := os.ReadFile(storageFile)
	if err != nil {
		// If the file does not exist, return an empty slice
		if os.IsNotExist(err) {
			return []Task{}, nil
		}
		fmt.Println("Error reading tasks file:", err)
		return nil, err
	}
	//  If the file is empty, return an empty slice
	if len(data) == 0 {
		return []Task{}, nil // Return empty slice if file is empty
	}
	var tasks []Task
	err = json.Unmarshal(data, &tasks)
	if err != nil {
		fmt.Println("Error unmarshalling tasks:", err)
		return nil, err
	}
	return tasks,nil
}

// AddTask adds a new task to the storage file
func AddTask(Description string) (Task, error) {
	// get existing tasks
	tasks, err := LoadTasks()
	if err != nil {
		return Task{}, err
	}
	// Create a new task
	newID := 1
	for _, t := range tasks {
		if t.ID >=newID {
			newID = t.ID +1
		}
	}
	currentTime := time.Now().Format(time.RFC3339)
	task := Task{
		ID: 		newID,
		Description: Description,
		Status:     "TO-DO",
		CreatedAt:  currentTime,
		UpdatedAt: currentTime,
	}
	tasks = append(tasks, task)
	// We have entire tasks list, rewrite this to the file
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return Task{}, err
	}
	err = os.WriteFile(storageFile, data, 0644)
	if err != nil {
		return Task{}, err
	}
	return task, nil
}

// DeleteTask deletes a task by ID
func DeleteTask(id int) error {
	tasks, err := LoadTasks()
	newTasks := make([]Task, 0, len(tasks))
	for _,t := range tasks{
		if t.ID !=id {
			newTasks = append(newTasks, t)
		}
	}
	data, err := json.MarshalIndent(newTasks, "", "  ")
	if err != nil {
		return err
	}
	err = os.WriteFile(storageFile, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

// Update existing task by ID
func UpdateTask(id int, Description string) error {
	tasks, err := LoadTasks()
	newTasks := make([]Task, 0, len(tasks))
	for _,t := range tasks{
		if t.ID !=id {
			newTasks = append(newTasks, t)
		} else {
			// Update the task
			t.Description = Description
			t.UpdatedAt = time.Now().Format(time.RFC3339)
			newTasks = append(newTasks, t)
		}
	}
	data, err := json.MarshalIndent(newTasks, "", "  ")
	if err != nil {
		return err
	}
	err = os.WriteFile(storageFile, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

//Mark a task as done by ID
func MarkTaskDone(id int) error {
	tasks, err := LoadTasks()
	newTasks := make([]Task, 0, len(tasks))
	for _,t := range tasks{
		if t.ID !=id {
			newTasks = append(newTasks, t)
		} else {
			// Update the task as Done
			t.Status = "DONE"
			t.UpdatedAt = time.Now().Format(time.RFC3339)
			newTasks = append(newTasks, t)
		}
	}
	data, err := json.MarshalIndent(newTasks, "", "  ")
	if err != nil {
		return err
	}
	err = os.WriteFile(storageFile, data, 0644)
	if err != nil {
		return err
	}
	return nil
} 

//Mark a task as done by ID
func MarkTaskInProgress(id int) error {
	tasks, err := LoadTasks()
	newTasks := make([]Task, 0, len(tasks))
	for _,t := range tasks{
		if t.ID !=id {
			newTasks = append(newTasks, t)
		} else {
			// Update the task as Done
			t.Status = "IN-PROGRESS"
			t.UpdatedAt = time.Now().Format(time.RFC3339)
			newTasks = append(newTasks, t)
		}
	}
	data, err := json.MarshalIndent(newTasks, "", "  ")
	if err != nil {
		return err
	}
	err = os.WriteFile(storageFile, data, 0644)
	if err != nil {
		return err
	}
	return nil
} 