package db

import (
	"movie-reservation/model"
	"strings"
	"time"
)

// Genre operations
func CreateGenre(name, description string) (int64, error) {
	result, err := db.Exec("INSERT INTO genres (name, description) VALUES (?, ?)", name, description)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func GetAllGenres() ([]model.Genre, error) {
	rows, err := db.Query("SELECT id, name, description, created_at FROM genres ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []model.Genre
	for rows.Next() {
		var genre model.Genre
		err := rows.Scan(&genre.ID, &genre.Name, &genre.Description, &genre.CreatedAt)
		if err != nil {
			return nil, err
		}
		genres = append(genres, genre)
	}
	return genres, rows.Err()
}

func GetGenreByID(id int64) (*model.Genre, error) {
	var genre model.Genre
	err := db.QueryRow("SELECT id, name, description, created_at FROM genres WHERE id = ?", id).
		Scan(&genre.ID, &genre.Name, &genre.Description, &genre.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &genre, nil
}

// Movie operations
func CreateMovie(movie model.CreateMovieRequest) (int64, error) {
	query := `INSERT INTO movies (title, description, poster_image, genre_id, duration_minutes, 
			  release_date, rating, language, director, cast_members) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := db.Exec(query, movie.Title, movie.Description, movie.PosterImage,
		movie.GenreID, movie.DurationMinutes, movie.ReleaseDate, movie.Rating,
		movie.Language, movie.Director, movie.CastMembers)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func GetAllMovies(includeInactive bool) ([]model.Movie, error) {
	query := `SELECT m.id, m.title, m.description, m.poster_image, m.genre_id, 
			  m.duration_minutes, m.release_date, m.rating, m.language, m.director, 
			  m.cast_members, m.is_active, m.created_at, m.updated_at, g.name as genre_name
			  FROM movies m 
			  LEFT JOIN genres g ON m.genre_id = g.id`

	if !includeInactive {
		query += " WHERE m.is_active = TRUE"
	}
	query += " ORDER BY m.created_at DESC"

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []model.Movie
	for rows.Next() {
		var movie model.Movie
		err := rows.Scan(&movie.ID, &movie.Title, &movie.Description, &movie.PosterImage,
			&movie.GenreID, &movie.DurationMinutes, &movie.ReleaseDate, &movie.Rating,
			&movie.Language, &movie.Director, &movie.CastMembers, &movie.IsActive,
			&movie.CreatedAt, &movie.UpdatedAt, &movie.GenreName)
		if err != nil {
			return nil, err
		}
		movies = append(movies, movie)
	}
	return movies, rows.Err()
}

func GetMovieByID(id int64) (*model.Movie, error) {
	query := `SELECT m.id, m.title, m.description, m.poster_image, m.genre_id, 
			  m.duration_minutes, m.release_date, m.rating, m.language, m.director, 
			  m.cast_members, m.is_active, m.created_at, m.updated_at, g.name as genre_name
			  FROM movies m 
			  LEFT JOIN genres g ON m.genre_id = g.id 
			  WHERE m.id = ?`

	var movie model.Movie
	err := db.QueryRow(query, id).Scan(&movie.ID, &movie.Title, &movie.Description,
		&movie.PosterImage, &movie.GenreID, &movie.DurationMinutes, &movie.ReleaseDate,
		&movie.Rating, &movie.Language, &movie.Director, &movie.CastMembers,
		&movie.IsActive, &movie.CreatedAt, &movie.UpdatedAt, &movie.GenreName)
	if err != nil {
		return nil, err
	}
	return &movie, nil
}

func UpdateMovie(id int64, movie model.CreateMovieRequest) error {
	query := `UPDATE movies SET title = ?, description = ?, poster_image = ?, genre_id = ?, 
			  duration_minutes = ?, release_date = ?, rating = ?, language = ?, 
			  director = ?, cast_members = ?, updated_at = CURRENT_TIMESTAMP 
			  WHERE id = ?`

	_, err := db.Exec(query, movie.Title, movie.Description, movie.PosterImage,
		movie.GenreID, movie.DurationMinutes, movie.ReleaseDate, movie.Rating,
		movie.Language, movie.Director, movie.CastMembers, id)
	return err
}

func DeactivateMovie(id int64) error {
	_, err := db.Exec("UPDATE movies SET is_active = FALSE WHERE id = ?", id)
	return err
}

func ActivateMovie(id int64) error {
	_, err := db.Exec("UPDATE movies SET is_active = TRUE WHERE id = ?", id)
	return err
}

func GetMoviesByGenre(genreID int64) ([]model.Movie, error) {
	query := `SELECT m.id, m.title, m.description, m.poster_image, m.genre_id, 
			  m.duration_minutes, m.release_date, m.rating, m.language, m.director, 
			  m.cast_members, m.is_active, m.created_at, m.updated_at, g.name as genre_name
			  FROM movies m 
			  LEFT JOIN genres g ON m.genre_id = g.id 
			  WHERE m.genre_id = ? AND m.is_active = TRUE
			  ORDER BY m.created_at DESC`

	rows, err := db.Query(query, genreID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []model.Movie
	for rows.Next() {
		var movie model.Movie
		err := rows.Scan(&movie.ID, &movie.Title, &movie.Description, &movie.PosterImage,
			&movie.GenreID, &movie.DurationMinutes, &movie.ReleaseDate, &movie.Rating,
			&movie.Language, &movie.Director, &movie.CastMembers, &movie.IsActive,
			&movie.CreatedAt, &movie.UpdatedAt, &movie.GenreName)
		if err != nil {
			return nil, err
		}
		movies = append(movies, movie)
	}
	return movies, rows.Err()
}

func SearchMovies(title, genre, language string) ([]model.Movie, error) {
	var conditions []string
	var args []interface{}

	conditions = append(conditions, "m.is_active = TRUE")

	if title != "" {
		conditions = append(conditions, "m.title LIKE ?")
		args = append(args, "%"+title+"%")
	}

	if genre != "" {
		conditions = append(conditions, "g.name LIKE ?")
		args = append(args, "%"+genre+"%")
	}

	if language != "" {
		conditions = append(conditions, "m.language LIKE ?")
		args = append(args, "%"+language+"%")
	}

	query := `SELECT m.id, m.title, m.description, m.poster_image, m.genre_id, 
			  m.duration_minutes, m.release_date, m.rating, m.language, m.director, 
			  m.cast_members, m.is_active, m.created_at, m.updated_at, g.name as genre_name
			  FROM movies m 
			  LEFT JOIN genres g ON m.genre_id = g.id 
			  WHERE ` + strings.Join(conditions, " AND ") + `
			  ORDER BY m.created_at DESC`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []model.Movie
	for rows.Next() {
		var movie model.Movie
		err := rows.Scan(&movie.ID, &movie.Title, &movie.Description, &movie.PosterImage,
			&movie.GenreID, &movie.DurationMinutes, &movie.ReleaseDate, &movie.Rating,
			&movie.Language, &movie.Director, &movie.CastMembers, &movie.IsActive,
			&movie.CreatedAt, &movie.UpdatedAt, &movie.GenreName)
		if err != nil {
			return nil, err
		}
		movies = append(movies, movie)
	}
	return movies, rows.Err()
}

// GetMoviesWithUpcomingShowtimes returns movies that have showtimes from today onwards
func GetMoviesWithUpcomingShowtimes(date time.Time) ([]model.Movie, error) {
	query := `SELECT DISTINCT m.id, m.title, m.description, m.poster_image, m.genre_id, 
			  m.duration_minutes, m.release_date, m.rating, m.language, m.director, 
			  m.cast_members, m.is_active, m.created_at, m.updated_at, g.name as genre_name
			  FROM movies m 
			  LEFT JOIN genres g ON m.genre_id = g.id 
			  INNER JOIN showtimes s ON m.id = s.movie_id
			  WHERE m.is_active = TRUE AND s.is_active = TRUE AND s.show_date >= ?
			  ORDER BY m.created_at DESC`

	rows, err := db.Query(query, date.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []model.Movie
	for rows.Next() {
		var movie model.Movie
		err := rows.Scan(&movie.ID, &movie.Title, &movie.Description, &movie.PosterImage,
			&movie.GenreID, &movie.DurationMinutes, &movie.ReleaseDate, &movie.Rating,
			&movie.Language, &movie.Director, &movie.CastMembers, &movie.IsActive,
			&movie.CreatedAt, &movie.UpdatedAt, &movie.GenreName)
		if err != nil {
			return nil, err
		}
		movies = append(movies, movie)
	}
	return movies, rows.Err()
}
