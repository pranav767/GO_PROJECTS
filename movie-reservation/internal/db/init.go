package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// InitializeDatabase creates tables and inserts initial data
func InitializeDatabase() error {
	log.Println("Initializing database...")

	// Read and execute SQL schema
	sqlContent, err := os.ReadFile("internal/db/db.sql")
	if err != nil {
		return fmt.Errorf("failed to read SQL file: %v", err)
	}

	// Split SQL content by statements (crude but effective)
	statements := strings.Split(string(sqlContent), ";")

	for _, statement := range statements {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}

		_, err := db.Exec(statement)
		if err != nil {
			// Log the error but continue (in case of "already exists" errors)
			log.Printf("SQL execution warning: %v", err)
		}
	}

	log.Println("Database initialization completed!")
	return nil
}

// SeedSampleData inserts sample movies and showtimes for testing
func SeedSampleData() error {
	log.Println("Seeding sample data...")

	// Always attempt to generate seats (idempotent)
	if err := generateSeatsForTheaters(); err != nil {
		log.Printf("Seat generation warning: %v", err)
	}

	// Check if we already have movies
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM movies").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		log.Println("Movies already present. Skipping movie insertion but ensuring baseline showtimes for today and tomorrow...")
		if err := ensureBaselineShowtimes(); err != nil {
			log.Printf("Baseline showtime ensure warning: %v", err)
		}
		if err := ensureSeatReservationsForExistingShowtimes(); err != nil {
			log.Printf("Seat reservation backfill warning: %v", err)
		}
		return nil
	}

	// Insert sample movies
	sampleMovies := []struct {
		title       string
		description string
		genreID     int64
		duration    int
		rating      string
		director    string
		cast        string
	}{
		{
			"The Amazing Adventure",
			"An epic journey through unknown lands filled with mystery and excitement.",
			1, 150, "PG-13", "John Director",
			`["Actor One", "Actress Two", "Actor Three"]`,
		},
		{
			"Comedy Night",
			"A hilarious comedy that will keep you laughing all night long.",
			2, 95, "PG", "Jane Comedy",
			`["Comic Actor", "Funny Actress", "Comedian Three"]`,
		},
		{
			"Deep Drama",
			"A thought-provoking drama about life, love, and redemption.",
			3, 125, "R", "Drama Master",
			`["Serious Actor", "Drama Queen", "Method Actor"]`,
		},
		{
			"Horror Nights",
			"A spine-chilling horror movie that will haunt your dreams.",
			4, 105, "R", "Scare Director",
			`["Scream Queen", "Horror Actor", "Final Girl"]`,
		},
		{
			"Love Story",
			"A beautiful romantic tale of two hearts finding each other.",
			5, 110, "PG-13", "Romance Director",
			`["Leading Man", "Leading Lady", "Best Friend"]`,
		},
	}

	for _, movie := range sampleMovies {
		_, err := db.Exec(`INSERT INTO movies (title, description, genre_id, duration_minutes, 
			rating, director, cast_members, poster_image, language, release_date) 
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURDATE())`,
			movie.title, movie.description, movie.genreID, movie.duration,
			movie.rating, movie.director, movie.cast, "/images/default-poster.jpg", "English")
		if err != nil {
			log.Printf("Failed to insert movie %s: %v", movie.title, err)
		}
	}

	// Insert sample showtimes for today and tomorrow (initial seed path)
	sampleShowtimes := []struct {
		movieID   int64
		theaterID int64
		date      string
		time      string
		price     float64
	}{
		{1, 1, "CURDATE()", "14:30", 12.99},
		{1, 1, "CURDATE()", "18:00", 12.99},
		{1, 2, "CURDATE()", "20:30", 15.99},
		{2, 2, "CURDATE()", "15:00", 10.99},
		{2, 3, "CURDATE()", "19:30", 18.99},
		{3, 1, "DATE_ADD(CURDATE(), INTERVAL 1 DAY)", "16:00", 13.99},
		{3, 2, "DATE_ADD(CURDATE(), INTERVAL 1 DAY)", "21:00", 16.99},
		{4, 3, "DATE_ADD(CURDATE(), INTERVAL 1 DAY)", "22:30", 19.99},
		{5, 1, "DATE_ADD(CURDATE(), INTERVAL 1 DAY)", "17:30", 14.99},
	}

	for _, showtime := range sampleShowtimes {
		// First get theater total seats
		var totalSeats int
		err := db.QueryRow("SELECT total_seats FROM theaters WHERE id = ?", showtime.theaterID).Scan(&totalSeats)
		if err != nil {
			continue
		}

		// Insert showtime
		result, err := db.Exec(`INSERT INTO showtimes (movie_id, theater_id, show_date, show_time, price, available_seats) 
			VALUES (?, ?, `+showtime.date+`, ?, ?, ?)`,
			showtime.movieID, showtime.theaterID, showtime.time, showtime.price, totalSeats)
		if err != nil {
			log.Printf("Failed to insert showtime: %v", err)
			continue
		}

		// Get the inserted showtime ID and initialize seat reservations
		showtimeID, err := result.LastInsertId()
		if err != nil {
			continue
		}

		// Initialize seat reservations
		_, err = db.Exec(`INSERT INTO seat_reservations (showtime_id, seat_id, status)
			SELECT ?, id, 'available' FROM seats WHERE theater_id = ? AND is_active = TRUE`,
			showtimeID, showtime.theaterID)
		if err != nil {
			log.Printf("Failed to initialize seat reservations for showtime %d: %v", showtimeID, err)
		}
	}

	// Backfill seat reservations for any showtimes (just created)
	if err := ensureSeatReservationsForExistingShowtimes(); err != nil {
		log.Printf("Seat reservation backfill warning: %v", err)
	}

	// Also ensure baseline for tomorrow (already covered but reuse logic for consistency)
	if err := ensureBaselineShowtimes(); err != nil {
		log.Printf("Baseline showtime ensure warning: %v", err)
	}

	log.Println("Sample data seeded successfully!")
	return nil
}

