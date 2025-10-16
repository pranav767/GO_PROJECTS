package db

import (
	"movie-reservation/model"
	"strings"
	"time"
)

// Theater operations
func GetAllTheaters() ([]model.Theater, error) {
	query := `SELECT id, name, location, total_seats, rows_count, seats_per_row, 
			  is_active, created_at FROM theaters WHERE is_active = TRUE ORDER BY name`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var theaters []model.Theater
	for rows.Next() {
		var theater model.Theater
		err := rows.Scan(&theater.ID, &theater.Name, &theater.Location,
			&theater.TotalSeats, &theater.RowsCount, &theater.SeatsPerRow,
			&theater.IsActive, &theater.CreatedAt)
		if err != nil {
			return nil, err
		}
		theaters = append(theaters, theater)
	}
	return theaters, rows.Err()
}

func GetTheaterByID(id int64) (*model.Theater, error) {
	var theater model.Theater
	query := `SELECT id, name, location, total_seats, rows_count, seats_per_row, 
			  is_active, created_at FROM theaters WHERE id = ?`

	err := db.QueryRow(query, id).Scan(&theater.ID, &theater.Name, &theater.Location,
		&theater.TotalSeats, &theater.RowsCount, &theater.SeatsPerRow,
		&theater.IsActive, &theater.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &theater, nil
}

// Seat operations
func GetSeatsByTheater(theaterID int64) ([]model.Seat, error) {
	query := `SELECT id, theater_id, row_label, seat_number, seat_type, is_active, created_at 
			  FROM seats WHERE theater_id = ? AND is_active = TRUE 
			  ORDER BY row_label, seat_number`

	rows, err := db.Query(query, theaterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var seats []model.Seat
	for rows.Next() {
		var seat model.Seat
		err := rows.Scan(&seat.ID, &seat.TheaterID, &seat.RowLabel,
			&seat.SeatNumber, &seat.SeatType, &seat.IsActive, &seat.CreatedAt)
		if err != nil {
			return nil, err
		}
		seats = append(seats, seat)
	}
	return seats, rows.Err()
}

func GetSeatByID(id int64) (*model.Seat, error) {
	var seat model.Seat
	query := `SELECT id, theater_id, row_label, seat_number, seat_type, is_active, created_at 
			  FROM seats WHERE id = ?`

	err := db.QueryRow(query, id).Scan(&seat.ID, &seat.TheaterID, &seat.RowLabel,
		&seat.SeatNumber, &seat.SeatType, &seat.IsActive, &seat.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &seat, nil
}

func GetSeatsByIDs(seatIDs []int64) ([]model.Seat, error) {
	if len(seatIDs) == 0 {
		return []model.Seat{}, nil
	}

	// Create placeholders for IN clause
	placeholders := make([]string, len(seatIDs))
	args := make([]interface{}, len(seatIDs))
	for i, id := range seatIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := `SELECT id, theater_id, row_label, seat_number, seat_type, is_active, created_at 
			  FROM seats WHERE id IN (` + strings.Join(placeholders, ",") + `) AND is_active = TRUE
			  ORDER BY row_label, seat_number`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var seats []model.Seat
	for rows.Next() {
		var seat model.Seat
		err := rows.Scan(&seat.ID, &seat.TheaterID, &seat.RowLabel,
			&seat.SeatNumber, &seat.SeatType, &seat.IsActive, &seat.CreatedAt)
		if err != nil {
			return nil, err
		}
		seats = append(seats, seat)
	}
	return seats, rows.Err()
}

// Showtime operations
func CreateShowtime(req model.CreateShowtimeRequest) (int64, error) {
	// First, get the theater's total seats to set available_seats
	theater, err := GetTheaterByID(req.TheaterID)
	if err != nil {
		return 0, err
	}

	query := `INSERT INTO showtimes (movie_id, theater_id, show_date, show_time, price, available_seats) 
			  VALUES (?, ?, ?, ?, ?, ?)`

	result, err := db.Exec(query, req.MovieID, req.TheaterID, req.ShowDate, req.ShowTime, req.Price, theater.TotalSeats)
	if err != nil {
		return 0, err
	}

	showtimeID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Initialize seat reservations for this showtime
	err = initializeSeatReservations(showtimeID, req.TheaterID)
	if err != nil {
		// Rollback the showtime creation if seat initialization fails
		db.Exec("DELETE FROM showtimes WHERE id = ?", showtimeID)
		return 0, err
	}

	return showtimeID, nil
}

func initializeSeatReservations(showtimeID, theaterID int64) error {
	query := `INSERT INTO seat_reservations (showtime_id, seat_id, status)
			  SELECT ?, id, 'available' FROM seats WHERE theater_id = ? AND is_active = TRUE`

	_, err := db.Exec(query, showtimeID, theaterID)
	return err
}

func GetShowtimesByMovie(movieID int64) ([]model.Showtime, error) {
	query := `SELECT s.id, s.movie_id, s.theater_id, s.show_date, s.show_time, 
			  s.price, s.available_seats, s.is_active, s.created_at, s.updated_at,
			  m.title as movie_title, t.name as theater_name
			  FROM showtimes s
			  LEFT JOIN movies m ON s.movie_id = m.id
			  LEFT JOIN theaters t ON s.theater_id = t.id
			  WHERE s.movie_id = ? AND s.is_active = TRUE AND s.show_date >= CURDATE()
			  ORDER BY s.show_date, s.show_time`

	rows, err := db.Query(query, movieID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var showtimes []model.Showtime
	for rows.Next() {
		var showtime model.Showtime
		err := rows.Scan(&showtime.ID, &showtime.MovieID, &showtime.TheaterID,
			&showtime.ShowDate, &showtime.ShowTime, &showtime.Price,
			&showtime.AvailableSeats, &showtime.IsActive, &showtime.CreatedAt,
			&showtime.UpdatedAt, &showtime.MovieTitle, &showtime.TheaterName)
		if err != nil {
			return nil, err
		}
		showtimes = append(showtimes, showtime)
	}
	return showtimes, rows.Err()
}

func GetShowtimesByDate(date time.Time) ([]model.Showtime, error) {
	query := `SELECT s.id, s.movie_id, s.theater_id, s.show_date, s.show_time, 
			  s.price, s.available_seats, s.is_active, s.created_at, s.updated_at,
			  m.title as movie_title, t.name as theater_name
			  FROM showtimes s
			  LEFT JOIN movies m ON s.movie_id = m.id
			  LEFT JOIN theaters t ON s.theater_id = t.id
			  WHERE s.show_date = ? AND s.is_active = TRUE AND m.is_active = TRUE
			  ORDER BY s.show_time, m.title`

	rows, err := db.Query(query, date.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var showtimes []model.Showtime
	for rows.Next() {
		var showtime model.Showtime
		err := rows.Scan(&showtime.ID, &showtime.MovieID, &showtime.TheaterID,
			&showtime.ShowDate, &showtime.ShowTime, &showtime.Price,
			&showtime.AvailableSeats, &showtime.IsActive, &showtime.CreatedAt,
			&showtime.UpdatedAt, &showtime.MovieTitle, &showtime.TheaterName)
		if err != nil {
			return nil, err
		}
		showtimes = append(showtimes, showtime)
	}
	return showtimes, rows.Err()
}

func GetShowtimeByID(id int64) (*model.Showtime, error) {
	var showtime model.Showtime
	query := `SELECT s.id, s.movie_id, s.theater_id, s.show_date, s.show_time, 
			  s.price, s.available_seats, s.is_active, s.created_at, s.updated_at,
			  m.title as movie_title, t.name as theater_name
			  FROM showtimes s
			  LEFT JOIN movies m ON s.movie_id = m.id
			  LEFT JOIN theaters t ON s.theater_id = t.id
			  WHERE s.id = ?`

	err := db.QueryRow(query, id).Scan(&showtime.ID, &showtime.MovieID, &showtime.TheaterID,
		&showtime.ShowDate, &showtime.ShowTime, &showtime.Price,
		&showtime.AvailableSeats, &showtime.IsActive, &showtime.CreatedAt,
		&showtime.UpdatedAt, &showtime.MovieTitle, &showtime.TheaterName)
	if err != nil {
		return nil, err
	}
	return &showtime, nil
}

func UpdateShowtimeAvailableSeats(showtimeID int64, change int) error {
	_, err := db.Exec("UPDATE showtimes SET available_seats = available_seats + ? WHERE id = ?", change, showtimeID)
	return err
}

func DeactivateShowtime(id int64) error {
	_, err := db.Exec("UPDATE showtimes SET is_active = FALSE WHERE id = ?", id)
	return err
}
