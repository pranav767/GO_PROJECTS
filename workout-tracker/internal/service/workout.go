package service

import (
	"errors"
	"time"
	"workout-tracker/internal/db"
	"workout-tracker/model"
)

// Exercise Services

func GetExercises(category, muscleGroup string) ([]model.Exercise, error) {
	if category != "" {
		return db.GetExercisesByCategory(category)
	}
	if muscleGroup != "" {
		return db.GetExercisesByMuscleGroup(muscleGroup)
	}
	return db.GetAllExercises()
}

func GetExerciseByID(id int64) (*model.Exercise, error) {
	return db.GetExerciseByID(id)
}

// Workout Plan Services

func CreateWorkoutPlan(userID int64, req model.CreateWorkoutPlanRequest) (*model.WorkoutPlan, error) {
	// Create the workout plan
	plan, err := db.CreateWorkoutPlan(userID, req.Name, req.Description)
	if err != nil {
		return nil, err
	}

	// Add exercises to the plan
	for _, exerciseReq := range req.Exercises {
		// Verify exercise exists
		_, err := db.GetExerciseByID(exerciseReq.ExerciseID)
		if err != nil {
			// Clean up by deleting the plan
			db.DeleteWorkoutPlan(plan.ID)
			return nil, errors.New("exercise not found")
		}

		_, err = db.CreateWorkoutExercise(
			plan.ID,
			exerciseReq.ExerciseID,
			exerciseReq.Sets,
			exerciseReq.Reps,
			exerciseReq.Weight,
			exerciseReq.DurationSeconds,
			exerciseReq.RestSeconds,
			exerciseReq.OrderInWorkout,
			exerciseReq.Notes,
		)
		if err != nil {
			// Clean up by deleting the plan
			db.DeleteWorkoutPlan(plan.ID)
			return nil, err
		}
	}

	// Return the plan with exercises loaded
	return db.GetWorkoutPlanByID(plan.ID)
}

func GetUserWorkoutPlans(userID int64, activeOnly bool) ([]model.WorkoutPlan, error) {
	return db.GetWorkoutPlansByUserID(userID, activeOnly)
}

func GetUserWorkoutPlan(userID, planID int64) (*model.WorkoutPlan, error) {
	plan, err := db.GetWorkoutPlanByUserAndID(userID, planID)
	if err != nil {
		return nil, errors.New("workout plan not found or access denied")
	}
	return plan, nil
}

func UpdateWorkoutPlan(userID, planID int64, req model.UpdateWorkoutPlanRequest) (*model.WorkoutPlan, error) {
	// Verify ownership
	_, err := db.GetWorkoutPlanByUserAndID(userID, planID)
	if err != nil {
		return nil, errors.New("workout plan not found or access denied")
	}

	// Update the plan
	err = db.UpdateWorkoutPlan(planID, req.Name, req.Description, req.IsActive)
	if err != nil {
		return nil, err
	}

	return db.GetWorkoutPlanByID(planID)
}

func DeleteWorkoutPlan(userID, planID int64) error {
	// Verify ownership
	_, err := db.GetWorkoutPlanByUserAndID(userID, planID)
	if err != nil {
		return errors.New("workout plan not found or access denied")
	}

	return db.DeleteWorkoutPlan(planID)
}

// Workout Session Services

func ScheduleWorkout(userID int64, req model.ScheduleWorkoutRequest) (*model.WorkoutSession, error) {
	// Verify the workout plan exists and belongs to the user
	_, err := db.GetWorkoutPlanByUserAndID(userID, req.WorkoutPlanID)
	if err != nil {
		return nil, errors.New("workout plan not found or access denied")
	}

	// Create the workout session
	session, err := db.CreateWorkoutSession(userID, req.WorkoutPlanID, req.ScheduledDate, req.Notes)
	if err != nil {
		return nil, err
	}

	// Create session exercises from the workout plan
	err = db.CreateSessionExercisesFromWorkoutPlan(session.ID, req.WorkoutPlanID)
	if err != nil {
		// Clean up by deleting the session
		db.DeleteWorkoutSession(session.ID)
		return nil, err
	}

	return db.GetWorkoutSessionByID(session.ID)
}

func GetUserWorkoutSessions(userID int64, status string, limit int) ([]model.WorkoutSession, error) {
	return db.GetWorkoutSessionsByUserID(userID, status, limit)
}

