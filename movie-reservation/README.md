# Movie Reservation System

A robust backend for movie reservations built with Go (Gin) and MySQL. Supports user and admin flows, dynamic seat logic, and real-time availability. All API examples use dynamic ID extraction for reproducibility.

## Features
- User registration, login, JWT authentication
- Browse movies, genres, showtimes
- Real-time seat selection and reservation
- Admin management for movies, genres, showtimes, theaters
- Reservation monitoring, analytics, and reporting
- Automatic cleanup and baseline data generation

## Prerequisites
- Go 1.19+
- MySQL 8.0+
- jq (for JSON parsing in shell examples)

## Environment Setup
1. Copy `.env.example` to `.env` and set your DB credentials.
2. For development, set `DB_RESET_ON_START=true` to auto-clean and reseed data.
3. Start the server: `go run cmd/main.go`

## API Usage: Step-by-Step

### Admin Flow
#### 1. Login
```bash
ADMIN_TOKEN=$(curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | jq -r '.token')
echo $ADMIN_TOKEN
```
#### 2. Create Genre
```bash
curl -X POST http://localhost:8080/admin/genres \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Mystery","description":"Mystery & suspense"}'
GENRE_ID=$(curl -s http://localhost:8080/genres | jq -r '.genres[0].id')
echo "Using GENRE_ID=$GENRE_ID"
```
#### 3. Create Movie
```bash
curl -X POST http://localhost:8080/admin/movies \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title":"Tomorrowland",
    "description":"A futuristic adventure",
    "genre_id":'$GENRE_ID',
    "duration_minutes":120,
    "release_date":"2025-10-01T00:00:00Z",
    "rating":"PG",
    "language":"English",
    "director":"Jane Director",
    "cast_members":"[\"Actor One\",\"Actor Two\"]"
  }'
MOVIE_ID=$(curl -s http://localhost:8080/movies | jq -r '.movies[] | select(.title=="Tomorrowland") | .id' | head -n1)
echo "Using MOVIE_ID=$MOVIE_ID"
```
#### 4. Create Showtime
```bash
THEATER_ID=$(curl -s http://localhost:8080/theaters | jq -r '.theaters[0].id')
TOMORROW=$(date -d "+1 day" +%F)
echo "Using THEATER_ID=$THEATER_ID MOVIE_ID=$MOVIE_ID TOMORROW=$TOMORROW"
# ⚠️ Always use '18:30' (HH:MM) for show_time. Do NOT use '18:30:00:00'.
curl -X POST http://localhost:8080/admin/showtimes \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"movie_id":'$MOVIE_ID',"theater_id":'$THEATER_ID',"show_date":"'$TOMORROW'","show_time":"18:30","price":15.00}'
SHOWTIME_ID=$(curl -s http://localhost:8080/movies/$MOVIE_ID/showtimes | jq -r '.showtimes[] | select(.show_date | startswith("'$TOMORROW'")) | .id' | head -n1)
echo "Using SHOWTIME_ID=$SHOWTIME_ID"
```

### User Flow
#### 1. Register
```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"username":"john_doe","password":"Password123!","email":"john@example.com","full_name":"John Doe"}'
```
#### 2. Login
```bash
USER_TOKEN=$(curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"john_doe","password":"Password123!"}' | jq -r '.token')
echo $USER_TOKEN
```
#### 3. View Showtimes
```bash
curl http://localhost:8080/movies/$MOVIE_ID/showtimes | jq '.showtimes'
```
#### 4. Check Seat Availability
```bash
curl http://localhost:8080/showtimes/$SHOWTIME_ID/seats | jq '.seat_availability[0]'
SEAT_IDS_JSON=$(curl -s http://localhost:8080/showtimes/$SHOWTIME_ID/seats | jq '[.seat_availability[] | select(.status=="available") | .seat.id] | .[0:2]' | tr -d '\n ')
echo "Using SEAT_IDS_JSON=$SEAT_IDS_JSON"
# Troubleshooting: If you see errors like "Failed to connect to 0.0.0.68 port 80" or "Cannot iterate over null (null)",
# double-check SHOWTIME_ID and use the correct URL: http://localhost:8080/showtimes/$SHOWTIME_ID/seats
```
#### 5. Make Reservation
```bash
curl -s -X POST http://localhost:8080/api/reservations \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"showtime_id":'$SHOWTIME_ID',"seat_ids":'$SEAT_IDS_JSON'}'
```
#### 6. View & Cancel Reservation
```bash
curl -s -H "Authorization: Bearer $USER_TOKEN" http://localhost:8080/api/reservations | jq '.reservations'
RES_ID=$(curl -s -H "Authorization: Bearer $USER_TOKEN" http://localhost:8080/api/reservations | jq -r '.reservations[0].id')
echo "Using RES_ID=$RES_ID"
curl -s -H "Authorization: Bearer $USER_TOKEN" http://localhost:8080/api/reservations/$RES_ID | jq '.reservation'
curl -s -X DELETE -H "Authorization: Bearer $USER_TOKEN" http://localhost:8080/api/reservations/$RES_ID
```

## Troubleshooting & Common Errors
- **Time format error:** Always use `show_time` as `HH:MM` (e.g., `18:30`).
- **Null/connection error:** Ensure all IDs are set and URLs use `localhost:8080`.
- **Booking/cancellation cutoff:** Adjust via `.env` (`BOOKING_CUTOFF_MINUTES`, `CANCEL_CUTOFF_HOURS`).
- **Seats auto-release:** Locked seats not reserved will auto-release after 5 minutes.

## Business Logic & Configuration
- Max 8 seats per reservation (configurable)
- Booking/cancellation cutoff windows (configurable)
- Dynamic pricing by seat type
- Automatic baseline showtime generation
- Environment variables documented in `.env.example`

## Contributing
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License
MIT License
