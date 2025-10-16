package model

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID        int64     `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Password  string    `json:"password,omitempty" db:"password_hash"` // omitempty prevents password from being returned in JSON
	Role      string    `json:"role" db:"role"`
	Email     string    `json:"email" db:"email"`
	FullName  string    `json:"full_name" db:"full_name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Genre represents a movie genre
type Genre struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Movie represents a movie in the system
type Movie struct {
	ID              int64     `json:"id" db:"id"`
	Title           string    `json:"title" db:"title"`
	Description     string    `json:"description" db:"description"`
	PosterImage     string    `json:"poster_image" db:"poster_image"`
	GenreID         int64     `json:"genre_id" db:"genre_id"`
	GenreName       string    `json:"genre_name,omitempty"` // For joined queries
	DurationMinutes int       `json:"duration_minutes" db:"duration_minutes"`
	ReleaseDate     time.Time `json:"release_date" db:"release_date"`
	Rating          string    `json:"rating" db:"rating"`
	Language        string    `json:"language" db:"language"`
	Director        string    `json:"director" db:"director"`
	CastMembers     string    `json:"cast_members" db:"cast_members"` // JSON string
	IsActive        bool      `json:"is_active" db:"is_active"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// Theater represents a theater/hall
type Theater struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Location    string    `json:"location" db:"location"`
	TotalSeats  int       `json:"total_seats" db:"total_seats"`
	RowsCount   int       `json:"rows_count" db:"rows_count"`
	SeatsPerRow int       `json:"seats_per_row" db:"seats_per_row"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Seat represents a seat in a theater
type Seat struct {
	ID         int64     `json:"id" db:"id"`
	TheaterID  int64     `json:"theater_id" db:"theater_id"`
	RowLabel   string    `json:"row_label" db:"row_label"`
	SeatNumber int       `json:"seat_number" db:"seat_number"`
	SeatType   string    `json:"seat_type" db:"seat_type"`
	IsActive   bool      `json:"is_active" db:"is_active"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// Showtime represents a movie showtime
type Showtime struct {
	ID             int64     `json:"id" db:"id"`
	MovieID        int64     `json:"movie_id" db:"movie_id"`
	MovieTitle     string    `json:"movie_title,omitempty"` // For joined queries
	TheaterID      int64     `json:"theater_id" db:"theater_id"`
	TheaterName    string    `json:"theater_name,omitempty"` // For joined queries
	ShowDate       time.Time `json:"show_date" db:"show_date"`
	ShowTime       string    `json:"show_time" db:"show_time"`
	Price          float64   `json:"price" db:"price"`
	AvailableSeats int       `json:"available_seats" db:"available_seats"`
	IsActive       bool      `json:"is_active" db:"is_active"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// Reservation represents a user's reservation
type Reservation struct {
	ID              int64     `json:"id" db:"id"`
	UserID          int64     `json:"user_id" db:"user_id"`
	ShowtimeID      int64     `json:"showtime_id" db:"showtime_id"`
	ReservationCode string    `json:"reservation_code" db:"reservation_code"`
	TotalSeats      int       `json:"total_seats" db:"total_seats"`
	TotalAmount     float64   `json:"total_amount" db:"total_amount"`
	Status          string    `json:"status" db:"status"`
	BookingDate     time.Time `json:"booking_date" db:"booking_date"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`

	// Additional fields for detailed reservation info
	MovieTitle    string            `json:"movie_title,omitempty"`
	TheaterName   string            `json:"theater_name,omitempty"`
	ShowDate      time.Time         `json:"show_date,omitempty"`
	ShowTime      string            `json:"show_time,omitempty"`
	Username      string            `json:"username,omitempty"`
	ReservedSeats []ReservationSeat `json:"reserved_seats,omitempty"`
}

// ReservationSeat represents individual seats in a reservation
type ReservationSeat struct {
	ID            int64     `json:"id" db:"id"`
	ReservationID int64     `json:"reservation_id" db:"reservation_id"`
	SeatID        int64     `json:"seat_id" db:"seat_id"`
	Price         float64   `json:"price" db:"price"`
	RowLabel      string    `json:"row_label,omitempty"`
	SeatNumber    int       `json:"seat_number,omitempty"`
	SeatType      string    `json:"seat_type,omitempty"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// SeatReservation represents seat availability for a specific showtime
type SeatReservation struct {
	ID            int64      `json:"id" db:"id"`
	ShowtimeID    int64      `json:"showtime_id" db:"showtime_id"`
	SeatID        int64      `json:"seat_id" db:"seat_id"`
	ReservationID *int64     `json:"reservation_id" db:"reservation_id"`
	Status        string     `json:"status" db:"status"`
	LockedUntil   *time.Time `json:"locked_until" db:"locked_until"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

// DTOs for API requests/responses
type CreateMovieRequest struct {
	Title           string    `json:"title" binding:"required"`
	Description     string    `json:"description"`
	PosterImage     string    `json:"poster_image"`
	GenreID         int64     `json:"genre_id" binding:"required"`
	DurationMinutes int       `json:"duration_minutes" binding:"required"`
	ReleaseDate     time.Time `json:"release_date" binding:"required"`
	Rating          string    `json:"rating"`
	Language        string    `json:"language"`
	Director        string    `json:"director"`
	CastMembers     string    `json:"cast_members"`
}

type CreateShowtimeRequest struct {
	MovieID   int64   `json:"movie_id" binding:"required"`
	TheaterID int64   `json:"theater_id" binding:"required"`
	ShowDate  string  `json:"show_date" binding:"required"` // Format: "2024-01-15"
	ShowTime  string  `json:"show_time" binding:"required"` // Format: "14:30"
	Price     float64 `json:"price" binding:"required"`
}

type CreateReservationRequest struct {
	ShowtimeID int64   `json:"showtime_id" binding:"required"`
	SeatIDs    []int64 `json:"seat_ids" binding:"required"`
}

type SeatAvailabilityResponse struct {
	Seat   Seat    `json:"seat"`
	Status string  `json:"status"` // available, reserved, locked
	Price  float64 `json:"price"`
}

type ReservationResponse struct {
	Reservation Reservation       `json:"reservation"`
	Seats       []ReservationSeat `json:"seats"`
}

// Admin reporting DTOs
type RevenueReport struct {
	TotalReservations int            `json:"total_reservations"`
	TotalRevenue      float64        `json:"total_revenue"`
	Period            string         `json:"period"`
	MovieBreakdown    []MovieRevenue `json:"movie_breakdown,omitempty"`
}

type MovieRevenue struct {
	MovieID      int64   `json:"movie_id"`
	MovieTitle   string  `json:"movie_title"`
	Reservations int     `json:"reservations"`
	Revenue      float64 `json:"revenue"`
}

type CapacityReport struct {
	TheaterID       int64   `json:"theater_id"`
	TheaterName     string  `json:"theater_name"`
	TotalSeats      int     `json:"total_seats"`
	ReservedSeats   int     `json:"reserved_seats"`
	CapacityPercent float64 `json:"capacity_percent"`
	Period          string  `json:"period"`
}
