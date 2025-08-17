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

// markInProgressCmd represents the markInProgress command
var markInProgressCmd = &cobra.Command{
	Use:   "mark-in-progress",
	Short: "Mark a task as in-progress by ID",
	Long: `Mark a task as IN-PROGRESS For example:

./task-tracker mark-in-progress 1`,
	Run: func(cmd *cobra.Command, args []string) {
		id, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("Invalid ID:", args[0])
			os.Exit(1)
		}
		err = tasks.MarkTaskInProgress(id)
		if err != nil {
			fmt.Println("Error Updating task:", err)
			os.Exit(1)
		}
		fmt.Println("Task Updated successfully:", id)
	},
}

func init() {
	rootCmd.AddCommand(markInProgressCmd)
}
