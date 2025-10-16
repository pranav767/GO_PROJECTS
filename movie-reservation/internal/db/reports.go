package db

import (
	"movie-reservation/model"
	"time"
)

// Admin reporting functions

func GetRevenueReport(startDate, endDate time.Time) (*model.RevenueReport, error) {
	// Get total revenue and reservations
	var totalReservations int
	var totalRevenue float64

	err := db.QueryRow(`SELECT COUNT(*) as total_reservations, COALESCE(SUM(total_amount), 0) as total_revenue
		FROM reservations 
		WHERE status = 'confirmed' AND booking_date >= ? AND booking_date <= ?`,
		startDate, endDate).Scan(&totalReservations, &totalRevenue)
	if err != nil {
		return nil, err
	}

	// Get movie breakdown
	query := `SELECT m.id, m.title, COUNT(r.id) as reservations, COALESCE(SUM(r.total_amount), 0) as revenue
		FROM movies m
		LEFT JOIN showtimes s ON m.id = s.movie_id
		LEFT JOIN reservations r ON s.id = r.showtime_id AND r.status = 'confirmed' 
			AND r.booking_date >= ? AND r.booking_date <= ?
		GROUP BY m.id, m.title
		HAVING reservations > 0
		ORDER BY revenue DESC`

	rows, err := db.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movieBreakdown []model.MovieRevenue
	for rows.Next() {
		var movie model.MovieRevenue
		err := rows.Scan(&movie.MovieID, &movie.MovieTitle, &movie.Reservations, &movie.Revenue)
		if err != nil {
			return nil, err
		}
		movieBreakdown = append(movieBreakdown, movie)
	}

	return &model.RevenueReport{
		TotalReservations: totalReservations,
		TotalRevenue:      totalRevenue,
		Period:            startDate.Format("2006-01-02") + " to " + endDate.Format("2006-01-02"),
		MovieBreakdown:    movieBreakdown,
	}, nil
}

func GetCapacityReport(startDate, endDate time.Time) ([]model.CapacityReport, error) {
	query := `SELECT t.id, t.name, t.total_seats,
		COUNT(CASE WHEN sr.status = 'reserved' AND s.show_date >= ? AND s.show_date <= ? THEN 1 END) as reserved_seats
		FROM theaters t
		LEFT JOIN showtimes s ON t.id = s.theater_id
		LEFT JOIN seat_reservations sr ON s.id = sr.showtime_id
		WHERE t.is_active = TRUE
		GROUP BY t.id, t.name, t.total_seats
		ORDER BY t.name`

	rows, err := db.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []model.CapacityReport
	for rows.Next() {
		var report model.CapacityReport
		err := rows.Scan(&report.TheaterID, &report.TheaterName, &report.TotalSeats, &report.ReservedSeats)
		if err != nil {
			return nil, err
		}

		if report.TotalSeats > 0 {
			report.CapacityPercent = (float64(report.ReservedSeats) / float64(report.TotalSeats)) * 100
		}
		report.Period = startDate.Format("2006-01-02") + " to " + endDate.Format("2006-01-02")

		reports = append(reports, report)
	}
	return reports, nil
}

func GetPopularMovies(startDate, endDate time.Time, limit int) ([]model.MovieRevenue, error) {
	query := `SELECT m.id, m.title, COUNT(r.id) as reservations, COALESCE(SUM(r.total_amount), 0) as revenue
		FROM movies m
		LEFT JOIN showtimes s ON m.id = s.movie_id
		LEFT JOIN reservations r ON s.id = r.showtime_id AND r.status = 'confirmed' 
			AND r.booking_date >= ? AND r.booking_date <= ?
		WHERE m.is_active = TRUE
		GROUP BY m.id, m.title
		ORDER BY reservations DESC, revenue DESC
		LIMIT ?`

	rows, err := db.Query(query, startDate, endDate, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []model.MovieRevenue
	for rows.Next() {
		var movie model.MovieRevenue
		err := rows.Scan(&movie.MovieID, &movie.MovieTitle, &movie.Reservations, &movie.Revenue)
		if err != nil {
			return nil, err
		}
		movies = append(movies, movie)
	}
	return movies, nil
}

func GetDailyReservationStats(date time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total reservations for the day
	var totalReservations int
	var totalRevenue float64
	err := db.QueryRow(`SELECT COUNT(*) as total_reservations, COALESCE(SUM(total_amount), 0) as total_revenue
		FROM reservations 
		WHERE status = 'confirmed' AND DATE(booking_date) = ?`, date.Format("2006-01-02")).
		Scan(&totalReservations, &totalRevenue)
	if err != nil {
		return nil, err
	}

	stats["total_reservations"] = totalReservations
	stats["total_revenue"] = totalRevenue

	// Reservations by theater
	theaterQuery := `SELECT t.name, COUNT(r.id) as reservations
		FROM theaters t
		LEFT JOIN showtimes s ON t.id = s.theater_id
		LEFT JOIN reservations r ON s.id = r.showtime_id AND r.status = 'confirmed' 
			AND DATE(r.booking_date) = ?
		WHERE t.is_active = TRUE
		GROUP BY t.id, t.name
		ORDER BY reservations DESC`

	rows, err := db.Query(theaterQuery, date.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	theaterStats := make([]map[string]interface{}, 0)
	for rows.Next() {
		var theaterName string
		var reservations int
		err := rows.Scan(&theaterName, &reservations)
		if err != nil {
			return nil, err
		}
		theaterStats = append(theaterStats, map[string]interface{}{
			"theater":      theaterName,
			"reservations": reservations,
		})
	}
	stats["by_theater"] = theaterStats

	return stats, nil
}

func GetOccupancyRate(showtimeID int64) (float64, error) {
	var totalSeats, reservedSeats int

	query := `SELECT 
		(SELECT COUNT(*) FROM seats s JOIN showtimes st ON s.theater_id = st.theater_id WHERE st.id = ? AND s.is_active = TRUE) as total_seats,
		COUNT(CASE WHEN sr.status = 'reserved' THEN 1 END) as reserved_seats
		FROM seat_reservations sr
		WHERE sr.showtime_id = ?`

	err := db.QueryRow(query, showtimeID, showtimeID).Scan(&totalSeats, &reservedSeats)
	if err != nil {
		return 0, err
	}

	if totalSeats == 0 {
		return 0, nil
	}

	return (float64(reservedSeats) / float64(totalSeats)) * 100, nil
}
