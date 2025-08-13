	/*
	Copyright © 2025 NAME HERE <EMAIL ADDRESS>
	*/
	package cmd

	import (
		"fmt"
		"os"
		"tmdb/tmdbapi"

		"github.com/spf13/cobra"
	)

	var movieType string

	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:   "tmdb [--type <movie_type>]",
		Short: "A CLI tool for The Movie Database (TMDB)",
		Long: `A command line interface for fetching movie data from TMDB.
	Supports different movie categories like playing, popular, top rated, and upcoming.
	Example usage:
	tmdb --type playing
	tmdb --type popular
	tmdb --type top_rated
	tmdb --type upcoming`,

		Run: func(cmd *cobra.Command, args []string) {
			switch movieType {
			case "playing":
				tmdbapi.Playing() // Call the Playing function from tmdbapi package
				// Fetch and display currently playing movies
				fmt.Println("Fetching currently playing movies...")
			case "popular":
				// Fetch and display popular movies
				tmdbapi.Popular() // Assuming you have a Popular function in tmdbapi
			case "top_rated":
				// Fetch and display top-rated movies
				tmdbapi.TopRated() // Assuming you have a TopRated function in tmdbapi
			case "upcoming":
				// Fetch and display upcoming movies
				tmdbapi.Upcoming() // Assuming you have an Upcoming function in tmdbapi
			default:
				// Handle unknown movie types
				fmt.Println("Unknown movie type. Please use one of: playing, popular, top_rated, upcoming.")
				os.Exit(1)
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

		// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.tmdb.yaml)")

		// Cobra also supports local flags, which will only run
		// when this action is called directly.
		rootCmd.Flags().StringVarP(&movieType, "type", "t", "", "Type of movies to fetch (playing, popular, top_rated, upcoming)")
		rootCmd.MarkFlagRequired("type")
		//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	}
