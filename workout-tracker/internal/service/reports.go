package service

import (
	"time"
	"workout-tracker/internal/db"
	"workout-tracker/model"
)

// Report Services

func GenerateWorkoutReport(userID int64, startDate, endDate time.Time) (*model.WorkoutReport, error) {
	// Get all completed sessions for the user in the date range
	sessions, err := db.GetWorkoutSessionsByUserID(userID, "completed", 0)
	if err != nil {
		return nil, err
	}

	// Filter sessions by date range
	var filteredSessions []model.WorkoutSession
	for _, session := range sessions {
		if session.CompletedAt != nil &&
			session.CompletedAt.After(startDate) &&
			session.CompletedAt.Before(endDate) {
			filteredSessions = append(filteredSessions, session)
		}
	}

	// Calculate basic stats
	totalWorkouts := len(filteredSessions)
	totalTimeMinutes := 0
	workoutsByCategory := make(map[string]int)

	for _, session := range filteredSessions {
		totalTimeMinutes += session.TotalDurationMinutes

		// Load session exercises to categorize
		sessionExercises, err := db.GetSessionExercisesBySessionID(session.ID)
		if err != nil {
			continue
		}

		categories := make(map[string]bool)
		for _, se := range sessionExercises {
			if !categories[se.Exercise.Category] {
				categories[se.Exercise.Category] = true
				workoutsByCategory[se.Exercise.Category]++
			}
		}
	}

	// Get exercise progress
	exerciseProgress, err := generateExerciseProgress(userID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Get recent sessions (last 10)
	recentSessions, err := db.GetWorkoutSessionsByUserID(userID, "", 10)
	if err != nil {
		return nil, err
	}

	report := &model.WorkoutReport{
		UserID:             userID,
		TotalWorkouts:      totalWorkouts,
		CompletedWorkouts:  totalWorkouts,
		TotalTimeMinutes:   totalTimeMinutes,
		WorkoutsByCategory: workoutsByCategory,
		ExerciseProgress:   exerciseProgress,
		RecentSessions:     recentSessions,
	}

	return report, nil
}

func generateExerciseProgress(userID int64, startDate, endDate time.Time) ([]model.ExerciseProgressReport, error) {
	// Get all completed sessions for the user
	sessions, err := db.GetWorkoutSessionsByUserID(userID, "completed", 0)
	if err != nil {
		return nil, err
	}

	// Track exercise performance
	exerciseStats := make(map[int64]*model.ExerciseProgressReport)

	for _, session := range sessions {
		if session.CompletedAt == nil ||
			session.CompletedAt.Before(startDate) ||
			session.CompletedAt.After(endDate) {
			continue
		}

		sessionExercises, err := db.GetSessionExercisesBySessionID(session.ID)
		if err != nil {
			continue
		}

		for _, se := range sessionExercises {
			if !se.Completed {
				continue
			}

			stats, exists := exerciseStats[se.ExerciseID]
			if !exists {
				stats = &model.ExerciseProgressReport{
					ExerciseID:   se.ExerciseID,
					ExerciseName: se.Exercise.Name,
				}
				exerciseStats[se.ExerciseID] = stats
			}

			stats.TotalSessions++

			// Update best records
			if se.ActualWeight != nil && (stats.BestWeight == nil || *se.ActualWeight > *stats.BestWeight) {
				stats.BestWeight = se.ActualWeight
			}

			if se.CompletedReps > stats.BestReps {
				stats.BestReps = se.CompletedReps
			}

			if se.ActualDurationSeconds != nil && (stats.BestDurationSeconds == nil || *se.ActualDurationSeconds > *stats.BestDurationSeconds) {
				stats.BestDurationSeconds = se.ActualDurationSeconds
			}

			// Update last performed
			if session.CompletedAt.After(stats.LastPerformed) {
				stats.LastPerformed = *session.CompletedAt
			}
		}
	}

	// Convert map to slice
	var progressReports []model.ExerciseProgressReport
	for _, stats := range exerciseStats {
		progressReports = append(progressReports, *stats)
	}

	return progressReports, nil
}

func GetPersonalRecords(userID int64) ([]model.ExerciseProgressReport, error) {
	// Get all completed sessions for the user
	sessions, err := db.GetWorkoutSessionsByUserID(userID, "completed", 0)
	if err != nil {
		return nil, err
	}

	// Track personal records for each exercise
	exerciseRecords := make(map[int64]*model.ExerciseProgressReport)

	for _, session := range sessions {
		if session.CompletedAt == nil {
			continue
		}

		sessionExercises, err := db.GetSessionExercisesBySessionID(session.ID)
		if err != nil {
			continue
		}

		for _, se := range sessionExercises {
			if !se.Completed {
				continue
			}

			record, exists := exerciseRecords[se.ExerciseID]
			if !exists {
				record = &model.ExerciseProgressReport{
					ExerciseID:   se.ExerciseID,
					ExerciseName: se.Exercise.Name,
				}
				exerciseRecords[se.ExerciseID] = record
			}

			record.TotalSessions++

			// Track best weight
			if se.ActualWeight != nil && (record.BestWeight == nil || *se.ActualWeight > *record.BestWeight) {
				record.BestWeight = se.ActualWeight
			}

			// Track best reps
			if se.CompletedReps > record.BestReps {
				record.BestReps = se.CompletedReps
			}

			// Track best duration
			if se.ActualDurationSeconds != nil && (record.BestDurationSeconds == nil || *se.ActualDurationSeconds > *record.BestDurationSeconds) {
				record.BestDurationSeconds = se.ActualDurationSeconds
			}

			// Update last performed
			if session.CompletedAt.After(record.LastPerformed) {
				record.LastPerformed = *session.CompletedAt
			}
		}
	}

	// Convert map to slice
	var records []model.ExerciseProgressReport
	for _, record := range exerciseRecords {
		records = append(records, *record)
	}

	return records, nil
}

func GetWorkoutStreaks(userID int64) (int, int, error) {
	// Get all completed sessions ordered by date
	sessions, err := db.GetWorkoutSessionsByUserID(userID, "completed", 0)
	if err != nil {
		return 0, 0, err
	}

	if len(sessions) == 0 {
		return 0, 0, nil
	}

	// Track workout dates (ignoring time)
	workoutDates := make(map[string]bool)
	for _, session := range sessions {
		if session.CompletedAt != nil {
			date := session.CompletedAt.Format("2006-01-02")
			workoutDates[date] = true
		}
	}

	// Calculate current streak
	currentStreak := 0
	today := time.Now()

	for i := 0; i < 365; i++ { // Check up to a year
		date := today.AddDate(0, 0, -i).Format("2006-01-02")
		if workoutDates[date] {
			currentStreak++
		} else if i > 0 { // Allow one day gap (today might not have a workout yet)
			break
		}
	}

	// Calculate longest streak
	longestStreak := 0
	tempStreak := 0

	// Create sorted slice of dates
	var dates []time.Time
	for dateStr := range workoutDates {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}
		dates = append(dates, date)
	}

	// Sort dates
	for i := 0; i < len(dates)-1; i++ {
		for j := i + 1; j < len(dates); j++ {
			if dates[i].After(dates[j]) {
				dates[i], dates[j] = dates[j], dates[i]
			}
		}
	}

	// Calculate longest streak
	for i, date := range dates {
		if i == 0 {
			tempStreak = 1
		} else {
			prevDate := dates[i-1]
			if date.Sub(prevDate).Hours() <= 48 { // Within 2 days
				tempStreak++
			} else {
				if tempStreak > longestStreak {
					longestStreak = tempStreak
				}
				tempStreak = 1
			}
		}
	}

	if tempStreak > longestStreak {
		longestStreak = tempStreak
	}

	return currentStreak, longestStreak, nil
}
