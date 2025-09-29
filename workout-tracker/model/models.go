package model

import (
	"time"
)

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Exercise struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	MuscleGroup string    `json:"muscle_group"`
	CreatedAt   time.Time `json:"created_at"`
}

type WorkoutPlan struct {
	ID          int64             `json:"id"`
	UserID      int64             `json:"user_id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	IsActive    bool              `json:"is_active"`
	Exercises   []WorkoutExercise `json:"exercises,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type WorkoutExercise struct {
	ID              int64     `json:"id"`
	WorkoutPlanID   int64     `json:"workout_plan_id"`
	ExerciseID      int64     `json:"exercise_id"`
	Exercise        Exercise  `json:"exercise,omitempty"`
	Sets            int       `json:"sets"`
	Reps            int       `json:"reps"`
	Weight          *float64  `json:"weight,omitempty"`
	DurationSeconds *int      `json:"duration_seconds,omitempty"`
	RestSeconds     int       `json:"rest_seconds"`
	OrderInWorkout  int       `json:"order_in_workout"`
	Notes           string    `json:"notes"`
	CreatedAt       time.Time `json:"created_at"`
}

type WorkoutSession struct {
	ID                   int64             `json:"id"`
	UserID               int64             `json:"user_id"`
	WorkoutPlanID        int64             `json:"workout_plan_id"`
	WorkoutPlan          WorkoutPlan       `json:"workout_plan,omitempty"`
	ScheduledDate        time.Time         `json:"scheduled_date"`
	StartedAt            *time.Time        `json:"started_at,omitempty"`
	CompletedAt          *time.Time        `json:"completed_at,omitempty"`
	Status               string            `json:"status"`
	Notes                string            `json:"notes"`
	TotalDurationMinutes int               `json:"total_duration_minutes"`
	Exercises            []SessionExercise `json:"exercises,omitempty"`
	CreatedAt            time.Time         `json:"created_at"`
	UpdatedAt            time.Time         `json:"updated_at"`
}

type SessionExercise struct {
	ID                     int64     `json:"id"`
	SessionID              int64     `json:"session_id"`
	ExerciseID             int64     `json:"exercise_id"`
	Exercise               Exercise  `json:"exercise,omitempty"`
	PlannedSets            int       `json:"planned_sets"`
	PlannedReps            int       `json:"planned_reps"`
	PlannedWeight          *float64  `json:"planned_weight,omitempty"`
	PlannedDurationSeconds *int      `json:"planned_duration_seconds,omitempty"`
	CompletedSets          int       `json:"completed_sets"`
	CompletedReps          int       `json:"completed_reps"`
	ActualWeight           *float64  `json:"actual_weight,omitempty"`
	ActualDurationSeconds  *int      `json:"actual_duration_seconds,omitempty"`
	Notes                  string    `json:"notes"`
	Completed              bool      `json:"completed"`
	CreatedAt              time.Time `json:"created_at"`
}

// Request/Response DTOs
type CreateWorkoutPlanRequest struct {
	Name        string                         `json:"name" binding:"required"`
	Description string                         `json:"description"`
	Exercises   []CreateWorkoutExerciseRequest `json:"exercises" binding:"required"`
}

type CreateWorkoutExerciseRequest struct {
	ExerciseID      int64    `json:"exercise_id" binding:"required"`
	Sets            int      `json:"sets" binding:"required,min=1"`
	Reps            int      `json:"reps" binding:"required,min=1"`
	Weight          *float64 `json:"weight,omitempty"`
	DurationSeconds *int     `json:"duration_seconds,omitempty"`
	RestSeconds     int      `json:"rest_seconds"`
	OrderInWorkout  int      `json:"order_in_workout"`
	Notes           string   `json:"notes"`
}

type UpdateWorkoutPlanRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    *bool  `json:"is_active,omitempty"`
}

type ScheduleWorkoutRequest struct {
	WorkoutPlanID int64     `json:"workout_plan_id" binding:"required"`
	ScheduledDate time.Time `json:"scheduled_date" binding:"required"`
	Notes         string    `json:"notes"`
}

type UpdateWorkoutSessionRequest struct {
	Status               string                         `json:"status"`
	Notes                string                         `json:"notes"`
	TotalDurationMinutes int                            `json:"total_duration_minutes"`
	Exercises            []UpdateSessionExerciseRequest `json:"exercises"`
}

type UpdateSessionExerciseRequest struct {
	ID                    int64    `json:"id" binding:"required"`
	CompletedSets         int      `json:"completed_sets"`
	CompletedReps         int      `json:"completed_reps"`
	ActualWeight          *float64 `json:"actual_weight,omitempty"`
	ActualDurationSeconds *int     `json:"actual_duration_seconds,omitempty"`
	Notes                 string   `json:"notes"`
	Completed             bool     `json:"completed"`
}

type WorkoutReport struct {
	UserID             int64                    `json:"user_id"`
	TotalWorkouts      int                      `json:"total_workouts"`
	CompletedWorkouts  int                      `json:"completed_workouts"`
	TotalTimeMinutes   int                      `json:"total_time_minutes"`
	WorkoutsByCategory map[string]int           `json:"workouts_by_category"`
	ExerciseProgress   []ExerciseProgressReport `json:"exercise_progress"`
	RecentSessions     []WorkoutSession         `json:"recent_sessions"`
}

type ExerciseProgressReport struct {
	ExerciseID          int64     `json:"exercise_id"`
	ExerciseName        string    `json:"exercise_name"`
	TotalSessions       int       `json:"total_sessions"`
	BestWeight          *float64  `json:"best_weight,omitempty"`
	BestReps            int       `json:"best_reps"`
	BestDurationSeconds *int      `json:"best_duration_seconds,omitempty"`
	LastPerformed       time.Time `json:"last_performed"`
}
