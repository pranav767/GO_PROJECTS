package main

import 
    (
      "time"
      "encoding/json"
      "fmt"
      "io"
      "net/http"
      "os"
      "github.com/gin-gonic/gin"
      "github.com/gin-contrib/cache"
      "github.com/gin-contrib/cache/persistence"
    )

func main() {
  router := gin.Default()
  router.GET("/ping", func(c *gin.Context) {
    c.JSON(200, gin.H{
      "message": "pong",
    })
  })

  store := persistence.NewRedisCacheWithURL("redis://localhost:6379", time.Minute*5)

  router.GET("/weather", cache.CachePage(store, time.Minute*5, getWeather))

  router.Run() // listen and serve on 0.0.0.0:8080
}

func getWeather(c *gin.Context) {

    // get query parameters
    Location := c.Query("location")
    if Location == "" {
      c.JSON(400, gin.H{"error": "Location query parameter is required"})
      return
    }
  
    // Get API key from env
    apiKey := os.Getenv("WEATHER_API_KEY")

    if apiKey == "" {
      c.JSON(500, gin.H{"error": "Weather API key is not set"})
      return
    }

    // Build the URL
    baseURL := "https://weather.visualcrossing.com/VisualCrossingWebServices/rest/services/timeline/"
    
    fullURL := fmt.Sprintf("%s%s?unitGroup=us&key=%s&contentType=json", baseURL, Location, apiKey)

    resp, err := http.Get(fullURL)
    if err != nil {
      c.JSON(500, gin.H{"error": "Failed to fetch weather data"})
      return
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK{
      c.JSON(resp.StatusCode, gin.H{"error": "Failed to fetch weather data"})
      return
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
      c.JSON(500, gin.H{"error": "Failed to read response body"})
      return
    }

    var apiResponse map[string]interface{}
    if err := json.Unmarshal(body, &apiResponse); err != nil {
      c.JSON(500, gin.H{"error": "Failed to parse weather data"})
      return
    }

    // Extract only the fields you want
    filteredResponse := gin.H{
        "address": apiResponse["address"],
        "alerts": apiResponse["alerts"],
        "currentConditions": apiResponse["currentConditions"],
    }
    c.JSON(200, filteredResponse)
  }