// generateSeatsForTheaters creates seat rows for each theater based on rows_count and seats_per_row
// It skips creation if seats already exist for a theater.
func generateSeatsForTheaters() error {
	rows, err := db.Query(`SELECT id, rows_count, seats_per_row, name FROM theaters`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type theaterInfo struct {
		id          int64
		rowsCount   int
		seatsPerRow int
		name        string
	}

	var theaters []theaterInfo
	for rows.Next() {
		var t theaterInfo
		if err := rows.Scan(&t.id, &t.rowsCount, &t.seatsPerRow, &t.name); err != nil {
			return err
		}
		theaters = append(theaters, t)
	}

	for _, t := range theaters {
		var existing int
		if err := db.QueryRow(`SELECT COUNT(*) FROM seats WHERE theater_id = ?`, t.id).Scan(&existing); err != nil {
			return err
		}
		if existing > 0 {
			continue // seats already present
		}

		// Insert seats row by row
		for r := 1; r <= t.rowsCount; r++ {
			rowLabel := string(rune('A' + r - 1))
			for s := 1; s <= t.seatsPerRow; s++ {
				seatType := "regular"
				// Apply premium/vip logic similar to original SQL
				if strings.Contains(t.name, "Premium") && r <= 3 { // first 3 rows premium
					seatType = "premium"
				} else if strings.Contains(t.name, "IMAX") && r <= 2 { // first 2 rows vip
					seatType = "vip"
				}
				_, err := db.Exec(`INSERT INTO seats (theater_id, row_label, seat_number, seat_type) VALUES (?,?,?,?)`,
					t.id, rowLabel, s, seatType)
				if err != nil {
					log.Printf("Seat insert warning theater %d row %s seat %d: %v", t.id, rowLabel, s, err)
				}
			}
		}
	}
	return nil
}

// ensureSeatReservationsForExistingShowtimes guarantees that every showtime has a seat_reservations row
// for every active seat in its theater.
func ensureSeatReservationsForExistingShowtimes() error {
	rows, err := db.Query(`SELECT id FROM showtimes`)
	if err != nil {
		return err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return err
		}
		ids = append(ids, id)
	}
	for _, id := range ids {
		if err := EnsureSeatReservations(id); err != nil {
			log.Printf("EnsureSeatReservations failed for showtime %d: %v", id, err)
		}
	}
	return nil
}

