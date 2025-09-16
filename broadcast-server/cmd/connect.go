package cmd

import (
    "broadcast-server/client"
    "fmt"

    "github.com/spf13/cobra"
)

var connectCmd = &cobra.Command{
    Use:   "connect",
    Short: "Connect to the broadcast server",
    Long:  `Connect to the WebSocket broadcast server as a client`,
    Run: func(cmd *cobra.Command, args []string) {
        host, _ := cmd.Flags().GetString("host")
        port, _ := cmd.Flags().GetInt("port")
        name, _ := cmd.Flags().GetString("name")
        
        fmt.Printf("Connecting to %s:%d as %s\n", host, port, name)
        client.StartClient(host, port, name)
    },
}

func init() {
    rootCmd.AddCommand(connectCmd)
    
    connectCmd.Flags().StringP("host", "H", "localhost", "Server host address")
    connectCmd.Flags().IntP("port", "p", 8080, "Server port")
    connectCmd.Flags().StringP("name", "n", "anonymous", "Client name")
}