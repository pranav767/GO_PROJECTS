/*
Copyright © 2025 Pranav <pranavppatil767@gmail.com>


*/
package cmd

import (
	"fmt"
	"task-tracker/tasks"
	"github.com/spf13/cobra"
	"strconv"
	"os"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a task by ID",
	Args:  cobra.MinimumNArgs(1), // Ensure at least one argument is provided
	Long: `Delete a task by ID. For example:

./task-tracker delete 1`,
	Run: func(cmd *cobra.Command, args []string) {
		id,err := strconv.Atoi(args[0])
		if err != nil {
            fmt.Println("Invalid ID:", args[0])
            os.Exit(1)
        }
		err = tasks.DeleteTask(id)
		if err != nil {
            fmt.Println("Error deleting task:", err)
            os.Exit(1)
        }
		fmt.Println("Task deleted successfully:", id)
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
