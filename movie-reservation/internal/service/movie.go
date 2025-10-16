package service

import (
	"errors"
	"movie-reservation/internal/db"
	"movie-reservation/model"
	"time"
)

// Movie service functions

func CreateMovie(req model.CreateMovieRequest) (*model.Movie, error) {
	// Validate genre exists
	_, err := db.GetGenreByID(req.GenreID)
	if err != nil {
		return nil, errors.New("invalid genre ID")
	}

	// Validate required fields
	if req.Title == "" {
		return nil, errors.New("movie title is required")
	}
	if req.DurationMinutes <= 0 {
		return nil, errors.New("duration must be greater than 0")
	}

	movieID, err := db.CreateMovie(req)
	if err != nil {
		return nil, err
	}

	return db.GetMovieByID(movieID)
}

func UpdateMovie(id int64, req model.CreateMovieRequest) (*model.Movie, error) {
	// Check if movie exists
	_, err := db.GetMovieByID(id)
	if err != nil {
		return nil, errors.New("movie not found")
	}

	// Validate genre exists
	_, err = db.GetGenreByID(req.GenreID)
	if err != nil {
		return nil, errors.New("invalid genre ID")
	}

	err = db.UpdateMovie(id, req)
	if err != nil {
		return nil, err
	}

	return db.GetMovieByID(id)
}

func DeleteMovie(id int64) error {
	_, err := db.GetMovieByID(id)
	if err != nil {
		return errors.New("movie not found")
	}

	return db.DeactivateMovie(id)
}

func GetAllMovies(includeInactive bool) ([]model.Movie, error) {
	return db.GetAllMovies(includeInactive)
}

func GetMovieByID(id int64) (*model.Movie, error) {
	return db.GetMovieByID(id)
}

func SearchMovies(title, genre, language string) ([]model.Movie, error) {
	return db.SearchMovies(title, genre, language)
}

func GetMoviesWithUpcomingShowtimes() ([]model.Movie, error) {
	return db.GetMoviesWithUpcomingShowtimes(time.Now())
}

// Genre service functions

func CreateGenre(name, description string) (*model.Genre, error) {
	if name == "" {
		return nil, errors.New("genre name is required")
	}

	genreID, err := db.CreateGenre(name, description)
	if err != nil {
		return nil, err
	}

	return db.GetGenreByID(genreID)
}

func GetAllGenres() ([]model.Genre, error) {
	return db.GetAllGenres()
}

func GetMoviesByGenre(genreID int64) ([]model.Movie, error) {
	// Validate genre exists
	_, err := db.GetGenreByID(genreID)
	if err != nil {
		return nil, errors.New("genre not found")
	}

	return db.GetMoviesByGenre(genreID)
}

// Theater service functions

func GetAllTheaters() ([]model.Theater, error) {
	return db.GetAllTheaters()
}

func GetTheaterByID(id int64) (*model.Theater, error) {
	return db.GetTheaterByID(id)
}