// ensureBaselineShowtimes guarantees at least a minimal set of showtimes exist for today and tomorrow
// for the first few movies and theaters. It is idempotent and safe to run on every startup.
func ensureBaselineShowtimes() error {
	// Pick up to first 3 active movies and first 2 active theaters
	movies, err := db.Query(`SELECT id, duration_minutes FROM movies WHERE is_active = TRUE ORDER BY id LIMIT 3`)
	if err != nil {
		return err
	}
	defer movies.Close()
	type mv struct {
		id       int64
		duration int
	}
	var mvList []mv
	for movies.Next() {
		var m mv
		if err := movies.Scan(&m.id, &m.duration); err != nil {
			return err
		}
		mvList = append(mvList, m)
	}
	if len(mvList) == 0 {
		return nil
	}

	theaters, err := db.Query(`SELECT id,total_seats FROM theaters WHERE is_active = TRUE ORDER BY id LIMIT 2`)
	if err != nil {
		return err
	}
	defer theaters.Close()
	type th struct {
		id    int64
		seats int
	}
	var thList []th
	for theaters.Next() {
		var t th
		if err := theaters.Scan(&t.id, &t.seats); err != nil {
			return err
		}
		thList = append(thList, t)
	}
	if len(thList) == 0 {
		return nil
	}

	// Desired times baseline (could be env-enabled later if needed)
	times := []string{"14:30", "18:00", "21:00"}

	// Configuration via environment:
	// BASELINE_SHOWTIME_START_OFFSET_DAYS: how many days from today to start (default 1 = only future)
	// BASELINE_SHOWTIME_DAYS: how many consecutive days to generate starting from offset (default 2)
	offsetDays := getEnvInt("BASELINE_SHOWTIME_START_OFFSET_DAYS", 1)
	totalDays := getEnvInt("BASELINE_SHOWTIME_DAYS", 2)
	if offsetDays < 0 {
		offsetDays = 0
	}
	if totalDays < 1 {
		totalDays = 1
	}

	start := time.Now().Truncate(24 * time.Hour).Add(time.Duration(offsetDays) * 24 * time.Hour)
	var days []time.Time
	for i := 0; i < totalDays; i++ {
		days = append(days, start.Add(time.Duration(i)*24*time.Hour))
	}

	for _, d := range days {
		dateStr := d.Format("2006-01-02")
		for _, m := range mvList {
			for _, t := range thList {
				for _, showTime := range times {
					// Does a showtime already exist?
					var existing int
					err := db.QueryRow(`SELECT COUNT(*) FROM showtimes WHERE movie_id=? AND theater_id=? AND show_date=? AND show_time=?`, m.id, t.id, dateStr, showTime).Scan(&existing)
					if err != nil {
						return err
					}
					if existing > 0 {
						continue
					}
					// Insert
					result, err := db.Exec(`INSERT INTO showtimes (movie_id, theater_id, show_date, show_time, price, available_seats) VALUES (?,?,?,?,?,?)`, m.id, t.id, dateStr, showTime, 12.99, t.seats)
					if err != nil {
						// If duplicate due to race, skip
						if err == sql.ErrNoRows {
							continue
						}
						log.Printf("Baseline showtime insert warning m:%d th:%d %s %s: %v", m.id, t.id, dateStr, showTime, err)
						continue
					}
					showtimeID, err := result.LastInsertId()
					if err == nil {
						// initialize seats
						if _, err := db.Exec(`INSERT INTO seat_reservations (showtime_id, seat_id, status)
							SELECT ?, id, 'available' FROM seats WHERE theater_id=? AND is_active=TRUE`, showtimeID, t.id); err != nil {
							log.Printf("Baseline seat_res init warning showtime %d: %v", showtimeID, err)
						}
					}
				}
			}
		}
	}
	return nil
}

// getEnvInt lightweight helper (local to init.go to avoid extra imports elsewhere)
func getEnvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}
