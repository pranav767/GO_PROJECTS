package main

import (
	//"net/http"
	//"os"
	//"github.com/gin-gonic/gin"
	"md-notes/routes"
)

/* Endpoints

1st important will be upload the note
can do following with ID
check the grammar,
view note in md
render it in HTML.

*/
func main() {
	router := routes.SetupRouter()
	router.Run(":8080")
}