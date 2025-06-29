package expense

import (
	"fmt"
	"encoding/json"
	"os"
	"time"
)

type Expense struct {
	ID		  	int    `json:"id"`
	Description string    `json:"description"`
	Amount     	float64   `json:"amount"`
	Date       	time.Time `json:"date"`
}

const expenseFile = "expenses.json"

// Load existing expenses from the JSON file
// This is done to keep ID unique and to load existing expenses
func LoadExpenses() ([]Expense, error) {
	//Read the file
	data, err := os.ReadFile(expenseFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []Expense{}, nil // If file does not exist, return empty slice
		}
		return nil, fmt.Errorf("error reading expenses file: %v", err)
	}
	// Unmarshal the JSON data into a slice of Expense
	var expenses []Expense
	err = json.Unmarshal(data, &expenses)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling expenses: %v", err)
	}
	return expenses, nil
}

// Add a new expense to the JSON file
func AddExpense(Description string, Amount float64) (error,int) {
	expenses, err := LoadExpenses()
	if err != nil {
		return err, 0
	}
	// Create a new expense with a unique ID
	newId := 1
	for _, expense := range expenses {
		if expense.ID >= newId {
			newId = expense.ID +1
		}
	}
	currentTime := time.Now()
	newExpense := Expense{
		ID:		  newId,
		Description: Description,
		Amount:      Amount,
		Date:		currentTime,
	}
	expenses = append(expenses, newExpense)
	// Marshal the updated expenses slice to JSON
	data, err := json.MarshalIndent(expenses, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling expenses: %v", err),0
	}
	err = os.WriteFile(expenseFile, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing expenses to file: %v", err),0
	}
	return nil, newId
}

// Summary of expenses
func Summary(month ...int) (float64, error) {
	expenses, err := LoadExpenses()
	if err != nil {
		return 0, fmt.Errorf("error loading expenses: %v", err)
	}
	totalAmount := 0.0
	if (month != nil && len(month) > 0) {
		// If month is provided, filter expenses by month
		for _, expense := range expenses {
			if expense.Date.Month() == time.Month(month[0]) {
				totalAmount += expense.Amount
			}
		} 
	} else { 
		for _, expense := range expenses{
			totalAmount += expense.Amount
		}
	}
	return totalAmount, nil
}

// DeleteExpense deletes an expense by ID
func DeleteExpense(ID int) error {
	expense, err := LoadExpenses()
	if err != nil {
		return fmt.Errorf("error loading expenses: %v", err)
	}
	// Check if the expense with the given ID exists
	found := false
	var updatedExpenses []Expense
	for _, e := range expense {
		if e.ID != ID {
			updatedExpenses = append(updatedExpenses, e)
		} else {
			found = true
			fmt.Printf("Deleted Expense: ID=%d, Description=%s, Amount=%.2f, Date=%s\n", e.ID, e.Description, e.Amount, e.Date.Format("2006-01-02"))
		}
	}
	if !found {
		return fmt.Errorf("expense with ID %d not found", ID)
	}
	data, err := json.MarshalIndent(updatedExpenses, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling expenses: %v", err)
	}
	err = os.WriteFile(expenseFile, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing expenses to file: %v", err)
	}
	return nil
}

// UpdateExpense updates an existing expense by ID
func UpdateExpense(ID int, Description string, Amount float64) error {
	expenses,err := LoadExpenses()
	if err != nil {
		return fmt.Errorf("error loading expenses: %v", err)
	}
	// Check if the expense with the given ID exists
	found := false
	var updatedExpenses []Expense
	for _, e := range expenses {
		if e.ID != ID {
			updatedExpenses = append(updatedExpenses, e)
		} else {
			found = true
			// Update the expense details
			if Description == "" {
				Description = e.Description // Keep the old description if new one is empty
			} else if Amount <= 0 {
				Amount = e.Amount // Keep the old amount if new one is not provided
			}
			e.Description = Description
			e.Amount = Amount
			e.Date = time.Now() // Update the date to current time
			updatedExpenses = append(updatedExpenses, e)
		}
	}
	if !found {
		return fmt.Errorf("expense with ID %d not found", ID)
	}
	data, err := json.MarshalIndent(updatedExpenses, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling expenses: %v", err)
	}
	err = os.WriteFile(expenseFile, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing expenses to file: %v", err)
	}
	return nil
}