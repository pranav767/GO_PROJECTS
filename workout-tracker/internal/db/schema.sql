CREATE TABLE IF NOT EXISTS users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Exercises table to store predefined exercises
CREATE TABLE IF NOT EXISTS exercises (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category ENUM('cardio', 'strength', 'flexibility', 'balance', 'sports') NOT NULL,
    muscle_group ENUM('chest', 'back', 'shoulders', 'arms', 'legs', 'core', 'full_body', 'cardio') NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_category (category),
    INDEX idx_muscle_group (muscle_group)
);

-- Workout plans table
CREATE TABLE IF NOT EXISTS workout_plans (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_active (user_id, is_active)
);

-- Junction table for exercises in workout plans
CREATE TABLE IF NOT EXISTS workout_exercises (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    workout_plan_id BIGINT NOT NULL,
    exercise_id BIGINT NOT NULL,
    sets INT NOT NULL DEFAULT 1,
    reps INT NOT NULL DEFAULT 1,
    weight DECIMAL(5,2) DEFAULT NULL,
    duration_seconds INT DEFAULT NULL,
    rest_seconds INT DEFAULT 60,
    order_in_workout INT NOT NULL DEFAULT 1,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (workout_plan_id) REFERENCES workout_plans(id) ON DELETE CASCADE,
    FOREIGN KEY (exercise_id) REFERENCES exercises(id) ON DELETE CASCADE,
    INDEX idx_workout_plan (workout_plan_id),
    INDEX idx_exercise (exercise_id)
);

-- Workout sessions table for tracking actual workouts
CREATE TABLE IF NOT EXISTS workout_sessions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    workout_plan_id BIGINT NOT NULL,
    scheduled_date DATETIME NOT NULL,
    started_at DATETIME,
    completed_at DATETIME,
    status ENUM('scheduled', 'in_progress', 'completed', 'cancelled') DEFAULT 'scheduled',
    notes TEXT,
    total_duration_minutes INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (workout_plan_id) REFERENCES workout_plans(id) ON DELETE CASCADE,
    INDEX idx_user_date (user_id, scheduled_date),
    INDEX idx_status (status),
    INDEX idx_workout_plan (workout_plan_id)
);

-- Exercise performance tracking in sessions
CREATE TABLE IF NOT EXISTS session_exercises (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    session_id BIGINT NOT NULL,
    exercise_id BIGINT NOT NULL,
    planned_sets INT NOT NULL,
    planned_reps INT NOT NULL,
    planned_weight DECIMAL(5,2),
    planned_duration_seconds INT,
    completed_sets INT DEFAULT 0,
    completed_reps INT DEFAULT 0,
    actual_weight DECIMAL(5,2),
    actual_duration_seconds INT,
    notes TEXT,
    completed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES workout_sessions(id) ON DELETE CASCADE,
    FOREIGN KEY (exercise_id) REFERENCES exercises(id) ON DELETE CASCADE,
    INDEX idx_session (session_id),
    INDEX idx_exercise (exercise_id)
);