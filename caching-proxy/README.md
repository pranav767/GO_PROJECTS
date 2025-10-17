# Caching Proxy

A simple CLI tool written in Go that acts as a caching proxy server. It forwards requests to an origin server, caches responses, and allows cache clearing at runtime.

> Project idea from: https://roadmap.sh/projects/caching-server

## Features
- Start a proxy server on a specified port
- Forward requests to an origin server
- Cache responses in memory
- Return cached responses for repeated requests
- Add `X-Cache: HIT` or `X-Cache: MISS` headers to indicate cache status
- Clear cache at runtime via CLI or HTTP endpoint

## Project Structure

```
caching-proxy/
├── main.go                # Entry point
├── cmd/
│   └── root.go            # CLI command setup (flags, execution)
├── internal/
│   ├── cache/
│   │   └── cache.go       # In-memory cache logic
│   ├── proxy/
│   │   └── server.go      # Proxy server logic (handles requests, uses cache)
│   └── server/
│       └── server.go      # Admin logic (clear cache request)
```

## How It Works

1. **Start the Proxy Server**
   - Run:
     ```sh
     go run main.go --port 3000 --origin http://dummyjson.com
     ```
   - The server listens on port 3000 and forwards requests to `http://dummyjson.com`.
   - Responses are cached in memory. Repeat requests return cached data with `X-Cache: HIT`.

2. **Clear the Cache While Running**
   - In a separate terminal, run:
     ```sh
     go run main.go --port 3000 --clear-cache
     ```
   - This sends a POST request to the running server's `/clear-cache` endpoint, clearing the cache.

3. **Check Cache Status**
   - Use `curl` or a browser to make requests to the proxy:
     ```sh
     curl -i http://localhost:3000/products
     ```
   - The response will include `X-Cache: HIT` or `X-Cache: MISS`.

## Extending the Project
- Add support for other HTTP methods (POST, PUT, etc.)
- Implement cache expiry or persistent storage
- Add authentication for admin endpoints
- Use Cobra subcommands for a more robust CLI (`serve`, `clear-cache`)