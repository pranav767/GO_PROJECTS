/*
Copyright © 2025 NAME HERE [pranavppatil767@gmail.com]

*/
package cmd

import (
	"fmt"
	"expense-tracker/expense"
	"github.com/spf13/cobra"
)

// summaryCmd represents the summary command
var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short:  "Summarize all expenses.",
	Long: `Summarize all expenses For example:

./expense-tracker summary
 Summariza all expenses for a specific month.
 
 ./expense-tracker summary --month 7`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("summary called")
		month, _ := cmd.Flags().GetInt("month")
		
		if month == 0 {
			expenses,err := expense.Summary()
			if err != nil {
				fmt.Printf("Error generating summary: %v\n", err)
				return
			}
			fmt.Printf("Summary of Expenses: %.2f\n", expenses)
		} else {
			if month < 1 || month > 12 {
			fmt.Println("Error: Month must be between 1 and 12.")
			return
			}
			expenses,err := expense.Summary(month)
			if err != nil {
				fmt.Printf("Error generating summary: %v\n", err)
				return
			}
			fmt.Printf("Summary of Expenses for month %d: %.2f\n", month, expenses)
		}
		//if expenses == 0 {
		//	fmt.Println("No expenses found.")
		//	return
		//}
	},
}

func init() {
	rootCmd.AddCommand(summaryCmd)
	summaryCmd.Flags().IntP("month", "m", 0, "Month for which to summarize expenses (1-12). If not provided, summarizes all expenses.")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// summaryCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// summaryCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