func GetUserWorkoutSession(userID, sessionID int64) (*model.WorkoutSession, error) {
	session, err := db.GetWorkoutSessionByUserAndID(userID, sessionID)
	if err != nil {
		return nil, errors.New("workout session not found or access denied")
	}
	return session, nil
}

func StartWorkoutSession(userID, sessionID int64) (*model.WorkoutSession, error) {
	// Verify ownership
	session, err := db.GetWorkoutSessionByUserAndID(userID, sessionID)
	if err != nil {
		return nil, errors.New("workout session not found or access denied")
	}

	// Check if session is in correct state
	if session.Status != "scheduled" {
		return nil, errors.New("workout session cannot be started")
	}

	// Update session to in_progress
	now := time.Now()
	err = db.UpdateWorkoutSession(sessionID, "in_progress", session.Notes, &now, nil, 0)
	if err != nil {
		return nil, err
	}

	return db.GetWorkoutSessionByID(sessionID)
}

func CompleteWorkoutSession(userID, sessionID int64, req model.UpdateWorkoutSessionRequest) (*model.WorkoutSession, error) {
	// Verify ownership
	session, err := db.GetWorkoutSessionByUserAndID(userID, sessionID)
	if err != nil {
		return nil, errors.New("workout session not found or access denied")
	}

	// Update session exercises if provided
	for _, exerciseReq := range req.Exercises {
		err = db.UpdateSessionExercise(
			exerciseReq.ID,
			exerciseReq.CompletedSets,
			exerciseReq.CompletedReps,
			exerciseReq.ActualWeight,
			exerciseReq.ActualDurationSeconds,
			exerciseReq.Notes,
			exerciseReq.Completed,
		)
		if err != nil {
			return nil, err
		}
	}

	// Update session status
	var completedAt *time.Time
	if req.Status == "completed" {
		now := time.Now()
		completedAt = &now
	}

	err = db.UpdateWorkoutSession(
		sessionID,
		req.Status,
		req.Notes,
		session.StartedAt,
		completedAt,
		req.TotalDurationMinutes,
	)
	if err != nil {
		return nil, err
	}

	return db.GetWorkoutSessionByID(sessionID)
}

func UpdateWorkoutSession(userID, sessionID int64, req model.UpdateWorkoutSessionRequest) (*model.WorkoutSession, error) {
	// Verify ownership
	session, err := db.GetWorkoutSessionByUserAndID(userID, sessionID)
	if err != nil {
		return nil, errors.New("workout session not found or access denied")
	}

	// Update session exercises if provided
	for _, exerciseReq := range req.Exercises {
		err = db.UpdateSessionExercise(
			exerciseReq.ID,
			exerciseReq.CompletedSets,
			exerciseReq.CompletedReps,
			exerciseReq.ActualWeight,
			exerciseReq.ActualDurationSeconds,
			exerciseReq.Notes,
			exerciseReq.Completed,
		)
		if err != nil {
			return nil, err
		}
	}

	// Update session
	var completedAt *time.Time
	if req.Status == "completed" && session.CompletedAt == nil {
		now := time.Now()
		completedAt = &now
	} else {
		completedAt = session.CompletedAt
	}

	err = db.UpdateWorkoutSession(
		sessionID,
		req.Status,
		req.Notes,
		session.StartedAt,
		completedAt,
		req.TotalDurationMinutes,
	)
	if err != nil {
		return nil, err
	}

	return db.GetWorkoutSessionByID(sessionID)
}

func DeleteWorkoutSession(userID, sessionID int64) error {
	// Verify ownership
	_, err := db.GetWorkoutSessionByUserAndID(userID, sessionID)
	if err != nil {
		return errors.New("workout session not found or access denied")
	}

	return db.DeleteWorkoutSession(sessionID)
}

func GetUpcomingWorkouts(userID int64, days int) ([]model.WorkoutSession, error) {
	// Get scheduled sessions for the next X days
	sessions, err := db.GetWorkoutSessionsByUserID(userID, "scheduled", 0)
	if err != nil {
		return nil, err
	}

	// Filter for upcoming sessions within the specified days
	var upcomingSessions []model.WorkoutSession
	now := time.Now()
	cutoff := now.AddDate(0, 0, days)

	for _, session := range sessions {
		if session.ScheduledDate.After(now) && session.ScheduledDate.Before(cutoff) {
			upcomingSessions = append(upcomingSessions, session)
		}
	}

	return upcomingSessions, nil
}
