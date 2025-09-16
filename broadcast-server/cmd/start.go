package cmd

import (
	"broadcast-server/server"

	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the WebSocket broadcast server",
	Long: `Start the WebSocket broadcast server and begin accepting client connections.
The server will listen for incoming WebSocket connections and broadcast
messages between all connected clients.`,
	Run: func(cmd *cobra.Command, args []string) {
		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetInt("port")
		server.StartServer(host, port)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().StringP("host", "H", "localhost", "Host address to bind to")
	startCmd.Flags().IntP("port", "p", 8080, "Port to listen on")
}
