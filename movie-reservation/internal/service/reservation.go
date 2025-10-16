package service

import (
	"errors"
	"fmt"
	"movie-reservation/internal/db"
	"movie-reservation/model"
	"movie-reservation/utils"
	"time"
)

// Reservation service with complex business logic and concurrency handling

func GetSeatAvailability(showtimeID int64) ([]model.SeatAvailabilityResponse, error) {
	// Validate showtime exists
	_, err := db.GetShowtimeByID(showtimeID)
	if err != nil {
		return nil, errors.New("showtime not found")
	}

	// Clean up expired locks before checking availability
	err = db.CleanupExpiredLocks()
	if err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: Failed to cleanup expired locks: %v\n", err)
	}

	availability, err := db.GetSeatAvailabilityForShowtime(showtimeID)
	if err != nil {
		return nil, err
	}
	if len(availability) == 0 {
		// Attempt to backfill and re-query once
		if err := db.EnsureSeatReservations(showtimeID); err == nil {
			availability, _ = db.GetSeatAvailabilityForShowtime(showtimeID)
		}
	}
	return availability, nil
}

// parseShowDateTime attempts to parse a date (time.Time, date part only) and a time string
// accepting layouts HH:MM or HH:MM:SS, returning a combined time.Time.
func parseShowDateTime(showDate time.Time, showTimeStr string) (time.Time, error) {
	datePart := showDate.Format("2006-01-02")
	// Try with seconds first
	if t, err := time.Parse("2006-01-02 15:04:05", datePart+" "+showTimeStr); err == nil {
		return t, nil
	}
	// Fallback HH:MM
	if t, err := time.Parse("2006-01-02 15:04", datePart+" "+showTimeStr[:5]); err == nil {
		return t, nil
	}
	return time.Time{}, errors.New("parse error")
}

func CreateReservation(userID int64, req model.CreateReservationRequest) (*model.Reservation, error) {
	// Validate showtime exists and is in the future
	showtime, err := db.GetShowtimeByID(req.ShowtimeID)
	if err != nil {
		return nil, errors.New("showtime not found")
	}

	// Ensure seat_reservations exist for this showtime (idempotent)
	_ = db.EnsureSeatReservations(req.ShowtimeID)

	// Check if showtime is in the future (support HH:MM or HH:MM:SS)
	showDateTime, err := parseShowDateTime(showtime.ShowDate, showtime.ShowTime)
	if err != nil {
		return nil, errors.New("invalid showtime format")
	}

	cutoff := utils.BookingCutoffDuration()
	if time.Now().After(showDateTime.Add(-cutoff)) {
		return nil, fmt.Errorf("cannot reserve seats within %v of showtime", cutoff)
	}

	// Validate seat selection
	if len(req.SeatIDs) == 0 {
		return nil, errors.New("at least one seat must be selected")
	}
	if len(req.SeatIDs) > 8 { // Maximum 8 seats per reservation
		return nil, errors.New("maximum 8 seats allowed per reservation")
	}

	// Validate all seats belong to the same theater as the showtime
	seats, err := db.GetSeatsByIDs(req.SeatIDs)
	if err != nil {
		return nil, errors.New("invalid seat selection")
	}

	if len(seats) != len(req.SeatIDs) {
		return nil, errors.New("some selected seats do not exist")
	}

	// Check if all seats belong to the showtime's theater
	for _, seat := range seats {
		if seat.TheaterID != showtime.TheaterID {
			return nil, fmt.Errorf("seat %s%d does not belong to the showtime's theater", seat.RowLabel, seat.SeatNumber)
		}
	}

	// Step 1: Lock seats temporarily (5 minutes to complete payment)
	lockDuration := 5 * time.Minute
	err = db.LockSeats(req.ShowtimeID, req.SeatIDs, lockDuration)
	if err != nil {
		return nil, fmt.Errorf("seat locking failed: %v", err)
	}

	// Step 2: Calculate total amount
	totalAmount := calculateTotalAmount(showtime.Price, seats)

	// Step 3: Create the reservation (this will handle the transaction)
	reservation, err := db.CreateReservation(userID, req.ShowtimeID, req.SeatIDs, totalAmount)
	if err != nil {
		// Release locks if reservation creation fails
		db.ReleaseSeatLocks(req.ShowtimeID, req.SeatIDs)
		return nil, fmt.Errorf("failed to create reservation: %v", err)
	}

	return reservation, nil
}

