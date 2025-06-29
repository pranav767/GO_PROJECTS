/*
Copyright © 2025 NAME HERE [pranavppatil767@gmail.com]

*/
package cmd

import (
	"fmt"
	"expense-tracker/expense"
	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new expense",
	Long: `Add a new expense. For example:

./expense-tracker add --description "Buy groceries" --amount 50.00`,
	Run: func(cmd *cobra.Command, args []string) {
		// Description and amount are required flags, subcommands are written as a function and flags are accessed using cmd.Flags().Get<Type>()
		// If the flags are not provided, Cobra will automatically handle the error and display usage information.
		description, _ := cmd.Flags().GetString("description")
		amount, _ := cmd.Flags().GetFloat64("amount")
		if description == "" || amount <= 0 {
			fmt.Println("Error: Description and amount are required.")
			return
		}
		err,ID := expense.AddExpense(description, amount)
		if err != nil {
			fmt.Printf("Error adding expense: %v\n", err)
			return
		}
		fmt.Println("Expense added successfully! ID:", ID)
		// You can also print the expense details if needed
		// fmt.Printf("Description: %s, Amount: %.2f\n", description, amount)
		// For now, just print a message
		fmt.Println("add called")
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	// Add flags to addCmd
	addCmd.Flags().StringP("description", "d", "", "Description of the expense (required)")
	addCmd.Flags().Float64P("amount", "a", 0, "Amount of the expense (required)")

	//Mark flags as required
	addCmd.MarkFlagRequired("description")
	addCmd.MarkFlagRequired("amount")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
