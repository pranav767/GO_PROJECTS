package controller

import (
	"net/http"
	"strconv"
	"time"

	"workout-tracker/internal/service"
	"workout-tracker/model"

	"github.com/gin-gonic/gin"
)

func RegisterHandler(c *gin.Context) {
	var user model.User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request Format"})
		return
	}
	err = service.RegisterUser(user.Username, user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
}

func LoginHandler(c *gin.Context) {
	var user model.User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request Format"})
		return
	}
	exist, err := service.AuthenticateUser(user.Username, user.Password)
	if !exist {
		if err != nil {
			switch err.Error() {
			case "user not found":
				c.JSON(http.StatusInternalServerError, gin.H{"error": "User does not exist"})
			case "invalid passwd":
				c.JSON(http.StatusForbidden, gin.H{"error": "Invalid password"})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication failed"})
			}
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Auth failed"})
		}
		return
	}
	token, err := service.GenerateJWT(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

// Exercise Handlers

func GetExercisesHandler(c *gin.Context) {
	category := c.Query("category")
	muscleGroup := c.Query("muscle_group")

	exercises, err := service.GetExercises(category, muscleGroup)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve exercises"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"exercises": exercises})
}

func GetExerciseByIDHandler(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid exercise ID"})
		return
	}

	exercise, err := service.GetExerciseByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exercise not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"exercise": exercise})
}

// Workout Plan Handlers

func CreateWorkoutPlanHandler(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req model.CreateWorkoutPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	plan, err := service.CreateWorkoutPlan(userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"workout_plan": plan})
}

func GetWorkoutPlansHandler(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	activeOnly := c.Query("active") == "true"

	plans, err := service.GetUserWorkoutPlans(userID, activeOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve workout plans"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"workout_plans": plans})
}

func GetWorkoutPlanHandler(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	planIDParam := c.Param("id")
	planID, err := strconv.ParseInt(planIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workout plan ID"})
		return
	}

	plan, err := service.GetUserWorkoutPlan(userID, planID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"workout_plan": plan})
}

func UpdateWorkoutPlanHandler(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	planIDParam := c.Param("id")
	planID, err := strconv.ParseInt(planIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workout plan ID"})
		return
	}

	var req model.UpdateWorkoutPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	plan, err := service.UpdateWorkoutPlan(userID, planID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"workout_plan": plan})
}

func DeleteWorkoutPlanHandler(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	planIDParam := c.Param("id")
	planID, err := strconv.ParseInt(planIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workout plan ID"})
		return
	}

	err = service.DeleteWorkoutPlan(userID, planID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Workout plan deleted successfully"})
}

// Workout Session Handlers

func ScheduleWorkoutHandler(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req model.ScheduleWorkoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	session, err := service.ScheduleWorkout(userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"workout_session": session})
}

func GetWorkoutSessionsHandler(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	status := c.Query("status")
	limitParam := c.Query("limit")
	limit := 0
	if limitParam != "" {
		var err error
		limit, err = strconv.Atoi(limitParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
			return
		}
	}

	sessions, err := service.GetUserWorkoutSessions(userID, status, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve workout sessions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"workout_sessions": sessions})
}

func GetWorkoutSessionHandler(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	sessionIDParam := c.Param("id")
	sessionID, err := strconv.ParseInt(sessionIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workout session ID"})
		return
	}

	session, err := service.GetUserWorkoutSession(userID, sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"workout_session": session})
}

func StartWorkoutHandler(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	sessionIDParam := c.Param("id")
	sessionID, err := strconv.ParseInt(sessionIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workout session ID"})
		return
	}

	session, err := service.StartWorkoutSession(userID, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"workout_session": session})
}

func UpdateWorkoutSessionHandler(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	sessionIDParam := c.Param("id")
	sessionID, err := strconv.ParseInt(sessionIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workout session ID"})
		return
	}

	var req model.UpdateWorkoutSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	session, err := service.UpdateWorkoutSession(userID, sessionID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"workout_session": session})
}

func CompleteWorkoutHandler(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	sessionIDParam := c.Param("id")
	sessionID, err := strconv.ParseInt(sessionIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workout session ID"})
		return
	}

	var req model.UpdateWorkoutSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	req.Status = "completed" // Ensure status is set to completed

	session, err := service.CompleteWorkoutSession(userID, sessionID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"workout_session": session})
}

func GetUpcomingWorkoutsHandler(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	daysParam := c.Query("days")
	days := 7 // default to 7 days
	if daysParam != "" {
		var err error
		days, err = strconv.Atoi(daysParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid days parameter"})
			return
		}
	}

	sessions, err := service.GetUpcomingWorkouts(userID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve upcoming workouts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"upcoming_workouts": sessions})
}

// Report Handlers

func GenerateWorkoutReportHandler(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	startDateParam := c.Query("start_date")
	endDateParam := c.Query("end_date")

	// Default to last 30 days if no dates provided
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	if startDateParam != "" {
		var err error
		startDate, err = time.Parse("2006-01-02", startDateParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format. Use YYYY-MM-DD"})
			return
		}
	}

	if endDateParam != "" {
		var err error
		endDate, err = time.Parse("2006-01-02", endDateParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format. Use YYYY-MM-DD"})
			return
		}
	}

	report, err := service.GenerateWorkoutReport(userID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate workout report"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"report": report})
}

func GetPersonalRecordsHandler(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	records, err := service.GetPersonalRecords(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve personal records"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"personal_records": records})
}

func GetWorkoutStreaksHandler(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	currentStreak, longestStreak, err := service.GetWorkoutStreaks(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve workout streaks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"current_streak": currentStreak,
		"longest_streak": longestStreak,
	})
}

// Helper function to extract user ID from JWT context
func getUserIDFromContext(c *gin.Context) int64 {
	userID, exists := c.Get("userID")
	if !exists {
		return 0
	}

	return userID.(int64)
}
