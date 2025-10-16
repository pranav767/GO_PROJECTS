-- Users table with role-based authentication
CREATE TABLE IF NOT EXISTS users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role ENUM('user', 'admin') DEFAULT 'user',
    email VARCHAR(255),
    full_name VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Genres table for movie categorization
CREATE TABLE IF NOT EXISTS genres (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Movies table
CREATE TABLE IF NOT EXISTS movies (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    poster_image VARCHAR(500),
    genre_id BIGINT,
    duration_minutes INT NOT NULL,
    release_date DATE,
    director VARCHAR(255),
    cast_members TEXT, -- JSON array of cast members
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (genre_id) REFERENCES genres(id) ON DELETE SET NULL,
    INDEX idx_movie_genre (genre_id),
    INDEX idx_movie_active (is_active),
    INDEX idx_movie_release (release_date)
);

-- Theaters/Halls table
CREATE TABLE IF NOT EXISTS theaters (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    location VARCHAR(500),
    total_seats INT NOT NULL,
    rows_count INT NOT NULL,
    seats_per_row INT NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Seats table for each theater
CREATE TABLE IF NOT EXISTS seats (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    theater_id BIGINT NOT NULL,
    row_label CHAR(2) NOT NULL, -- A, B, C, etc.
    seat_number INT NOT NULL,
    seat_type ENUM('regular', 'premium', 'vip') DEFAULT 'regular',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (theater_id) REFERENCES theaters(id) ON DELETE CASCADE,
    UNIQUE KEY unique_theater_seat (theater_id, row_label, seat_number),
    INDEX idx_theater_seats (theater_id),
    INDEX idx_seat_active (is_active)
);

-- Showtimes table
CREATE TABLE IF NOT EXISTS showtimes (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    movie_id BIGINT NOT NULL,
    theater_id BIGINT NOT NULL,
    show_date DATE NOT NULL,
    show_time TIME NOT NULL,
    price DECIMAL(8,2) NOT NULL,
    available_seats INT NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE CASCADE,
    FOREIGN KEY (theater_id) REFERENCES theaters(id) ON DELETE CASCADE,
    UNIQUE KEY unique_showtime (movie_id, theater_id, show_date, show_time),
    INDEX idx_showtime_date (show_date),
    INDEX idx_showtime_movie (movie_id),
    INDEX idx_showtime_theater (theater_id),
    INDEX idx_showtime_active (is_active)
);

-- Reservations table
CREATE TABLE IF NOT EXISTS reservations (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    showtime_id BIGINT NOT NULL,
    reservation_code VARCHAR(20) UNIQUE NOT NULL,
    total_seats INT NOT NULL,
    total_amount DECIMAL(10,2) NOT NULL,
    status ENUM('confirmed', 'cancelled', 'completed') DEFAULT 'confirmed',
    booking_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (showtime_id) REFERENCES showtimes(id) ON DELETE CASCADE,
    INDEX idx_reservation_user (user_id),
    INDEX idx_reservation_showtime (showtime_id),
    INDEX idx_reservation_status (status),
    INDEX idx_reservation_date (booking_date)
);

-- Reservation seats table (many-to-many relationship)
CREATE TABLE IF NOT EXISTS reservation_seats (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    reservation_id BIGINT NOT NULL,
    seat_id BIGINT NOT NULL,
    price DECIMAL(8,2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (reservation_id) REFERENCES reservations(id) ON DELETE CASCADE,
    FOREIGN KEY (seat_id) REFERENCES seats(id) ON DELETE CASCADE,
    UNIQUE KEY unique_reservation_seat (reservation_id, seat_id),
    INDEX idx_res_seat_reservation (reservation_id),
    INDEX idx_res_seat_seat (seat_id)
);

-- Seat availability for specific showtimes (to prevent overbooking)
CREATE TABLE IF NOT EXISTS seat_reservations (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    showtime_id BIGINT NOT NULL,
    seat_id BIGINT NOT NULL,
    reservation_id BIGINT,
    status ENUM('available', 'reserved', 'locked') DEFAULT 'available',
    locked_until TIMESTAMP NULL, -- For temporary locks during booking process
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (showtime_id) REFERENCES showtimes(id) ON DELETE CASCADE,
    FOREIGN KEY (seat_id) REFERENCES seats(id) ON DELETE CASCADE,
    FOREIGN KEY (reservation_id) REFERENCES reservations(id) ON DELETE SET NULL,
    UNIQUE KEY unique_showtime_seat (showtime_id, seat_id),
    INDEX idx_seat_res_showtime (showtime_id),
    INDEX idx_seat_res_status (status),
    INDEX idx_seat_res_locked (locked_until)
);

-- Insert initial admin user. NOTE: The baked-in bcrypt hash below corresponds to the plaintext 'password'.
-- You can override/reset this at runtime by setting ADMIN_DEFAULT_PASSWORD in the environment.
INSERT IGNORE INTO users (username, password_hash, role, email, full_name) 
VALUES ('admin', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin', 'admin@moviereservation.com', 'System Administrator');

-- Insert sample genres
INSERT IGNORE INTO genres (name, description) VALUES
('Action', 'High-energy movies with exciting sequences'),
('Comedy', 'Movies designed to make audiences laugh'),
('Drama', 'Serious narrative fiction movies'),
('Horror', 'Movies intended to frighten and create suspense'),
('Romance', 'Movies focused on love stories'),
('Sci-Fi', 'Science fiction movies with futuristic concepts'),
('Thriller', 'Movies designed to keep audiences on edge'),
('Adventure', 'Movies featuring exciting journeys or quests'),
('Animation', 'Movies created using animation techniques'),
('Documentary', 'Non-fiction movies about real subjects');

-- Insert sample theaters
INSERT IGNORE INTO theaters (name, location, total_seats, rows_count, seats_per_row) VALUES
('Theater A - Main Hall', 'Ground Floor, Main Building', 100, 10, 10),
('Theater B - Premium', 'First Floor, East Wing', 80, 8, 10),
('Theater C - IMAX', 'Second Floor, West Wing', 120, 12, 10);

-- Dynamic bulk seat generation removed due to MySQL compatibility issues.
-- Seats will now be generated programmatically in Go during seeding.
-- (Original multi CROSS JOIN approach commented out.)
-- INSERT logic moved to SeedSampleData / generateSeatsForTheaters.