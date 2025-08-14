// Makes a http request to clear cache, after proxy server has started
package server

import (
	"fmt"
	"net/http"
)

func RequestClearCache(port string) {
	resp, err := http.Post("http://localhost:"+port+"/clear-cache", "", nil)
	if err != nil {
		fmt.Println("Failed to clear cache:", err)
		return
	}
	defer resp.Body.Close()
	fmt.Println("Cache cleared (server responded with status:", resp.Status, ")")
}
