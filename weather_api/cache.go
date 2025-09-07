package main

import (
  "fmt"
  "time"

  "github.com/gin-contrib/cache"
  "github.com/gin-contrib/cache/persistence"
  "github.com/gin-gonic/gin"
)

func StartCacheServer() {
  r := gin.Default()

  // Basic usage:
  store := persistence.NewRedisCacheWithURL("redis://localhost:6379", time.Minute)

  // Advanced configuration with password and DB number:
  // store := persistence.NewRedisCacheWithURL("redis://:password@localhost:6379/0", time.Minute)

  r.GET("/ping", func(c *gin.Context) {
    c.String(200, "pong "+fmt.Sprint(time.Now().Unix()))
  })
  // Cached Page
  r.GET("/cache_ping", cache.CachePage(store, time.Minute, func(c *gin.Context) {
    c.String(200, "pong "+fmt.Sprint(time.Now().Unix()))
  }))

  // Listen and serve on 0.0.0.0:8080
  r.Run(":8080")
}