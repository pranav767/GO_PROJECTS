/*
Copyright © 2025 Pranav <pranavppatil767@gmail.com>

*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"task-tracker/tasks"
	"os"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new task",
	Args:  cobra.MinimumNArgs(1), // Ensure at least one argument is provided
	Long: `Adds a new task to existing list. For example:

./task-tracker add "Buy groceries"`,
	Run: func(cmd *cobra.Command, args []string) {
		task, err := tasks.AddTask(args[0])
		if err != nil {
			fmt.Println("Error adding task:", err)
   			os.Exit(1)
  		}
		fmt.Println("Task added successfully:", task.ID)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
