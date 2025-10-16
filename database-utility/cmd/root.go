package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dbbackup",
	Short: "A flexible database backup utility",
	Long:  "dbbackup is a CLI tool to backup and restore various databases with scheduling and cloud storage.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
