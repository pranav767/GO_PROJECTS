/*
Copyright © 2025 NAME HERE [pranavppatil767@gmail.com]

*/
package cmd

import (
	"fmt"
	"expense-tracker/expense"
	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a specific expense by ID",
	Long: `Update a specific expense by ID For example:

./expense-tracker update --id 1 --description "Updated description" --amount 100.00`,
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
		
		description, _ := cmd.Flags().GetString("description")
		amount, _ := cmd.Flags().GetFloat64("amount")
		if description == "" && amount <= 0 {
			fmt.Println("Error: At least one of description or amount must be provided.")
			return
		}
		
		err = expense.UpdateExpense(id, description, amount)
		if err != nil {
			fmt.Printf("Error updating expense with ID %d: %v\n", id, err)
			return
		}
		fmt.Println("Expense updated successfully!")
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)

	// Add flags to updateCmd
	updateCmd.Flags().IntP("id", "i", 0, "ID of the expense to update (required)")
	updateCmd.Flags().StringP("description", "d", "", "New description of the expense")
	updateCmd.Flags().Float64P("amount", "a", 0, "New amount of the expense")

	// Mark the ID flag as required
	updateCmd.MarkFlagRequired("id")
	// Mark the description and amount flags as optional, but provide a message if they are not set
	//updateCmd.MarkFlagRequired("description")
	//updateCmd.MarkFlagRequired("amount")
	// You can also add validation to ensure that at least one of the description or amount is provided
	//updateCmd.Flags().MarkDeprecated("description", "Please provide a new description if you want to update it.")
	//updateCmd.Flags().MarkDeprecated("amount", "Please provide a new amount if you want to update it.")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// updateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// updateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
