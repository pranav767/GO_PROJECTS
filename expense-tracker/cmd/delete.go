/*
Copyright © 2025 NAME HERE [pranavppatil767@gmail.com]

*/
package cmd

import (
	"fmt"
	"expense-tracker/expense"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an expense by ID",
	Long: `Delete an expense by ID
	For example:

./expense-tracker delete --id 1`,
	Run: func(cmd *cobra.Command, args []string) {
		id, err := cmd.Flags().GetInt("id")
		if err != nil {
			fmt.Println("Error retrieving ID:", err)
			return
		}
		if id <= 0 {
			fmt.Println("Error: ID must be a positive integer.")
			return
		}
		err = expense.DeleteExpense(id)
		if err != nil {
			fmt.Printf("Error deleting expense with ID %d: %v\n", id, err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	// Add flags to deleteCmd
	deleteCmd.Flags().IntP("id", "i", 0, "ID of the expense to delete (required)")
	// Mark the ID flag as required
	deleteCmd.MarkFlagRequired("id")
	

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deleteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deleteCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
