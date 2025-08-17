/*
Copyright © 2025 Pranav <pranavppatil767@gmail.com>

*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)



// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "task-tracker",
	Short: "Task Tracker CLI",
	Long: `Task Tracker CLI is a command-line tool to manage your tasks efficiently.
It allows you to add, delete, and manage tasks with ease. You can mark tasks as done, in-progress, or to-do.
For example:

./task-tracker add "Buy groceries"
./task-tracker delete 1
./task-tracker mark-done 2
./task-tracker list
./task-tracker list done
./task-tracker list todo
./task-tracker list in-progress`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}


