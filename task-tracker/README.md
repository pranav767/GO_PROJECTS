# Task Tracker CLI

Task Tracker CLI is a command-line tool to manage your tasks efficiently. It allows you to add, list, update, and delete tasks, as well as mark them as done or in-progress. This project is built in Go using the [Cobra](https://github.com/spf13/cobra) library for command-line interfaces.

## Features

- Add new tasks with descriptions
- List all tasks or filter by status (`done`, `todo`, `in-progress`)
- Update task descriptions
- Mark tasks as done or in-progress
- Delete tasks by ID
- Shell autocompletion support

## How It Was Built

- **Language:** Go
- **CLI Framework:** [Cobra](https://github.com/spf13/cobra)
**Task Storage:** Tasks are stored in a simple `tasks.json` file in the project directory. No external database is used; all data is managed via Go structs and file I/O (see the `tasks` package).
- **Structure:** Each command (add, list, delete, etc.) is implemented as a separate file in the `cmd` directory and registered with the root command.

## Installation

1. **Clone the repository:**
   ```sh
   git clone https://github.com/pranav767/GO_PROJECTS.git
   cd GO_PROJECTS/task-tracker
   ```

2. **Build the CLI:**
   ```sh
   go build -o task-tracker
   ```

## Usage

### Add a new task
```sh
./task-tracker add "Buy groceries"
```

### List all tasks
```sh
./task-tracker list
```

### List tasks by status
```sh
./task-tracker list done
./task-tracker list todo
./task-tracker list in-progress
```

### Mark a task as done or in-progress
```sh
./task-tracker mark-done 2
./task-tracker mark-in-progress 3
```

### Update a task's description
```sh
./task-tracker update 1 "Buy groceries and cook dinner"
```

### Delete a task
```sh
./task-tracker delete 1
```

### Enable Shell Autocompletion

Generate a completion script for your shell:

- **Bash:**
  ```sh
  ./task-tracker completion bash > /etc/bash_completion.d/task-tracker
  ```
- **Zsh:**
  ```sh
  ./task-tracker completion zsh > "${fpath[1]}/_task-tracker"
  ```

## Project Structure

```
cmd/         # Cobra commands (add, list, delete, etc.)
tasks/       # Task management logic (structs, file I/O)
main.go      # Entry point
```

## Storage Details

This project does **not** use a database. All tasks are saved in a local `tasks.json` file, which is automatically created and updated as you use the CLI. This makes it easy to use and portable—just keep your `tasks.json` file to retain your tasks.

If you delete or move the `tasks.json` file, your tasks will be lost or moved accordingly. No external dependencies or database setup is required.


Built with Go and [Cobra](https://github.com/spf13/cobra)
