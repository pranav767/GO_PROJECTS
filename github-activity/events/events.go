package events

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Function to fetch events for a given username
func FetchEvents(username string) {

	resp, err := http.Get(fmt.Sprintf("https://api.github.com/users/%s/events", username))
	if err != nil {
		fmt.Println("Error fetching URL:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		os.Exit(1)
	}

	var events []map[string]interface{}
	err = json.Unmarshal(body, &events)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		os.Exit(1)
	}

	// Gives a sample output of the first event
	//b, _ := json.MarshalIndent(events[2], "", "  ")
	//fmt.Println(string(b))

	for _, event := range events {
		eventType, _ := event["type"].(string)
		repo := ""
		if r, ok := event["repo"].(map[string]interface{}); ok {
			repo, _ = r["name"].(string)
		} else if r, ok := event["repo"].(map[string]string); ok {
			repo = r["name"]
		} else if r, ok := event["repo"].(map[string]any); ok {
			if name, ok := r["name"].(string); ok {
				repo = name
			}
		}
		switch eventType {
		case "PushEvent":
			payload, _ := event["payload"].(map[string]interface{})
			commits := 0
			if c, ok := payload["commits"].([]interface{}); ok {
				commits = len(c)
			}
			fmt.Printf("Pushed %d commits to %s\n", commits, repo)
		case "IssuesEvent":
			payload, _ := event["payload"].(map[string]interface{})
			action, _ := payload["action"].(string)
			fmt.Printf("%s an issue in %s\n", capitalize(action), repo)
		case "WatchEvent":
			fmt.Printf("Starred %s\n", repo)
		case "ForkEvent":
			fmt.Printf("Forked %s\n", repo)
		case "PullRequestEvent":
			payload, _ := event["payload"].(map[string]interface{})
			action, _ := payload["action"].(string)
			fmt.Printf("%s a pull request in %s\n", capitalize(action), repo)
		// Add more cases as needed for other event types
		default:
			// Uncomment to see all event types
			// fmt.Printf("%s in %s\n", eventType, repo)
		}
	}
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(s[0]-32) + s[1:]
}
