/*
Copyright © 2025 Pranav <pranavppatil767@gmail.com>

*/
package cmd

import (
	"fmt"
	"task-tracker/tasks"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tasks",
	Long: `List all tasks. For example:

./task-tracker list`,
	Run: func(cmd *cobra.Command, args []string) {
		task, err := tasks.LoadTasks()
		if err != nil {
			fmt.Println("Error loading tasks:", err)
		}
		for _, t := range task {
			fmt.Printf("ID: %d, Description: %s, Status: %s, Created At: %s, Updated At: %s\n", t.ID, t.Description, t.Status, t.CreatedAt, t.UpdatedAt)
		}
	},
}

var listDoneCmd = &cobra.Command{
	Use:   "done",
	Short: "List all tasks that are marked as done",
	Long: `A Subcommand for list which lists out all tasks marked as done For example:

./task-tracker list done`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("list done called")
		task,err := tasks.LoadTasks()
		if err != nil {
			fmt.Println("Error loading tasks:", err)
		}
		for _, t := range task {
			if t.Status == "DONE" {
				fmt.Printf("ID: %d, Description: %s, Status: %s, Created At: %s, Updated At: %s\n",t.ID, t.Description, t.Status, t.CreatedAt, t.UpdatedAt)
			}
		}
	},
}

var listToDoCmd = &cobra.Command{
	Use:   "todo",
	Short: "List all tasks that are marked as To-Do",
	Long: `A Subcommand for list which lists out all tasks marked as To-Do For example:

./task-tracker list todo`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("list todo called")
		task,err := tasks.LoadTasks()
		if err != nil {
			fmt.Println("Error loading tasks:", err)
		}
		for _, t := range task {
			if t.Status == "To-DO" {
				fmt.Printf("ID: %d, Description: %s, Status: %s, Created At: %s, Updated At: %s\n",t.ID, t.Description, t.Status, t.CreatedAt, t.UpdatedAt)
			}
		}
	},
}

var listInProgressCmd = &cobra.Command{
	Use:   "in-progress",
	Short: "List all tasks that are marked as In-Progress",
	Long: `A Subcommand for list which lists out all tasks marked as IN-PROGRESS. For example:

./task-tracker list in-progress`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("list in-progress called")
		task,err := tasks.LoadTasks()
		if err != nil {
			fmt.Println("Error loading tasks:", err)
		}
		for _, t := range task {
			if t.Status == "IN-PROGRESS" {
				fmt.Printf("ID: %d, Description: %s, Status: %s, Created At: %s, Updated At: %s\n",t.ID, t.Description, t.Status, t.CreatedAt, t.UpdatedAt)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.AddCommand(listDoneCmd)
	listCmd.AddCommand(listToDoCmd)
	listCmd.AddCommand(listInProgressCmd)
}
