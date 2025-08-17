/*
Copyright © 2025 Pranav <pranavppatil767@gmail.com>

*/
package cmd

import (
	"fmt"
	"strconv"
	"os"
	"github.com/spf13/cobra"
	"task-tracker/tasks"
)

// markDoneCmd represents the markDone command
var markDoneCmd = &cobra.Command{
	Use:   "mark-done",
	Short: "Mark a task as done by ID",
	Args:  cobra.MinimumNArgs(1), // Ensure at least one argument is provided
	Long: `Mark a task as DONE by ID For example:

./task-tracker mark-done 1`,
	Run: func(cmd *cobra.Command, args []string) {
		id,err := strconv.Atoi(args[0])
		if err != nil {
            fmt.Println("Invalid ID:", args[0])
            os.Exit(1)
        }
		err = tasks.MarkTaskDone(id)
		if err != nil {
            fmt.Println("Error Updating task:", err)
            os.Exit(1)
        }
		fmt.Println("Task Updated successfully:", id)
	},
}

func init() {
	rootCmd.AddCommand(markDoneCmd)
}
