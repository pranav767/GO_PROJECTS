/*
*/
package cmd

import (
	"caching-proxy/internal/proxy"
	"caching-proxy/internal/server"
	"os"

	"github.com/spf13/cobra"
)

var (
	port   string
	origin string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "caching-proxy",
	Short: "Caching proxy",
	Long: `Caching proxy is a simple HTTP proxy server that forwards requests to a specified origin.
	example usage:
	./caching-proxy --port 3000 --origin http://dummyjson.com
	./caching-proxy --clear-cache`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		// Call the proxy (server.go) if --port & -- origin set
		if port != "" && origin != "" {
			proxy.StartProxy(port, origin)
		}
		if cmd.Flags().Changed("clear-cache") {
			if port == "" {
				fmt.Println("Error: --port is required when using --clear-cache")
				os.Exit(1)
			}
			server.RequestClearCache(port)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.caching-proxy.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().StringVarP(&port, "port", "p", "", "Port to run the server on")
	rootCmd.Flags().StringVarP(&origin, "origin", "o", "", "Origin of the request")
	rootCmd.Flags().Bool("clear-cache", false, "Clear the cache")
}
