package db

import (
	"database/sql"
	"fmt"
	"movie-reservation/model"
	"strings"
	"time"
)

// Reservation operations with concurrency handling

func GetSeatAvailabilityForShowtime(showtimeID int64) ([]model.SeatAvailabilityResponse, error) {
	query := `SELECT s.id, s.theater_id, s.row_label, s.seat_number, s.seat_type, 
			  s.is_active, s.created_at, sr.status, st.price
			  FROM seats s
			  JOIN seat_reservations sr ON s.id = sr.seat_id
			  JOIN showtimes st ON sr.showtime_id = st.id
			  WHERE sr.showtime_id = ? AND s.is_active = TRUE
			  ORDER BY s.row_label, s.seat_number`

	rows, err := db.Query(query, showtimeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var availability []model.SeatAvailabilityResponse
	for rows.Next() {
		var seat model.Seat
		var status string
		var price float64

		err := rows.Scan(&seat.ID, &seat.TheaterID, &seat.RowLabel,
			&seat.SeatNumber, &seat.SeatType, &seat.IsActive, &seat.CreatedAt,
			&status, &price)
		if err != nil {
			return nil, err
		}

		// Adjust price based on seat type
		adjustedPrice := price
		switch seat.SeatType {
		case "premium":
			adjustedPrice = price * 1.5
		case "vip":
			adjustedPrice = price * 2.0
		}

		availability = append(availability, model.SeatAvailabilityResponse{
			Seat:   seat,
			Status: status,
			Price:  adjustedPrice,
		})
	}
	return availability, rows.Err()
}

// EnsureSeatReservations backfills seat_reservations rows in case seats were created
// after a showtime (or an earlier initialization failed quietly).
func EnsureSeatReservations(showtimeID int64) error {
	// Insert any missing seat_reservation rows for this showtime
	stmt := `INSERT INTO seat_reservations (showtime_id, seat_id, status)
			 SELECT st.id, s.id, 'available'
			 FROM showtimes st
			 JOIN seats s ON s.theater_id = st.theater_id AND s.is_active = TRUE
			 LEFT JOIN seat_reservations sr ON sr.showtime_id = st.id AND sr.seat_id = s.id
			 WHERE st.id = ? AND sr.id IS NULL`
	_, err := db.Exec(stmt, showtimeID)
	return err
}

func LockSeats(showtimeID int64, seatIDs []int64, lockDuration time.Duration) error {
	if len(seatIDs) == 0 {
		return fmt.Errorf("no seats provided")
	}

	// Create placeholders for IN clause
	placeholders := make([]string, len(seatIDs))
	args := make([]interface{}, 0, len(seatIDs)+2)
	args = append(args, showtimeID)
	args = append(args, time.Now().Add(lockDuration))

	for i, id := range seatIDs {
		placeholders[i] = "?"
		args = append(args, id)
	}

	query := `UPDATE seat_reservations 
			  SET status = 'locked', locked_until = ?, updated_at = CURRENT_TIMESTAMP
			  WHERE showtime_id = ? AND seat_id IN (` + strings.Join(placeholders, ",") + `) 
			  AND status = 'available'`

	// Move the args properly
	finalArgs := []interface{}{time.Now().Add(lockDuration), showtimeID}
	for _, id := range seatIDs {
		finalArgs = append(finalArgs, id)
	}

	result, err := db.Exec(query, finalArgs...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if int(rowsAffected) != len(seatIDs) {
		return fmt.Errorf("could not lock all requested seats, only %d out of %d were available", rowsAffected, len(seatIDs))
	}

	return nil
}

func ReleaseSeatLocks(showtimeID int64, seatIDs []int64) error {
	if len(seatIDs) == 0 {
		return nil
	}

	placeholders := make([]string, len(seatIDs))
	args := make([]interface{}, 0, len(seatIDs)+1)
	args = append(args, showtimeID)

	for i, id := range seatIDs {
		placeholders[i] = "?"
		args = append(args, id)
	}

	query := `UPDATE seat_reservations 
			  SET status = 'available', locked_until = NULL, updated_at = CURRENT_TIMESTAMP
			  WHERE showtime_id = ? AND seat_id IN (` + strings.Join(placeholders, ",") + `) 
			  AND status = 'locked'`

	_, err := db.Exec(query, args...)
	return err
}

func CreateReservation(userID, showtimeID int64, seatIDs []int64, totalAmount float64) (*model.Reservation, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Generate reservation code
	reservationCode := fmt.Sprintf("RES%d%d", time.Now().Unix(), userID)

	// Create reservation
	result, err := tx.Exec(`INSERT INTO reservations (user_id, showtime_id, reservation_code, 
		total_seats, total_amount, status, booking_date) VALUES (?, ?, ?, ?, ?, 'confirmed', NOW())`,
		userID, showtimeID, reservationCode, len(seatIDs), totalAmount)
	if err != nil {
		return nil, err
	}

	reservationID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Get seat prices for each seat
	seats, err := GetSeatsByIDs(seatIDs)
	if err != nil {
		return nil, err
	}

	// Get showtime price
	showtime, err := GetShowtimeByID(showtimeID)
	if err != nil {
		return nil, err
	}

	// Create reservation seats and update seat reservations
	for _, seat := range seats {
		// Calculate price based on seat type
		seatPrice := showtime.Price
		switch seat.SeatType {
		case "premium":
			seatPrice = showtime.Price * 1.5
		case "vip":
			seatPrice = showtime.Price * 2.0
		}

		// Insert into reservation_seats
		_, err = tx.Exec(`INSERT INTO reservation_seats (reservation_id, seat_id, price) 
			VALUES (?, ?, ?)`, reservationID, seat.ID, seatPrice)
		if err != nil {
			return nil, err
		}

		// Update seat_reservations to mark as reserved
		_, err = tx.Exec(`UPDATE seat_reservations 
			SET status = 'reserved', reservation_id = ?, locked_until = NULL, updated_at = CURRENT_TIMESTAMP
			WHERE showtime_id = ? AND seat_id = ? AND status = 'locked'`,
			reservationID, showtimeID, seat.ID)
		if err != nil {
			return nil, err
		}
	}

	// Update available seats count
	err = UpdateShowtimeAvailableSeatsInTx(tx, showtimeID, -len(seatIDs))
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	// Return the created reservation
	return GetReservationByID(reservationID)
}

func UpdateShowtimeAvailableSeatsInTx(tx *sql.Tx, showtimeID int64, change int) error {
	_, err := tx.Exec("UPDATE showtimes SET available_seats = available_seats + ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		change, showtimeID)
	return err
}

func GetReservationByID(id int64) (*model.Reservation, error) {
	query := `SELECT r.id, r.user_id, r.showtime_id, r.reservation_code, r.total_seats, 
			  r.total_amount, r.status, r.booking_date, r.created_at, r.updated_at,
			  m.title as movie_title, t.name as theater_name, s.show_date, s.show_time, u.username
			  FROM reservations r
			  LEFT JOIN showtimes s ON r.showtime_id = s.id
			  LEFT JOIN movies m ON s.movie_id = m.id
			  LEFT JOIN theaters t ON s.theater_id = t.id
			  LEFT JOIN users u ON r.user_id = u.id
			  WHERE r.id = ?`

	var reservation model.Reservation
	err := db.QueryRow(query, id).Scan(&reservation.ID, &reservation.UserID, &reservation.ShowtimeID,
		&reservation.ReservationCode, &reservation.TotalSeats, &reservation.TotalAmount,
		&reservation.Status, &reservation.BookingDate, &reservation.CreatedAt, &reservation.UpdatedAt,
		&reservation.MovieTitle, &reservation.TheaterName, &reservation.ShowDate, &reservation.ShowTime,
		&reservation.Username)
	if err != nil {
		return nil, err
	}

	// Get reserved seats
	seats, err := GetReservationSeats(reservation.ID)
	if err != nil {
		return nil, err
	}
	reservation.ReservedSeats = seats

	return &reservation, nil
}

func GetReservationSeats(reservationID int64) ([]model.ReservationSeat, error) {
	query := `SELECT rs.id, rs.reservation_id, rs.seat_id, rs.price, rs.created_at,
			  s.row_label, s.seat_number, s.seat_type
			  FROM reservation_seats rs
			  JOIN seats s ON rs.seat_id = s.id
			  WHERE rs.reservation_id = ?
			  ORDER BY s.row_label, s.seat_number`

	rows, err := db.Query(query, reservationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var seats []model.ReservationSeat
	for rows.Next() {
		var seat model.ReservationSeat
		err := rows.Scan(&seat.ID, &seat.ReservationID, &seat.SeatID, &seat.Price,
			&seat.CreatedAt, &seat.RowLabel, &seat.SeatNumber, &seat.SeatType)
		if err != nil {
			return nil, err
		}
		seats = append(seats, seat)
	}
	return seats, rows.Err()
}

func GetUserReservations(userID int64) ([]model.Reservation, error) {
	query := `SELECT r.id, r.user_id, r.showtime_id, r.reservation_code, r.total_seats, 
			  r.total_amount, r.status, r.booking_date, r.created_at, r.updated_at,
			  m.title as movie_title, t.name as theater_name, s.show_date, s.show_time
			  FROM reservations r
			  LEFT JOIN showtimes s ON r.showtime_id = s.id
			  LEFT JOIN movies m ON s.movie_id = m.id
			  LEFT JOIN theaters t ON s.theater_id = t.id
			  WHERE r.user_id = ?
			  ORDER BY r.created_at DESC`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reservations []model.Reservation
	for rows.Next() {
		var reservation model.Reservation
		err := rows.Scan(&reservation.ID, &reservation.UserID, &reservation.ShowtimeID,
			&reservation.ReservationCode, &reservation.TotalSeats, &reservation.TotalAmount,
			&reservation.Status, &reservation.BookingDate, &reservation.CreatedAt, &reservation.UpdatedAt,
			&reservation.MovieTitle, &reservation.TheaterName, &reservation.ShowDate, &reservation.ShowTime)
		if err != nil {
			return nil, err
		}
		reservations = append(reservations, reservation)
	}
	return reservations, rows.Err()
}

func CancelReservation(reservationID, userID int64) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get reservation details
	var showtimeID int64
	var totalSeats int
	var status string
	err = tx.QueryRow("SELECT showtime_id, total_seats, status FROM reservations WHERE id = ? AND user_id = ?",
		reservationID, userID).Scan(&showtimeID, &totalSeats, &status)
	if err != nil {
		return err
	}

	if status == "cancelled" {
		return fmt.Errorf("reservation is already cancelled")
	}

	// Check if showtime is in the future
	var showDate time.Time
	var showTime string
	err = tx.QueryRow("SELECT show_date, show_time FROM showtimes WHERE id = ?", showtimeID).
		Scan(&showDate, &showTime)
	if err != nil {
		return err
	}

	// Parse show time robustly (accept HH:MM or HH:MM:SS)
	showDateTime, err := time.Parse("2006-01-02 15:04:05", showDate.Format("2006-01-02")+" "+showTime)
	if err != nil {
		// Try fallback to HH:MM
		showDateTime, err = time.Parse("2006-01-02 15:04", showDate.Format("2006-01-02")+" "+showTime)
		if err != nil {
			return fmt.Errorf("invalid show_time format: %v", err)
		}
	}

	if time.Now().After(showDateTime) {
		return fmt.Errorf("cannot cancel past reservations")
	}

	// Update reservation status
	_, err = tx.Exec("UPDATE reservations SET status = 'cancelled', updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		reservationID)
	if err != nil {
		return err
	}

	// Release seats
	_, err = tx.Exec(`UPDATE seat_reservations 
		SET status = 'available', reservation_id = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE reservation_id = ?`, reservationID)
	if err != nil {
		return err
	}

	// Update available seats count
	err = UpdateShowtimeAvailableSeatsInTx(tx, showtimeID, totalSeats)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func GetAllReservations() ([]model.Reservation, error) {
	query := `SELECT r.id, r.user_id, r.showtime_id, r.reservation_code, r.total_seats, 
			  r.total_amount, r.status, r.booking_date, r.created_at, r.updated_at,
			  m.title as movie_title, t.name as theater_name, s.show_date, s.show_time, u.username
			  FROM reservations r
			  LEFT JOIN showtimes s ON r.showtime_id = s.id
			  LEFT JOIN movies m ON s.movie_id = m.id
			  LEFT JOIN theaters t ON s.theater_id = t.id
			  LEFT JOIN users u ON r.user_id = u.id
			  ORDER BY r.created_at DESC`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reservations []model.Reservation
	for rows.Next() {
		var reservation model.Reservation
		err := rows.Scan(&reservation.ID, &reservation.UserID, &reservation.ShowtimeID,
			&reservation.ReservationCode, &reservation.TotalSeats, &reservation.TotalAmount,
			&reservation.Status, &reservation.BookingDate, &reservation.CreatedAt, &reservation.UpdatedAt,
			&reservation.MovieTitle, &reservation.TheaterName, &reservation.ShowDate, &reservation.ShowTime,
			&reservation.Username)
		if err != nil {
			return nil, err
		}
		reservations = append(reservations, reservation)
	}
	return reservations, rows.Err()
}

// Cleanup function to release expired locks
func CleanupExpiredLocks() error {
	_, err := db.Exec(`UPDATE seat_reservations 
		SET status = 'available', locked_until = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE status = 'locked' AND locked_until < NOW()`)
	return err
}
