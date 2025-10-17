# Expense Tracker CLI

A command-line expense tracking application built with Go and Cobra CLI framework. Track your daily expenses with easy-to-use commands.

> Project idea from: https://roadmap.sh/projects/expense-tracker

## Features

- ✅ Add new expenses with description and amount
- ✅ List all expenses with detailed information
- ✅ Delete expenses by ID
- ✅ Update existing expenses
- ✅ Generate expense summaries (total or by month)
- ✅ Data persistence with JSON file storage

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd expense-tracker
```

2. Build the application:
```bash
go build -o expense-tracker
```

3. (Optional) Make it globally accessible:
```bash
sudo mv expense-tracker /usr/local/bin/
```

## Usage

### Basic Commands

#### Add an Expense
```bash
# Add expense with description and amount
./expense-tracker add --description "Lunch" --amount 15.50

# Short form
./expense-tracker add -d "Coffee" -a 4.25
```

#### List All Expenses
```bash
./expense-tracker list
```
**Output:**
```
ID: 2, Description: gym, Amount: 2000.00, Date: 2025-06-29
ID: 3, Description: , Amount: 250.00, Date: 2025-06-29
ID: 4, Description: bought some cool fidget, Amount: 250.00, Date: 2025-06-29
```

#### Delete an Expense
```bash
# Delete expense by ID
./expense-tracker delete --id 3

# Short form
./expense-tracker delete -i 3
```

#### Update an Expense
```bash
# Update expense description and amount
./expense-tracker update --id 2 --description "Monthly gym membership" --amount 2500

# Short form
./expense-tracker update -i 2 -d "New description" -a 100.00
```

#### Generate Summary
```bash
# Summary of all expenses
./expense-tracker summary

# Summary for specific month (1-12)
./expense-tracker summary --month 6
./expense-tracker summary -m 12
```

### Command Reference

| Command | Description | Flags |
|---------|-------------|-------|
| `add` | Add a new expense | `--description, -d` (string)<br>`--amount, -a` (float64) |
| `list` | List all expenses | None |
| `delete` | Delete an expense by ID | `--id, -i` (int) |
| `update` | Update an existing expense | `--id, -i` (int)<br>`--description, -d` (string)<br>`--amount, -a` (float64) |
| `summary` | Show expense summary | `--month, -m` (int, optional) |
| `help` | Show help information | None |

### Examples

```bash
# Add various expenses
./expense-tracker add -d "Groceries" -a 85.50
./expense-tracker add -d "Gas" -a 45.00
./expense-tracker add -d "Movie tickets" -a 24.00

# List all expenses
./expense-tracker list

# Get total summary
./expense-tracker summary
# Output: Summary of Expenses: 154.50

# Get summary for June (month 6)
./expense-tracker summary --month 6
# Output: Summary of Expenses for month 6: 154.50

# Update an expense
./expense-tracker update --id 1 --description "Weekly groceries" --amount 90.00

# Delete an expense
./expense-tracker delete --id 2
```

## Data Storage

- Expenses are stored in `expenses.json` in the current directory
- Each expense contains:
  - `id`: Unique identifier (auto-generated)
  - `description`: Text description of the expense
  - `amount`: Expense amount (float64)
  - `date`: Timestamp when expense was created

### Sample JSON Structure
```json
[
  {
    "id": 1,
    "description": "gym",
    "amount": 2000,
    "date": "2025-06-29T18:00:41.651503319+05:30"
  },
  {
    "id": 2,
    "description": "bought some cool fidget",
    "amount": 250,
    "date": "2025-06-29T21:52:20.363753682+05:30"
  }
]
```

## Error Handling

The application provides clear error messages for common scenarios:

- **Missing required flags**: "Error: Description and amount are required"
- **Invalid ID**: "Error: expense with ID X not found"
- **Invalid month**: "Error: Month must be between 1 and 12"
- **File operations**: Detailed error messages for JSON read/write issues

## Technical Details

- **Language**: Go 1.19+
- **CLI Framework**: Cobra CLI
- **Data Format**: JSON
- **Date Format**: RFC3339 with nanoseconds
- **File Permissions**: 0644 for expense data file

## Project Structure

```
expense-tracker/
├── cmd/
│   ├── root.go          # Root command setup
│   ├── add.go           # Add expense command
│   ├── list.go          # List expenses command
│   ├── delete.go        # Delete expense command
│   ├── update.go        # Update expense command
│   └── summary.go       # Summary command
├── expense/
│   └── expense.go       # Core expense logic
├── main.go              # Application entry point
├── expenses.json        # Data storage (auto-created)
├── go.mod              # Go module file
└── README.md           # This file
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is open source and available under the MIT License.