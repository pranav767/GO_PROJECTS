package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

const dsn = "admin:adminpass@tcp(db:3306)/e_commerce_products?parseTime=true"

func main() {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping db: %v", err)
	}

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	// product endpoints
	r.GET("/products", func(c *gin.Context) {
		rows, err := db.Query("SELECT id, name, description, price, inventory FROM products")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
		var prods []map[string]interface{}
		for rows.Next() {
			var id int
			var name, desc string
			var price float64
			var inventory int
			if err := rows.Scan(&id, &name, &desc, &price, &inventory); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			prods = append(prods, map[string]interface{}{"id": id, "name": name, "description": desc, "price": price, "inventory": inventory})
		}
		c.JSON(http.StatusOK, prods)
	})

	r.GET("/products/:id", func(c *gin.Context) {
		id := c.Param("id")
		var pid int
		var name, desc string
		var price float64
		var inventory int
		err := db.QueryRow("SELECT id, name, description, price, inventory FROM products WHERE id = ?", id).Scan(&pid, &name, &desc, &price, &inventory)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"id": pid, "name": name, "description": desc, "price": price, "inventory": inventory})
	})

	if err := r.Run(":8081"); err != nil {
		log.Fatalf("product service failed to run: %v", err)
	}
}
