package tmdbapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Movie struct {
	Title string `json:"title"`
}

type ApiResponse struct {
	Results []Movie `json:"results"`
}

func fetchAndPrintMovies(url string, label string) {
	apiKey := os.Getenv("TMDB_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set the TMDB_API_KEY environment variable.")
		os.Exit(1)
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error while reading response body:", err)
		return
	}
	// Parse the JSON response
	var apiResp ApiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}
	if len(apiResp.Results) == 0 {
		fmt.Printf("No %s movies found.\n", label)
		return
	}
	fmt.Printf("%s Movies:\n", label)
	for _, movie := range apiResp.Results {
		fmt.Println("- " + movie.Title)
	}
}

func Playing() {
	// Implement the logic to fetch currently playing movies
	fmt.Println("Fetching currently playing movies...")
	url := "https://api.themoviedb.org/3/movie/now_playing?language=en-US&page=1"
	fetchAndPrintMovies(url, "Currently Playing")
}

func Popular() {
	fmt.Println("Fetching popular movies...")
	url := "https://api.themoviedb.org/3/movie/popular?language=en-US&page=1"
	fetchAndPrintMovies(url, "Popular")
}

func TopRated() {
	fmt.Println("Fetching top-rated movies...")
	url := "https://api.themoviedb.org/3/movie/top_rated?language=en-US&page=1"
	fetchAndPrintMovies(url, "Top Rated")
}

func Upcoming() {
	fmt.Println("Fetching upcoming movies...")
	url := "https://api.themoviedb.org/3/movie/upcoming?language=en-US&page=1"
	fetchAndPrintMovies(url, "Upcoming")
}
