package db

import (
	"workout-tracker/model"
)

// Workout Plan database operations

func CreateWorkoutPlan(userID int64, name, description string) (*model.WorkoutPlan, error) {
	query := `
		INSERT INTO workout_plans (user_id, name, description)
		VALUES (?, ?, ?)
	`
	result, err := GetDB().Exec(query, userID, name, description)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return GetWorkoutPlanByID(id)
}

func GetWorkoutPlanByID(id int64) (*model.WorkoutPlan, error) {
	query := `
		SELECT id, user_id, name, description, is_active, created_at, updated_at
		FROM workout_plans
		WHERE id = ?
	`
	var plan model.WorkoutPlan
	err := GetDB().QueryRow(query, id).Scan(
		&plan.ID,
		&plan.UserID,
		&plan.Name,
		&plan.Description,
		&plan.IsActive,
		&plan.CreatedAt,
		&plan.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Load exercises for this workout plan
	exercises, err := GetWorkoutExercisesByPlanID(id)
	if err != nil {
		return nil, err
	}
	plan.Exercises = exercises

	return &plan, nil
}

func GetWorkoutPlansByUserID(userID int64, activeOnly bool) ([]model.WorkoutPlan, error) {
	query := `
		SELECT id, user_id, name, description, is_active, created_at, updated_at
		FROM workout_plans
		WHERE user_id = ?
	`
	args := []interface{}{userID}

	if activeOnly {
		query += " AND is_active = TRUE"
	}

	query += " ORDER BY created_at DESC"

	rows, err := GetDB().Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []model.WorkoutPlan
	for rows.Next() {
		var plan model.WorkoutPlan
		err := rows.Scan(
			&plan.ID,
			&plan.UserID,
			&plan.Name,
			&plan.Description,
			&plan.IsActive,
			&plan.CreatedAt,
			&plan.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}
	return plans, rows.Err()
}

func UpdateWorkoutPlan(id int64, name, description string, isActive *bool) error {
	query := `
		UPDATE workout_plans
		SET name = ?, description = ?
	`
	args := []interface{}{name, description}

	if isActive != nil {
		query += ", is_active = ?"
		args = append(args, *isActive)
	}

	query += " WHERE id = ?"
	args = append(args, id)

	_, err := GetDB().Exec(query, args...)
	return err
}

func DeleteWorkoutPlan(id int64) error {
	query := `DELETE FROM workout_plans WHERE id = ?`
	_, err := GetDB().Exec(query, id)
	return err
}

func GetWorkoutPlanByUserAndID(userID, planID int64) (*model.WorkoutPlan, error) {
	query := `
		SELECT id, user_id, name, description, is_active, created_at, updated_at
		FROM workout_plans
		WHERE id = ? AND user_id = ?
	`
	var plan model.WorkoutPlan
	err := GetDB().QueryRow(query, planID, userID).Scan(
		&plan.ID,
		&plan.UserID,
		&plan.Name,
		&plan.Description,
		&plan.IsActive,
		&plan.CreatedAt,
		&plan.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Load exercises for this workout plan
	exercises, err := GetWorkoutExercisesByPlanID(planID)
	if err != nil {
		return nil, err
	}
	plan.Exercises = exercises

	return &plan, nil
}

// Workout Exercise operations

func CreateWorkoutExercise(planID, exerciseID int64, sets, reps int, weight *float64, durationSeconds *int, restSeconds, orderInWorkout int, notes string) (*model.WorkoutExercise, error) {
	query := `
		INSERT INTO workout_exercises (workout_plan_id, exercise_id, sets, reps, weight, duration_seconds, rest_seconds, order_in_workout, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := GetDB().Exec(query, planID, exerciseID, sets, reps, weight, durationSeconds, restSeconds, orderInWorkout, notes)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return GetWorkoutExerciseByID(id)
}

func GetWorkoutExerciseByID(id int64) (*model.WorkoutExercise, error) {
	query := `
		SELECT we.id, we.workout_plan_id, we.exercise_id, we.sets, we.reps, we.weight, 
		       we.duration_seconds, we.rest_seconds, we.order_in_workout, we.notes, we.created_at,
		       e.name, e.description, e.category, e.muscle_group, e.created_at
		FROM workout_exercises we
		JOIN exercises e ON we.exercise_id = e.id
		WHERE we.id = ?
	`
	var workoutExercise model.WorkoutExercise
	var exercise model.Exercise

	err := GetDB().QueryRow(query, id).Scan(
		&workoutExercise.ID,
		&workoutExercise.WorkoutPlanID,
		&workoutExercise.ExerciseID,
		&workoutExercise.Sets,
		&workoutExercise.Reps,
		&workoutExercise.Weight,
		&workoutExercise.DurationSeconds,
		&workoutExercise.RestSeconds,
		&workoutExercise.OrderInWorkout,
		&workoutExercise.Notes,
		&workoutExercise.CreatedAt,
		&exercise.Name,
		&exercise.Description,
		&exercise.Category,
		&exercise.MuscleGroup,
		&exercise.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	exercise.ID = workoutExercise.ExerciseID
	workoutExercise.Exercise = exercise

	return &workoutExercise, nil
}

func GetWorkoutExercisesByPlanID(planID int64) ([]model.WorkoutExercise, error) {
	query := `
		SELECT we.id, we.workout_plan_id, we.exercise_id, we.sets, we.reps, we.weight,
		       we.duration_seconds, we.rest_seconds, we.order_in_workout, we.notes, we.created_at,
		       e.name, e.description, e.category, e.muscle_group, e.created_at
		FROM workout_exercises we
		JOIN exercises e ON we.exercise_id = e.id
		WHERE we.workout_plan_id = ?
		ORDER BY we.order_in_workout
	`
	rows, err := GetDB().Query(query, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workoutExercises []model.WorkoutExercise
	for rows.Next() {
		var workoutExercise model.WorkoutExercise
		var exercise model.Exercise

		err := rows.Scan(
			&workoutExercise.ID,
			&workoutExercise.WorkoutPlanID,
			&workoutExercise.ExerciseID,
			&workoutExercise.Sets,
			&workoutExercise.Reps,
			&workoutExercise.Weight,
			&workoutExercise.DurationSeconds,
			&workoutExercise.RestSeconds,
			&workoutExercise.OrderInWorkout,
			&workoutExercise.Notes,
			&workoutExercise.CreatedAt,
			&exercise.Name,
			&exercise.Description,
			&exercise.Category,
			&exercise.MuscleGroup,
			&exercise.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		exercise.ID = workoutExercise.ExerciseID
		workoutExercise.Exercise = exercise
		workoutExercises = append(workoutExercises, workoutExercise)
	}
	return workoutExercises, rows.Err()
}

func DeleteWorkoutExercise(id int64) error {
	query := `DELETE FROM workout_exercises WHERE id = ?`
	_, err := GetDB().Exec(query, id)
	return err
}

func DeleteWorkoutExercisesByPlanID(planID int64) error {
	query := `DELETE FROM workout_exercises WHERE workout_plan_id = ?`
	_, err := GetDB().Exec(query, planID)
	return err
}
