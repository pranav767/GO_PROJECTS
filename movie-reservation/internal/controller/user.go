package controller

import (
	"movie-reservation/internal/service"
	"movie-reservation/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Public Movie Browsing Controllers (No authentication required)

// GetMoviesHandler returns active movies for public viewing
func GetMoviesHandler(c *gin.Context) {
	movies, err := service.GetAllMovies(false) // Only active movies
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch movies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"movies": movies})
}

// GetMovieByIDHandler returns a specific movie by ID
func GetMovieByIDHandler(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie ID"})
		return
	}

	movie, err := service.GetMovieByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"movie": movie})
}

// SearchMoviesHandler handles movie search
func SearchMoviesHandler(c *gin.Context) {
	title := c.Query("title")
	genre := c.Query("genre")
	language := c.Query("language")

	movies, err := service.SearchMovies(title, genre, language)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search movies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"movies": movies})
}

// GetGenresHandler returns all available genres
func GetGenresHandler(c *gin.Context) {
	genres, err := service.GetAllGenres()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch genres"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"genres": genres})
}

// GetMoviesByGenreHandler returns movies by genre
func GetMoviesByGenreHandler(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid genre ID"})
		return
	}

	movies, err := service.GetMoviesByGenre(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"movies": movies})
}

// Showtime Controllers

// GetShowtimesByMovieHandler returns showtimes for a specific movie
func GetShowtimesByMovieHandler(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie ID"})
		return
	}

	showtimes, err := service.GetShowtimesByMovie(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch showtimes"})
		return
	}

	// Ensure empty slice ([]) instead of null when no showtimes
	if showtimes == nil {
		showtimes = []model.Showtime{}
	}
	c.JSON(http.StatusOK, gin.H{"showtimes": showtimes})
}

// GetShowtimesByDateHandler returns all showtimes for a specific date
func GetShowtimesByDateHandler(c *gin.Context) {
	date := c.Param("date")
	if date == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Date parameter is required (format: YYYY-MM-DD)"})
		return
	}

	showtimes, err := service.GetShowtimesByDate(date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if showtimes == nil {
		showtimes = []model.Showtime{}
	}
	c.JSON(http.StatusOK, gin.H{"showtimes": showtimes})
}

// GetShowtimeByIDHandler returns a specific showtime
func GetShowtimeByIDHandler(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid showtime ID"})
		return
	}

	showtime, err := service.GetShowtimeByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Showtime not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"showtime": showtime})
}

// Theater Controllers

// GetTheatersHandler returns all theaters
func GetTheatersHandler(c *gin.Context) {
	theaters, err := service.GetAllTheaters()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch theaters"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"theaters": theaters})
}

// Seat and Reservation Controllers (Authentication required)

// GetSeatAvailabilityHandler returns seat availability for a showtime
func GetSeatAvailabilityHandler(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid showtime ID"})
		return
	}

	availability, err := service.GetSeatAvailability(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"seat_availability": availability})
}

// CreateReservationHandler handles seat reservations
func CreateReservationHandler(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	userID, err := strconv.ParseInt(userIDStr.(string), 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	var req model.CreateReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	reservation, err := service.CreateReservation(userID, req)
	if err != nil {
		errMsg := err.Error()
		status := http.StatusBadRequest
		switch errMsg {
		case "showtime not found":
			status = http.StatusNotFound
		case "at least one seat must be selected", "maximum 8 seats allowed per reservation", "invalid seat selection", "some selected seats do not exist":
			status = http.StatusUnprocessableEntity
		}
		if len(errMsg) >= 18 && errMsg[:18] == "seat locking failed" { // concurrency/conflict
			status = http.StatusConflict
		}
		if len(errMsg) >= 23 && errMsg[:23] == "cannot reserve seats wit" { // dynamic cutoff message
			status = http.StatusUnprocessableEntity
		}
		c.JSON(status, gin.H{"error": errMsg})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "reservation created successfully", "reservation": reservation})
}

// GetUserReservationsHandler returns all reservations for the authenticated user
func GetUserReservationsHandler(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	userID, err := strconv.ParseInt(userIDStr.(string), 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	reservations, err := service.GetUserReservations(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reservations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"reservations": reservations})
}

// GetReservationByIDHandler returns a specific reservation for the authenticated user
func GetReservationByIDHandler(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	userID, err := strconv.ParseInt(userIDStr.(string), 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	reservationID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reservation ID"})
		return
	}

	reservation, err := service.GetReservationByID(reservationID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"reservation": reservation})
}

// CancelReservationHandler handles reservation cancellation
func CancelReservationHandler(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	userID, err := strconv.ParseInt(userIDStr.(string), 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	reservationID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reservation ID"})
		return
	}

	err = service.CancelReservation(reservationID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reservation cancelled successfully"})
}

// GetUpcomingMoviesHandler returns movies with upcoming showtimes
func GetUpcomingMoviesHandler(c *gin.Context) {
	movies, err := service.GetMoviesWithUpcomingShowtimes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch upcoming movies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"movies": movies})
}
