package routes

import (
	"workout-tracker/internal/controller"
	"workout-tracker/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// Public routes (no authentication required)
	r.POST("/register", controller.RegisterHandler)
	r.POST("/login", controller.LoginHandler)

	// Protected routes (JWT authentication required)
	api := r.Group("/api")
	api.Use(middleware.JWTAuthMiddleware()) // Apply JWT middleware to this group
	{
		// Exercise endpoints
		api.GET("/exercises", controller.GetExercisesHandler)
		api.GET("/exercises/:id", controller.GetExerciseByIDHandler)

		// Workout plan endpoints
		api.POST("/workout-plans", controller.CreateWorkoutPlanHandler)
		api.GET("/workout-plans", controller.GetWorkoutPlansHandler)
		api.GET("/workout-plans/:id", controller.GetWorkoutPlanHandler)
		api.PUT("/workout-plans/:id", controller.UpdateWorkoutPlanHandler)
		api.DELETE("/workout-plans/:id", controller.DeleteWorkoutPlanHandler)

		// Workout session endpoints
		api.POST("/workout-sessions", controller.ScheduleWorkoutHandler)
		api.GET("/workout-sessions", controller.GetWorkoutSessionsHandler)
		api.GET("/workout-sessions/:id", controller.GetWorkoutSessionHandler)
		api.PUT("/workout-sessions/:id", controller.UpdateWorkoutSessionHandler)
		api.POST("/workout-sessions/:id/start", controller.StartWorkoutHandler)
		api.POST("/workout-sessions/:id/complete", controller.CompleteWorkoutHandler)
		api.GET("/upcoming-workouts", controller.GetUpcomingWorkoutsHandler)

		// Report endpoints
		api.GET("/reports/workout", controller.GenerateWorkoutReportHandler)
		api.GET("/reports/personal-records", controller.GetPersonalRecordsHandler)
		api.GET("/reports/streaks", controller.GetWorkoutStreaksHandler)
	}
}
