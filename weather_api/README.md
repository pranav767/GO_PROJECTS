# Weather API with Redis Caching

A Go web API that fetches weather data from Visual Crossing Weather API with Redis caching for improved performance.

## Prerequisites

- Go 1.19+
- Docker
- Weather API key from [Visual Crossing](https://www.visualcrossing.com/weather-api)

## Setup

### 1. Start Redis with Docker

First, start a Redis container:

```bash
docker run -d --name redis-cache -p 6379:6379 redis:latest
```

Verify Redis is running:
```bash
docker ps
```

### 2. Set Environment Variable

Set your Weather API key:
```bash
export WEATHER_API_KEY=your_actual_api_key_here
```

### 3. Install Dependencies

```bash
go mod tidy
```

### 4. Run the Application

```bash
go run main.go
```

The server will start on `http://localhost:8080`

## API Endpoints

### Test Connection
```
GET http://localhost:8080/ping
```

### Get Weather Data
```
GET http://localhost:8080/weather?location=ireland
GET http://localhost:8080/weather?location=london
GET http://localhost:8080/weather?location=new%20york
```

## Caching

The API uses Redis caching with a **5-minute** expiration time. This means:
- First request for a location fetches from the weather API
- Subsequent requests for the same location return cached data
- Cache expires after 5 minutes

## Testing Cache

### Method 1: Check Response Time
```bash
# First request (slower - hits external API)
time curl "http://localhost:8080/weather?location=ireland"

# Second request (faster - hits cache)
time curl "http://localhost:8080/weather?location=ireland"
```

### Method 2: Monitor Redis Activity

Connect to Redis container and monitor commands:
```bash
# Connect to Redis CLI
docker exec -it redis-cache redis-cli

# Monitor all Redis commands in real-time
127.0.0.1:6379> MONITOR
```

In another terminal, make API requests and watch Redis activity.

### Method 3: Check Cache Keys

```bash
# Connect to Redis CLI
docker exec -it redis-cache redis-cli

# List all cache keys
127.0.0.1:6379> KEYS *

# List only gin cache keys
127.0.0.1:6379> KEYS "gincontrib.page.cache:*"

# Check TTL (time to live) of a specific key
127.0.0.1:6379> TTL "gincontrib.page.cache:your_key_here"

# Get cached content
127.0.0.1:6379> GET "gincontrib.page.cache:your_key_here"
```

## Example Usage

```bash
# Test basic connectivity
curl http://localhost:8080/ping

# Get weather for Ireland (first request - hits API)
curl "http://localhost:8080/weather?location=ireland"

# Get weather for Ireland again (second request - hits cache)
curl "http://localhost:8080/weather?location=ireland"

# Get weather for a different location (new cache entry)
curl "http://localhost:8080/weather?location=tokyo"
```

## Response Format

The API returns filtered weather data:
```json
{
  "address": "ireland",
  "alerts": [],
  "currentConditions": {
    "cloudcover": 45.3,
    "conditions": "Partially cloudy",
    "temp": 66.9,
    "humidity": 69.9,
    "windspeed": 6,
    "pressure": 1021.7,
    "visibility": 6.2
  }
}
```

## Cleanup

Stop and remove Redis container:
```bash
docker stop redis-cache
docker rm redis-cache
```

## Troubleshooting

### Redis Connection Issues
- Ensure Docker is running
- Check if Redis container is running: `docker ps`
- Verify Redis port 6379 is available

### API Key Issues
- Make sure `WEATHER_API_KEY` environment variable is set
- Verify API key is valid at Visual Crossing

### Cache Not Working
- Check Redis logs: `docker logs redis-cache`
- Monitor Redis activity: `docker exec -it redis-cache redis-cli MONITOR`
- Verify cache keys exist: `docker exec -it redis-cache redis-cli KEYS "*"`
