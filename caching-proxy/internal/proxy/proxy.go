// Set up proxy to forward requests
package proxy

import (
	"caching-proxy/internal/cache"
	"fmt"
	"net/http"
	"io"
	"log"
)

func StartProxy(port, origin string) {
	fmt.Printf("Starting proxy server on :%s, forwarding to %s\n", port, origin)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Forward the request to the origin
		// Get the key from the request , i.e. say /products
		// Get the key from the request path, e.g. /products
		cacheKey := r.URL.Path
		if cachedResp, found := cache.Get(cacheKey); found {
			// If found set cached headers
			for k, v := range cachedResp.Headers {
				for _, vv := range v {
					w.Header().Add(k, vv)
				}
			}
			w.Header().Set("X-Cache", "HIT")
			w.WriteHeader(cachedResp.Status)
			w.Write(cachedResp.Body)
			return
		}

		resp, err := http.Get(origin + cacheKey)
		if err != nil {
			http.Error(w, "Failed to retrieve from origin", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
        if err != nil {
            http.Error(w, "Failed to read origin response", http.StatusInternalServerError)
            return
        }

        // Cache the response (body, headers, status)
        cachedResp := cache.CachedResponse{
            Body:    body,
            Headers: resp.Header.Clone(),
            Status:  resp.StatusCode,
        }
        cache.Set(cacheKey, cachedResp)

        // Set headers from origin
        for k, v := range resp.Header {
            for _, vv := range v {
                w.Header().Add(k, vv)
            }
        }
        w.Header().Set("X-Cache", "MISS")
        w.WriteHeader(resp.StatusCode)
        w.Write(body)
    })
	// Handler for clear Cache
	http.HandleFunc("/clear-cache", func(w http.ResponseWriter, r *http.Request) {
		// Clear the cache here
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		cache.Clear()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Cache cleared successfully"))
	})

    if err := http.ListenAndServe(":"+port, nil); err != nil {
        log.Fatal("Failed to start proxy server: ", err)
    }
}
