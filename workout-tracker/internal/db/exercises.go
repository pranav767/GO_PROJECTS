package db

import (
	"workout-tracker/model"
)

// Exercise database operations

func CreateExercise(name, description, category, muscleGroup string) (*model.Exercise, error) {
	query := `
		INSERT INTO exercises (name, description, category, muscle_group)
		VALUES (?, ?, ?, ?)
	`
	result, err := GetDB().Exec(query, name, description, category, muscleGroup)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return GetExerciseByID(id)
}

func GetExerciseByID(id int64) (*model.Exercise, error) {
	query := `
		SELECT id, name, description, category, muscle_group, created_at
		FROM exercises
		WHERE id = ?
	`
	var exercise model.Exercise
	err := GetDB().QueryRow(query, id).Scan(
		&exercise.ID,
		&exercise.Name,
		&exercise.Description,
		&exercise.Category,
		&exercise.MuscleGroup,
		&exercise.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &exercise, nil
}

func GetAllExercises() ([]model.Exercise, error) {
	query := `
		SELECT id, name, description, category, muscle_group, created_at
		FROM exercises
		ORDER BY category, name
	`
	rows, err := GetDB().Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exercises []model.Exercise
	for rows.Next() {
		var exercise model.Exercise
		err := rows.Scan(
			&exercise.ID,
			&exercise.Name,
			&exercise.Description,
			&exercise.Category,
			&exercise.MuscleGroup,
			&exercise.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		exercises = append(exercises, exercise)
	}
	return exercises, rows.Err()
}

func GetExercisesByCategory(category string) ([]model.Exercise, error) {
	query := `
		SELECT id, name, description, category, muscle_group, created_at
		FROM exercises
		WHERE category = ?
		ORDER BY name
	`
	rows, err := GetDB().Query(query, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exercises []model.Exercise
	for rows.Next() {
		var exercise model.Exercise
		err := rows.Scan(
			&exercise.ID,
			&exercise.Name,
			&exercise.Description,
			&exercise.Category,
			&exercise.MuscleGroup,
			&exercise.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		exercises = append(exercises, exercise)
	}
	return exercises, rows.Err()
}

func GetExercisesByMuscleGroup(muscleGroup string) ([]model.Exercise, error) {
	query := `
		SELECT id, name, description, category, muscle_group, created_at
		FROM exercises
		WHERE muscle_group = ?
		ORDER BY name
	`
	rows, err := GetDB().Query(query, muscleGroup)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exercises []model.Exercise
	for rows.Next() {
		var exercise model.Exercise
		err := rows.Scan(
			&exercise.ID,
			&exercise.Name,
			&exercise.Description,
			&exercise.Category,
			&exercise.MuscleGroup,
			&exercise.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		exercises = append(exercises, exercise)
	}
	return exercises, rows.Err()
}

func UpdateExercise(id int64, name, description, category, muscleGroup string) error {
	query := `
		UPDATE exercises
		SET name = ?, description = ?, category = ?, muscle_group = ?
		WHERE id = ?
	`
	_, err := GetDB().Exec(query, name, description, category, muscleGroup, id)
	return err
}

func DeleteExercise(id int64) error {
	query := `DELETE FROM exercises WHERE id = ?`
	_, err := GetDB().Exec(query, id)
	return err
}
