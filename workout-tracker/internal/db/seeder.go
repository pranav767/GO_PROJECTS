package db

import (
	"log"
)

// SeedExercises populates the database with predefined exercises
func SeedExercises() error {
	exercises := []struct {
		name        string
		description string
		category    string
		muscleGroup string
	}{
		// Chest exercises
		{"Push-ups", "Body weight exercise targeting chest, shoulders, and triceps", "strength", "chest"},
		{"Bench Press", "Classic barbell or dumbbell press for chest development", "strength", "chest"},
		{"Chest Flyes", "Isolation exercise for chest using dumbbells or cables", "strength", "chest"},
		{"Incline Bench Press", "Upper chest focused pressing movement", "strength", "chest"},
		{"Decline Push-ups", "Elevated feet push-ups for upper chest emphasis", "strength", "chest"},

		// Back exercises
		{"Pull-ups", "Body weight vertical pulling exercise", "strength", "back"},
		{"Deadlifts", "Compound movement targeting posterior chain", "strength", "back"},
		{"Bent-over Rows", "Horizontal pulling movement with barbell or dumbbells", "strength", "back"},
		{"Lat Pulldowns", "Cable exercise targeting latissimus dorsi", "strength", "back"},
		{"T-Bar Rows", "Compound rowing movement", "strength", "back"},

		// Shoulder exercises
		{"Overhead Press", "Vertical pressing movement for shoulders", "strength", "shoulders"},
		{"Lateral Raises", "Isolation exercise for shoulder width", "strength", "shoulders"},
		{"Front Raises", "Anterior deltoid isolation exercise", "strength", "shoulders"},
		{"Rear Delt Flyes", "Posterior deltoid isolation movement", "strength", "shoulders"},
		{"Pike Push-ups", "Body weight shoulder exercise", "strength", "shoulders"},

		// Arms exercises
		{"Bicep Curls", "Classic arm isolation exercise", "strength", "arms"},
		{"Tricep Dips", "Body weight tricep exercise", "strength", "arms"},
		{"Hammer Curls", "Neutral grip bicep exercise", "strength", "arms"},
		{"Tricep Extensions", "Overhead tricep isolation", "strength", "arms"},
		{"Close-grip Push-ups", "Push-up variation emphasizing triceps", "strength", "arms"},

		// Legs exercises
		{"Squats", "Fundamental lower body compound movement", "strength", "legs"},
		{"Lunges", "Single-leg strength and stability exercise", "strength", "legs"},
		{"Calf Raises", "Isolation exercise for calves", "strength", "legs"},
		{"Leg Press", "Machine-based leg exercise", "strength", "legs"},
		{"Bulgarian Split Squats", "Single-leg squat variation", "strength", "legs"},
		{"Romanian Deadlifts", "Hip-hinge movement targeting hamstrings", "strength", "legs"},

		// Core exercises
		{"Plank", "Isometric core stability exercise", "strength", "core"},
		{"Crunches", "Classic abdominal exercise", "strength", "core"},
		{"Russian Twists", "Rotational core exercise", "strength", "core"},
		{"Mountain Climbers", "Dynamic core and cardio exercise", "strength", "core"},
		{"Dead Bug", "Core stability and coordination exercise", "strength", "core"},
		{"Bicycle Crunches", "Dynamic abdominal exercise", "strength", "core"},

		// Cardio exercises
		{"Running", "Classic cardiovascular exercise", "cardio", "cardio"},
		{"Cycling", "Low-impact cardio exercise", "cardio", "cardio"},
		{"Jump Rope", "High-intensity cardio exercise", "cardio", "cardio"},
		{"Burpees", "Full-body explosive movement", "cardio", "full_body"},
		{"High Knees", "Running in place with high knee lifts", "cardio", "cardio"},
		{"Jumping Jacks", "Classic calisthenics exercise", "cardio", "full_body"},

		// Flexibility exercises
		{"Forward Fold", "Hamstring and lower back stretch", "flexibility", "legs"},
		{"Downward Dog", "Full-body yoga stretch", "flexibility", "full_body"},
		{"Child's Pose", "Relaxing back and hip stretch", "flexibility", "back"},
		{"Cat-Cow Stretch", "Spinal mobility exercise", "flexibility", "back"},
		{"Pigeon Pose", "Hip flexor and glute stretch", "flexibility", "legs"},
		{"Shoulder Rolls", "Shoulder mobility exercise", "flexibility", "shoulders"},

		// Balance exercises
		{"Single-leg Stand", "Basic balance challenge", "balance", "legs"},
		{"Tree Pose", "Yoga balance pose", "balance", "legs"},
		{"Balance Board", "Proprioception training", "balance", "legs"},
		{"Warrior III", "Dynamic balance yoga pose", "balance", "full_body"},

		// Full body exercises
		{"Burpees", "Full-body conditioning exercise", "strength", "full_body"},
		{"Thrusters", "Squat to press combination", "strength", "full_body"},
		{"Turkish Get-ups", "Complex full-body movement", "strength", "full_body"},
		{"Bear Crawl", "Quadrupedal movement pattern", "strength", "full_body"},
		{"Man Makers", "Burpee with dumbbell rows", "strength", "full_body"},
	}

	for _, exercise := range exercises {
		// Check if exercise already exists
		existing, err := GetAllExercises()
		if err != nil {
			return err
		}

		exists := false
		for _, ex := range existing {
			if ex.Name == exercise.name {
				exists = true
				break
			}
		}

		if !exists {
			_, err = CreateExercise(exercise.name, exercise.description, exercise.category, exercise.muscleGroup)
			if err != nil {
				log.Printf("Error creating exercise %s: %v", exercise.name, err)
				return err
			}
			log.Printf("Created exercise: %s", exercise.name)
		}
	}

	log.Println("Exercise seeding completed successfully")
	return nil
}