func calculateTotalAmount(basePrice float64, seats []model.Seat) float64 {
	total := 0.0
	for _, seat := range seats {
		seatPrice := basePrice
		switch seat.SeatType {
		case "premium":
			seatPrice = basePrice * 1.5
		case "vip":
			seatPrice = basePrice * 2.0
		}
		total += seatPrice
	}
	return total
}

func GetUserReservations(userID int64) ([]model.Reservation, error) {
	return db.GetUserReservations(userID)
}

func GetReservationByID(reservationID, userID int64) (*model.Reservation, error) {
	reservation, err := db.GetReservationByID(reservationID)
	if err != nil {
		return nil, errors.New("reservation not found")
	}

	// Verify the reservation belongs to the user
	if reservation.UserID != userID {
		return nil, errors.New("reservation not found")
	}

	return reservation, nil
}

func CancelReservation(reservationID, userID int64) error {
	// Get reservation to verify ownership and check if it's cancellable
	reservation, err := db.GetReservationByID(reservationID)
	if err != nil {
		return errors.New("reservation not found")
	}

	if reservation.UserID != userID {
		return errors.New("reservation not found")
	}

	if reservation.Status != "confirmed" {
		return errors.New("only confirmed reservations can be cancelled")
	}

	// Check if the show is at least 2 hours in the future
	showDateTime, err := parseShowDateTime(reservation.ShowDate, reservation.ShowTime)
	if err != nil {
		return errors.New("invalid showtime format")
	}

	cancelCutoff := utils.CancelCutoffDuration()
	if time.Now().After(showDateTime.Add(-cancelCutoff)) {
		return fmt.Errorf("cannot cancel reservations within %v of showtime", cancelCutoff)
	}

	return db.CancelReservation(reservationID, userID)
}

// Admin functions

func GetAllReservations() ([]model.Reservation, error) {
	return db.GetAllReservations()
}

func GetReservationsByShowtime(showtimeID int64) ([]model.Reservation, error) {
	// First validate showtime exists
	_, err := db.GetShowtimeByID(showtimeID)
	if err != nil {
		return nil, errors.New("showtime not found")
	}

	reservations, err := db.GetAllReservations()
	if err != nil {
		return nil, err
	}

	// Filter by showtime
	var filtered []model.Reservation
	for _, res := range reservations {
		if res.ShowtimeID == showtimeID {
			filtered = append(filtered, res)
		}
	}

	return filtered, nil
}

func GetOccupancyRate(showtimeID int64) (float64, error) {
	_, err := db.GetShowtimeByID(showtimeID)
	if err != nil {
		return 0, errors.New("showtime not found")
	}

	return db.GetOccupancyRate(showtimeID)
}

// Background cleanup service
func CleanupExpiredLocks() error {
	return db.CleanupExpiredLocks()
}

// Validate seat adjacency (optional business rule)
func ValidateSeatAdjacency(seatIDs []int64) error {
	if len(seatIDs) <= 1 {
		return nil // Single seat or no seats don't need adjacency check
	}

	seats, err := db.GetSeatsByIDs(seatIDs)
	if err != nil {
		return err
	}

	// Group seats by row
	seatsByRow := make(map[string][]model.Seat)
	for _, seat := range seats {
		seatsByRow[seat.RowLabel] = append(seatsByRow[seat.RowLabel], seat)
	}

	// Check if seats are in adjacent rows or same row
	if len(seatsByRow) > 2 {
		return errors.New("seats must be in the same row or adjacent rows")
	}

	// For seats in the same row, check if they are consecutive
	for _, rowSeats := range seatsByRow {
		if len(rowSeats) > 1 {
			// Sort seats by number and check consecutiveness
			for i := 1; i < len(rowSeats); i++ {
				if rowSeats[i].SeatNumber-rowSeats[i-1].SeatNumber > 1 {
					return errors.New("seats in the same row must be consecutive")
				}
			}
		}
	}

	return nil
}
