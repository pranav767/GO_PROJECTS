package service

import (
	"errors"
	"fmt"
	"movie-reservation/internal/db"
	"movie-reservation/model"
	"time"
)

// Showtime service functions

func CreateShowtime(req model.CreateShowtimeRequest) (*model.Showtime, error) {
	// Validate movie exists and is active
	movie, err := db.GetMovieByID(req.MovieID)
	if err != nil {
		return nil, errors.New("movie not found")
	}
	if !movie.IsActive {
		return nil, errors.New("movie is not active")
	}

	// Validate theater exists and is active
	theater, err := db.GetTheaterByID(req.TheaterID)
	if err != nil {
		return nil, errors.New("theater not found")
	}
	if !theater.IsActive {
		return nil, errors.New("theater is not active")
	}

	// Validate show date (must be in the future)
	showDate, err := time.Parse("2006-01-02", req.ShowDate)
	if err != nil {
		return nil, errors.New("invalid date format, use YYYY-MM-DD")
	}
	if showDate.Before(time.Now().Truncate(24 * time.Hour)) {
		return nil, errors.New("show date must be today or in the future")
	}

	// Validate show time format
	_, err = time.Parse("15:04", req.ShowTime)
	if err != nil {
		return nil, errors.New("invalid time format, use HH:MM")
	}

	// Validate price
	if req.Price <= 0 {
		return nil, errors.New("price must be greater than 0")
	}

	// Check for scheduling conflicts (same theater, overlapping times)
	err = validateShowtimeSchedule(req, movie.DurationMinutes)
	if err != nil {
		return nil, err
	}

	showtimeID, err := db.CreateShowtime(req)
	if err != nil {
		return nil, err
	}

	return db.GetShowtimeByID(showtimeID)
}

func validateShowtimeSchedule(req model.CreateShowtimeRequest, durationMinutes int) error {
	// Parse the new showtime
	showDateTime, err := time.Parse("2006-01-02 15:04", req.ShowDate+" "+req.ShowTime)
	if err != nil {
		return err
	}

	// Calculate end time (duration + 30 minutes buffer for cleaning)
	endTime := showDateTime.Add(time.Duration(durationMinutes+30) * time.Minute)

	// Get existing showtimes for the same theater and date
	showDate, _ := time.Parse("2006-01-02", req.ShowDate)
	existingShowtimes, err := db.GetShowtimesByDate(showDate)
	if err != nil {
		return err
	}

	for _, existing := range existingShowtimes {
		if existing.TheaterID != req.TheaterID {
			continue
		}

		// Parse existing showtime
		existingDateTime, err := time.Parse("2006-01-02 15:04", existing.ShowDate.Format("2006-01-02")+" "+existing.ShowTime)
		if err != nil {
			continue
		}

		// Get existing movie duration
		existingMovie, err := db.GetMovieByID(existing.MovieID)
		if err != nil {
			continue
		}

		existingEndTime := existingDateTime.Add(time.Duration(existingMovie.DurationMinutes+30) * time.Minute)

		// Check for overlap
		if showDateTime.Before(existingEndTime) && endTime.After(existingDateTime) {
			return fmt.Errorf("showtime conflicts with existing show at %s", existing.ShowTime)
		}
	}

	return nil
}

func GetShowtimesByMovie(movieID int64) ([]model.Showtime, error) {
	return db.GetShowtimesByMovie(movieID)
}

func GetShowtimesByDate(date string) ([]model.Showtime, error) {
	showDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, errors.New("invalid date format, use YYYY-MM-DD")
	}

	return db.GetShowtimesByDate(showDate)
}

func GetShowtimeByID(id int64) (*model.Showtime, error) {
	return db.GetShowtimeByID(id)
}

func DeactivateShowtime(id int64) error {
	// Check if showtime exists
	_, err := db.GetShowtimeByID(id)
	if err != nil {
		return errors.New("showtime not found")
	}

	return db.DeactivateShowtime(id)
}

// Get available showtimes for today and future dates
func GetUpcomingShowtimes() ([]model.Showtime, error) {
	return db.GetShowtimesByDate(time.Now())
}
