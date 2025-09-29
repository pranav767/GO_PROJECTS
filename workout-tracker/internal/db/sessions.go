package db

import (
	"database/sql"
	"time"
	"workout-tracker/model"
)

// Workout Session database operations

func CreateWorkoutSession(userID, workoutPlanID int64, scheduledDate time.Time, notes string) (*model.WorkoutSession, error) {
	query := `
		INSERT INTO workout_sessions (user_id, workout_plan_id, scheduled_date, notes)
		VALUES (?, ?, ?, ?)
	`
	result, err := GetDB().Exec(query, userID, workoutPlanID, scheduledDate, notes)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return GetWorkoutSessionByID(id)
}

func GetWorkoutSessionByID(id int64) (*model.WorkoutSession, error) {
	query := `
		SELECT ws.id, ws.user_id, ws.workout_plan_id, ws.scheduled_date, ws.started_at, 
		       ws.completed_at, ws.status, ws.notes, ws.total_duration_minutes, 
		       ws.created_at, ws.updated_at,
		       wp.name, wp.description, wp.is_active
		FROM workout_sessions ws
		JOIN workout_plans wp ON ws.workout_plan_id = wp.id
		WHERE ws.id = ?
	`
	var session model.WorkoutSession
	var plan model.WorkoutPlan
	var startedAt, completedAt sql.NullTime
	var notes sql.NullString

	err := GetDB().QueryRow(query, id).Scan(
		&session.ID,
		&session.UserID,
		&session.WorkoutPlanID,
		&session.ScheduledDate,
		&startedAt,
		&completedAt,
		&session.Status,
		&notes,
		&session.TotalDurationMinutes,
		&session.CreatedAt,
		&session.UpdatedAt,
		&plan.Name,
		&plan.Description,
		&plan.IsActive,
	)
	if err != nil {
		return nil, err
	}

	// Handle nullable fields
	if startedAt.Valid {
		session.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		session.CompletedAt = &completedAt.Time
	}
	if notes.Valid {
		session.Notes = notes.String
	}

	plan.ID = session.WorkoutPlanID
	plan.UserID = session.UserID
	session.WorkoutPlan = plan

	// Load session exercises
	exercises, err := GetSessionExercisesBySessionID(id)
	if err != nil {
		return nil, err
	}
	session.Exercises = exercises

	return &session, nil
}

func GetWorkoutSessionsByUserID(userID int64, status string, limit int) ([]model.WorkoutSession, error) {
	query := `
		SELECT ws.id, ws.user_id, ws.workout_plan_id, ws.scheduled_date, ws.started_at,
		       ws.completed_at, ws.status, ws.notes, ws.total_duration_minutes,
		       ws.created_at, ws.updated_at,
		       wp.name, wp.description, wp.is_active
		FROM workout_sessions ws
		JOIN workout_plans wp ON ws.workout_plan_id = wp.id
		WHERE ws.user_id = ?
	`
	args := []interface{}{userID}

	if status != "" {
		query += " AND ws.status = ?"
		args = append(args, status)
	}

	query += " ORDER BY ws.scheduled_date DESC"

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := GetDB().Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []model.WorkoutSession
	for rows.Next() {
		var session model.WorkoutSession
		var plan model.WorkoutPlan
		var startedAt, completedAt sql.NullTime
		var notes sql.NullString

		err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.WorkoutPlanID,
			&session.ScheduledDate,
			&startedAt,
			&completedAt,
			&session.Status,
			&notes,
			&session.TotalDurationMinutes,
			&session.CreatedAt,
			&session.UpdatedAt,
			&plan.Name,
			&plan.Description,
			&plan.IsActive,
		)
		if err != nil {
			return nil, err
		}

		// Handle nullable fields
		if startedAt.Valid {
			session.StartedAt = &startedAt.Time
		}
		if completedAt.Valid {
			session.CompletedAt = &completedAt.Time
		}
		if notes.Valid {
			session.Notes = notes.String
		}

		plan.ID = session.WorkoutPlanID
		plan.UserID = session.UserID
		session.WorkoutPlan = plan
		sessions = append(sessions, session)
	}
	return sessions, rows.Err()
}

