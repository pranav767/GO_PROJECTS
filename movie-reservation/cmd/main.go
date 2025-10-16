package main

import (
	"log"
	"movie-reservation/internal/db"
	"movie-reservation/internal/service"
	"movie-reservation/routes"
	"movie-reservation/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env if present
	_ = godotenv.Load()

	// Initialize database connection
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database connection: %v", err)
	}
	defer db.Close()

	// Initialize database schema and seed data
	// Optional full data reset (controlled by DB_RESET_ON_START)
	if err := db.ResetDatabase(); err != nil {
		log.Fatalf("Database reset failed: %v", err)
	}

	if err := db.InitializeDatabase(); err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}

	// Seed sample data for testing (only if tables are empty)
	if err := db.SeedSampleData(); err != nil {
		log.Printf("Warning: Failed to seed sample data: %v", err)
	}

	// Sync / override admin password if requested via env
	db.SyncAdminPasswordFromEnv()

	// Start background cleanup service for expired seat locks
	go startCleanupService()

	// Set Gin to release mode for production
	// gin.SetMode(gin.ReleaseMode)

	// Initialize Gin router
	r := gin.Default()

	// Setup routes
	routes.SetupRoutes(r)

	log.Println("🎬 Movie Reservation System Server Starting on :8080")
	log.Println("")
	log.Println("📋 Available Endpoints:")
	log.Println("")
	log.Println("🔓 Public Routes:")
	log.Println("  POST   /register                    - User registration")
	log.Println("  POST   /login                       - User login")
	log.Println("  GET    /movies                      - List all active movies")
	log.Println("  GET    /movies/upcoming             - Movies with upcoming shows")
	log.Println("  GET    /movies/:id                  - Get movie details")
	log.Println("  GET    /movies/search               - Search movies")
	log.Println("  GET    /genres                      - List all genres")
	log.Println("  GET    /genres/:id/movies           - Movies by genre")
	log.Println("  GET    /movies/:id/showtimes        - Showtimes for movie")
	log.Println("  GET    /showtimes/date/:date        - Showtimes by date")
	log.Println("  GET    /showtimes/:id               - Showtime details")
	log.Println("  GET    /theaters                    - List all theaters")
	log.Println("  GET    /showtimes/:id/seats         - Check seat availability")
	log.Println("")
	log.Println("🔐 Protected Routes (require JWT):")
	log.Println("  POST   /api/reservations            - Create reservation")
	log.Println("  GET    /api/reservations            - Get user reservations")
	log.Println("  GET    /api/reservations/:id        - Get reservation details")
	log.Println("  DELETE /api/reservations/:id        - Cancel reservation")
	log.Println("")
	log.Println("👑 Admin Routes (require admin role):")
	log.Println("  POST   /admin/movies                - Create movie")
	log.Println("  PUT    /admin/movies/:id            - Update movie")
	log.Println("  DELETE /admin/movies/:id            - Delete movie")
	log.Println("  GET    /admin/movies                - List all movies (including inactive)")
	log.Println("  POST   /admin/genres                - Create genre")
	log.Println("  POST   /admin/showtimes             - Create showtime")
	log.Println("  DELETE /admin/showtimes/:id         - Delete showtime")
	log.Println("  GET    /admin/reservations          - Get all reservations")
	log.Println("  GET    /admin/showtimes/:id/reservations - Reservations by showtime")
	log.Println("  GET    /admin/reports/revenue       - Revenue reports")
	log.Println("  GET    /admin/reports/capacity      - Capacity reports")
	log.Println("  GET    /admin/reports/popular-movies - Popular movies")
	log.Println("  GET    /admin/dashboard             - Admin dashboard")
	log.Println("")
	log.Println("💡 Default admin credentials:")
	log.Println("   Username: admin")
	log.Println("   Password: admin123")
	log.Println("")
	log.Println("🚀 Server is ready!")

	// Log effective dynamic configuration
	log.Printf("⏱ Booking cutoff: %v | Cancellation cutoff: %v", utils.BookingCutoffDuration(), utils.CancelCutoffDuration())

	// Start server
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

// Background service to cleanup expired seat locks
func startCleanupService() {
	ticker := time.NewTicker(1 * time.Minute) // Run every minute
	defer ticker.Stop()

	for range ticker.C {
		err := service.CleanupExpiredLocks()
		if err != nil {
			log.Printf("Cleanup service error: %v", err)
		}
	}
}
