package controller

import (
	"movie-reservation/internal/service"
	"movie-reservation/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Admin Movie Management Controllers

// CreateMovieHandler handles movie creation (Admin only)
func CreateMovieHandler(c *gin.Context) {
	var req model.CreateMovieRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	movie, err := service.CreateMovie(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Movie created successfully", "movie": movie})
}

// UpdateMovieHandler handles movie updates (Admin only)
func UpdateMovieHandler(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie ID"})
		return
	}

	var req model.CreateMovieRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	movie, err := service.UpdateMovie(id, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Movie updated successfully", "movie": movie})
}

// DeleteMovieHandler handles movie deletion (Admin only)
func DeleteMovieHandler(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie ID"})
		return
	}

	err = service.DeleteMovie(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Movie deactivated successfully"})
}

// GetAllMoviesHandler returns all movies including inactive ones (Admin only)
func GetAllMoviesAdminHandler(c *gin.Context) {
	includeInactive := c.Query("include_inactive") == "true"

	movies, err := service.GetAllMovies(includeInactive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch movies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"movies": movies})
}

// Genre Management Controllers

// CreateGenreHandler handles genre creation (Admin only)
func CreateGenreHandler(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	genre, err := service.CreateGenre(req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Genre created successfully", "genre": genre})
}

// Showtime Management Controllers

// CreateShowtimeHandler handles showtime creation (Admin only)
func CreateShowtimeHandler(c *gin.Context) {
	var req model.CreateShowtimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	showtime, err := service.CreateShowtime(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Showtime created successfully", "showtime": showtime})
}

// DeleteShowtimeHandler handles showtime deactivation (Admin only)
func DeleteShowtimeHandler(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid showtime ID"})
		return
	}

	err = service.DeactivateShowtime(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Showtime deactivated successfully"})
}

// Admin Reporting Controllers

// GetRevenueReportHandler generates revenue reports (Admin only)
func GetRevenueReportHandler(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date parameters are required (format: YYYY-MM-DD)"})
		return
	}

	report, err := service.GenerateRevenueReport(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"report": report})
}

// GetCapacityReportHandler generates capacity reports (Admin only)
func GetCapacityReportHandler(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date parameters are required (format: YYYY-MM-DD)"})
		return
	}

	report, err := service.GenerateCapacityReport(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"report": report})
}

// GetPopularMoviesHandler returns popular movies report (Admin only)
func GetPopularMoviesHandler(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	limitStr := c.DefaultQuery("limit", "10")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date parameters are required (format: YYYY-MM-DD)"})
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	movies, err := service.GetPopularMovies(startDate, endDate, limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"popular_movies": movies})
}

// GetDashboardHandler returns admin dashboard data (Admin only)
func GetDashboardHandler(c *gin.Context) {
	data, err := service.GetDashboardData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch dashboard data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"dashboard": data})
}

// GetAllReservationsHandler returns all reservations (Admin only)
func GetAllReservationsHandler(c *gin.Context) {
	reservations, err := service.GetAllReservations()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reservations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"reservations": reservations})
}

// GetShowtimeReservationsHandler returns reservations for a specific showtime (Admin only)
func GetShowtimeReservationsHandler(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid showtime ID"})
		return
	}

	reservations, err := service.GetReservationsByShowtime(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Also get occupancy rate
	occupancy, err := service.GetOccupancyRate(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reservations":   reservations,
		"occupancy_rate": occupancy,
	})
}
