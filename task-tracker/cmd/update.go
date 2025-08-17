/*
Copyright © 2025 Pranav <pranavppatil767@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"strconv"
	"task-tracker/tasks"

	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a task by ID",
	Args:  cobra.MinimumNArgs(2),
	Long: `Update a task by ID For example:

./task-tracker update 1 "New Task Description"`,
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("Invalid ID:", args[0])
			os.Exit(1)
		}
		err = tasks.UpdateTask(id, args[1])
		if err != nil {
			fmt.Println("Error Updating task:", err)
			os.Exit(1)
		}
		fmt.Println("Task Updated successfully:", id)

	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