func UpdateWorkoutSession(id int64, status, notes string, startedAt, completedAt *time.Time, totalDurationMinutes int) error {
	query := `
		UPDATE workout_sessions
		SET status = ?, notes = ?, total_duration_minutes = ?
	`
	args := []interface{}{status, notes, totalDurationMinutes}

	if startedAt != nil {
		query += ", started_at = ?"
		args = append(args, *startedAt)
	}

	if completedAt != nil {
		query += ", completed_at = ?"
		args = append(args, *completedAt)
	}

	query += " WHERE id = ?"
	args = append(args, id)

	_, err := GetDB().Exec(query, args...)
	return err
}

func DeleteWorkoutSession(id int64) error {
	query := `DELETE FROM workout_sessions WHERE id = ?`
	_, err := GetDB().Exec(query, id)
	return err
}

func GetWorkoutSessionByUserAndID(userID, sessionID int64) (*model.WorkoutSession, error) {
	query := `
		SELECT ws.id, ws.user_id, ws.workout_plan_id, ws.scheduled_date, ws.started_at,
		       ws.completed_at, ws.status, ws.notes, ws.total_duration_minutes,
		       ws.created_at, ws.updated_at,
		       wp.name, wp.description, wp.is_active
		FROM workout_sessions ws
		JOIN workout_plans wp ON ws.workout_plan_id = wp.id
		WHERE ws.id = ? AND ws.user_id = ?
	`
	var session model.WorkoutSession
	var plan model.WorkoutPlan
	var startedAt, completedAt sql.NullTime
	var notes sql.NullString

	err := GetDB().QueryRow(query, sessionID, userID).Scan(
		&session.ID,
		&session.UserID,
		&session.WorkoutPlanID,
		&session.ScheduledDate,
		&startedAt,
		&completedAt,
		&session.Status,
		&notes,
		&session.TotalDurationMinutes,
		&session.CreatedAt,
		&session.UpdatedAt,
		&plan.Name,
		&plan.Description,
		&plan.IsActive,
	)
	if err != nil {
		return nil, err
	}

	// Handle nullable fields
	if startedAt.Valid {
		session.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		session.CompletedAt = &completedAt.Time
	}
	if notes.Valid {
		session.Notes = notes.String
	}

	plan.ID = session.WorkoutPlanID
	plan.UserID = session.UserID
	session.WorkoutPlan = plan

	// Load session exercises
	exercises, err := GetSessionExercisesBySessionID(sessionID)
	if err != nil {
		return nil, err
	}
	session.Exercises = exercises

	return &session, nil
}

// Session Exercise operations

