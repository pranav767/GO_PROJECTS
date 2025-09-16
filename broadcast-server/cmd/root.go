// Package cmd provides the command-line interface for the broadcast server application.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "broadcast-server",
	Short: "A WebSocket broadcast server and client",
	Long: `A real-time chat application using WebSocket technology.
This application can be run either as a server to accept connections
or as a client to connect to an existing server.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
