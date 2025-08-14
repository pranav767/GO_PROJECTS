// Setting up cache to get, set or clear memory
package cache

import (
	"net/http"
	"sync"
)

// CachedResponse holds both the body and headers
type CachedResponse struct {
	Body    []byte
	Headers http.Header
	Status  int
}

// Setup cache variable which will be a map with string keys and CachedResponse values
var (
    cache      = make(map[string]CachedResponse)
    cacheMutex sync.RWMutex
)

// Get Function will retrieve a value from the cache
func Get(key string) (CachedResponse, bool) {
	// RLock locks rw for reading.
	cacheMutex.RLock()
	// Unlock the mutex when the function returns
	defer cacheMutex.RUnlock()
	// Look for the value in cache, return value and boolean (if found)
	value, ok := cache[key]
	return value, ok
}

// Set function will save a value to the cache
func Set(key string, resp CachedResponse) {
	// Lock the mutex for writing
	cacheMutex.Lock()
	// Unlock the mutex when the function returns
	defer cacheMutex.Unlock()
	// Set the Value in cache
	cache[key] = resp
}

// Clear function will flush the cache
func Clear() {
	// Lock the mutex for writing
	cacheMutex.Lock()
	// Unlock the mutex when the function returns
	defer cacheMutex.Unlock()
	// Clear the cache
	cache = make(map[string]CachedResponse)
}