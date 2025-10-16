package routes

import (
	"movie-reservation/internal/controller"
	"movie-reservation/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// Enable CORS middleware for web applications
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Public routes (no authentication required)
	setupPublicRoutes(r)

	// Protected routes (JWT authentication required)
	setupProtectedRoutes(r)

	// Admin routes (JWT + admin role required)
	setupAdminRoutes(r)
}

func setupPublicRoutes(r *gin.Engine) {
	// Authentication
	r.POST("/register", controller.RegisterHandler)
	r.POST("/login", controller.LoginHandler)

	// Public movie browsing
	r.GET("/movies", controller.GetMoviesHandler)
	r.GET("/movies/upcoming", controller.GetUpcomingMoviesHandler)
	r.GET("/movies/:id", controller.GetMovieByIDHandler)
	r.GET("/movies/search", controller.SearchMoviesHandler)

	// Genres
	r.GET("/genres", controller.GetGenresHandler)
	r.GET("/genres/:id/movies", controller.GetMoviesByGenreHandler)

	// Showtimes (public viewing)
	r.GET("/movies/:id/showtimes", controller.GetShowtimesByMovieHandler)
	r.GET("/showtimes/date/:date", controller.GetShowtimesByDateHandler)
	r.GET("/showtimes/:id", controller.GetShowtimeByIDHandler)

	// Theaters
	r.GET("/theaters", controller.GetTheatersHandler)

	// Seat availability (no auth needed for checking)
	r.GET("/showtimes/:id/seats", controller.GetSeatAvailabilityHandler)
}

func setupProtectedRoutes(r *gin.Engine) {
	api := r.Group("/api")
	api.Use(middleware.JWTAuthMiddleware())

	// User reservations
	api.POST("/reservations", controller.CreateReservationHandler)
	api.GET("/reservations", controller.GetUserReservationsHandler)
	api.GET("/reservations/:id", controller.GetReservationByIDHandler)
	api.DELETE("/reservations/:id", controller.CancelReservationHandler)
}

func setupAdminRoutes(r *gin.Engine) {
	admin := r.Group("/admin")
	admin.Use(middleware.JWTAuthMiddleware())
	admin.Use(middleware.AdminAuthMiddleware())

	// Movie management
	admin.POST("/movies", controller.CreateMovieHandler)
	admin.PUT("/movies/:id", controller.UpdateMovieHandler)
	admin.DELETE("/movies/:id", controller.DeleteMovieHandler)
	admin.GET("/movies", controller.GetAllMoviesAdminHandler)

	// Genre management
	admin.POST("/genres", controller.CreateGenreHandler)

	// Showtime management
	admin.POST("/showtimes", controller.CreateShowtimeHandler)
	admin.DELETE("/showtimes/:id", controller.DeleteShowtimeHandler)

	// Reservation management
	admin.GET("/reservations", controller.GetAllReservationsHandler)
	admin.GET("/showtimes/:id/reservations", controller.GetShowtimeReservationsHandler)

	// Reports and analytics
	admin.GET("/reports/revenue", controller.GetRevenueReportHandler)
	admin.GET("/reports/capacity", controller.GetCapacityReportHandler)
	admin.GET("/reports/popular-movies", controller.GetPopularMoviesHandler)
	admin.GET("/dashboard", controller.GetDashboardHandler)
}