func CreateSessionExercise(sessionID, exerciseID int64, plannedSets, plannedReps int, plannedWeight *float64, plannedDurationSeconds *int) (*model.SessionExercise, error) {
	query := `
		INSERT INTO session_exercises (session_id, exercise_id, planned_sets, planned_reps, planned_weight, planned_duration_seconds)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	result, err := GetDB().Exec(query, sessionID, exerciseID, plannedSets, plannedReps, plannedWeight, plannedDurationSeconds)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return GetSessionExerciseByID(id)
}

func GetSessionExerciseByID(id int64) (*model.SessionExercise, error) {
	query := `
		SELECT se.id, se.session_id, se.exercise_id, se.planned_sets, se.planned_reps,
		       se.planned_weight, se.planned_duration_seconds, se.completed_sets, se.completed_reps,
		       se.actual_weight, se.actual_duration_seconds, se.notes, se.completed, se.created_at,
		       e.name, e.description, e.category, e.muscle_group, e.created_at
		FROM session_exercises se
		JOIN exercises e ON se.exercise_id = e.id
		WHERE se.id = ?
	`
	var sessionExercise model.SessionExercise
	var exercise model.Exercise
	var plannedWeight, actualWeight sql.NullFloat64
	var plannedDuration, actualDuration sql.NullInt64
	var notes sql.NullString

	err := GetDB().QueryRow(query, id).Scan(
		&sessionExercise.ID,
		&sessionExercise.SessionID,
		&sessionExercise.ExerciseID,
		&sessionExercise.PlannedSets,
		&sessionExercise.PlannedReps,
		&plannedWeight,
		&plannedDuration,
		&sessionExercise.CompletedSets,
		&sessionExercise.CompletedReps,
		&actualWeight,
		&actualDuration,
		&notes,
		&sessionExercise.Completed,
		&sessionExercise.CreatedAt,
		&exercise.Name,
		&exercise.Description,
		&exercise.Category,
		&exercise.MuscleGroup,
		&exercise.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Handle nullable fields
	if plannedWeight.Valid {
		sessionExercise.PlannedWeight = &plannedWeight.Float64
	}
	if plannedDuration.Valid {
		duration := int(plannedDuration.Int64)
		sessionExercise.PlannedDurationSeconds = &duration
	}
	if actualWeight.Valid {
		sessionExercise.ActualWeight = &actualWeight.Float64
	}
	if actualDuration.Valid {
		duration := int(actualDuration.Int64)
		sessionExercise.ActualDurationSeconds = &duration
	}
	if notes.Valid {
		sessionExercise.Notes = notes.String
	}

	exercise.ID = sessionExercise.ExerciseID
	sessionExercise.Exercise = exercise

	return &sessionExercise, nil
}

func GetSessionExercisesBySessionID(sessionID int64) ([]model.SessionExercise, error) {
	query := `
		SELECT se.id, se.session_id, se.exercise_id, se.planned_sets, se.planned_reps,
		       se.planned_weight, se.planned_duration_seconds, se.completed_sets, se.completed_reps,
		       se.actual_weight, se.actual_duration_seconds, se.notes, se.completed, se.created_at,
		       e.name, e.description, e.category, e.muscle_group, e.created_at
		FROM session_exercises se
		JOIN exercises e ON se.exercise_id = e.id
		WHERE se.session_id = ?
		ORDER BY se.id
	`
	rows, err := GetDB().Query(query, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessionExercises []model.SessionExercise
	for rows.Next() {
		var sessionExercise model.SessionExercise
		var exercise model.Exercise
		var plannedWeight, actualWeight sql.NullFloat64
		var plannedDuration, actualDuration sql.NullInt64
		var notes sql.NullString

		err := rows.Scan(
			&sessionExercise.ID,
			&sessionExercise.SessionID,
			&sessionExercise.ExerciseID,
			&sessionExercise.PlannedSets,
			&sessionExercise.PlannedReps,
			&plannedWeight,
			&plannedDuration,
			&sessionExercise.CompletedSets,
			&sessionExercise.CompletedReps,
			&actualWeight,
			&actualDuration,
			&notes,
			&sessionExercise.Completed,
			&sessionExercise.CreatedAt,
			&exercise.Name,
			&exercise.Description,
			&exercise.Category,
			&exercise.MuscleGroup,
			&exercise.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Handle nullable fields
		if plannedWeight.Valid {
			sessionExercise.PlannedWeight = &plannedWeight.Float64
		}
		if plannedDuration.Valid {
			duration := int(plannedDuration.Int64)
			sessionExercise.PlannedDurationSeconds = &duration
		}
		if actualWeight.Valid {
			sessionExercise.ActualWeight = &actualWeight.Float64
		}
		if actualDuration.Valid {
			duration := int(actualDuration.Int64)
			sessionExercise.ActualDurationSeconds = &duration
		}
		if notes.Valid {
			sessionExercise.Notes = notes.String
		}

		exercise.ID = sessionExercise.ExerciseID
		sessionExercise.Exercise = exercise
		sessionExercises = append(sessionExercises, sessionExercise)
	}
	return sessionExercises, rows.Err()
}

func UpdateSessionExercise(id int64, completedSets, completedReps int, actualWeight *float64, actualDurationSeconds *int, notes string, completed bool) error {
	query := `
		UPDATE session_exercises
		SET completed_sets = ?, completed_reps = ?, actual_weight = ?, actual_duration_seconds = ?, notes = ?, completed = ?
		WHERE id = ?
	`
	_, err := GetDB().Exec(query, completedSets, completedReps, actualWeight, actualDurationSeconds, notes, completed, id)
	return err
}

func DeleteSessionExercise(id int64) error {
	query := `DELETE FROM session_exercises WHERE id = ?`
	_, err := GetDB().Exec(query, id)
	return err
}

func CreateSessionExercisesFromWorkoutPlan(sessionID, workoutPlanID int64) error {
	// Get workout exercises from the plan
	workoutExercises, err := GetWorkoutExercisesByPlanID(workoutPlanID)
	if err != nil {
		return err
	}

	// Create session exercises based on workout plan
	for _, we := range workoutExercises {
		_, err = CreateSessionExercise(
			sessionID,
			we.ExerciseID,
			we.Sets,
			we.Reps,
			we.Weight,
			we.DurationSeconds,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
