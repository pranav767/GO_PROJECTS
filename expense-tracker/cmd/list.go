/*
Copyright © 2025 NAME HERE [pranavppatil767@gmail.com]

*/
package cmd

import (
	"fmt"
	"expense-tracker/expense"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all expenses",
	Long: `List of all expenses. For example:

./expense-tracker list`,
	Run: func(cmd *cobra.Command, args []string) {
		expense, err := expense.LoadExpenses()
		if err != nil {
			fmt.Printf("Error loading expenses: %v\n", err)
			return
		}
		if len(expense) == 0 {
			fmt.Println("No expenses found.")
			return
		}
		fmt.Println("Expenses:")
		fmt.Println("ID\tDescription\tAmount\t\tDate")
		for _,e := range expense {
			fmt.Printf("%d\t%s\t\t%.2f\t\t%s\n", e.ID, e.Description, e.Amount, e.Date.Format("2006-01-02 15:04:05"))
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
