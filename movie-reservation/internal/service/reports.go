package service

import (
	"errors"
	"movie-reservation/internal/db"
	"movie-reservation/model"
	"time"
)

// Reporting service for admin analytics

func GenerateRevenueReport(startDate, endDate string) (*model.RevenueReport, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, errors.New("invalid start date format, use YYYY-MM-DD")
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, errors.New("invalid end date format, use YYYY-MM-DD")
	}

	if start.After(end) {
		return nil, errors.New("start date must be before or equal to end date")
	}

	// Add 23:59:59 to end date to include the full day
	end = end.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	return db.GetRevenueReport(start, end)
}

func GenerateCapacityReport(startDate, endDate string) ([]model.CapacityReport, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, errors.New("invalid start date format, use YYYY-MM-DD")
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, errors.New("invalid end date format, use YYYY-MM-DD")
	}

	if start.After(end) {
		return nil, errors.New("start date must be before or equal to end date")
	}

	return db.GetCapacityReport(start, end)
}

func GetPopularMovies(startDate, endDate string, limit int) ([]model.MovieRevenue, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, errors.New("invalid start date format, use YYYY-MM-DD")
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, errors.New("invalid end date format, use YYYY-MM-DD")
	}

	if start.After(end) {
		return nil, errors.New("start date must be before or equal to end date")
	}

	if limit <= 0 || limit > 50 {
		limit = 10 // Default limit
	}

	end = end.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	return db.GetPopularMovies(start, end, limit)
}

func GetDailyStats(date string) (map[string]interface{}, error) {
	targetDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, errors.New("invalid date format, use YYYY-MM-DD")
	}

	return db.GetDailyReservationStats(targetDate)
}

func GetWeeklyReport() (map[string]interface{}, error) {
	// Get the start of the current week (Sunday)
	now := time.Now()
	weekday := int(now.Weekday())
	startOfWeek := now.AddDate(0, 0, -weekday)
	endOfWeek := startOfWeek.AddDate(0, 0, 6)

	report := make(map[string]interface{})

	// Get revenue for the week
	revenue, err := db.GetRevenueReport(startOfWeek, endOfWeek)
	if err != nil {
		return nil, err
	}

	report["revenue"] = revenue

	// Get capacity report
	capacity, err := db.GetCapacityReport(startOfWeek, endOfWeek)
	if err != nil {
		return nil, err
	}

	report["capacity"] = capacity

	// Get popular movies
	popular, err := db.GetPopularMovies(startOfWeek, endOfWeek, 5)
	if err != nil {
		return nil, err
	}

	report["popular_movies"] = popular
	report["period"] = startOfWeek.Format("2006-01-02") + " to " + endOfWeek.Format("2006-01-02")

	return report, nil
}

func GetMonthlyReport(year int, month int) (map[string]interface{}, error) {
	if year < 2020 || year > time.Now().Year()+1 {
		return nil, errors.New("invalid year")
	}
	if month < 1 || month > 12 {
		return nil, errors.New("invalid month")
	}

	startOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, -1)

	report := make(map[string]interface{})

	// Get revenue for the month
	revenue, err := db.GetRevenueReport(startOfMonth, endOfMonth)
	if err != nil {
		return nil, err
	}

	report["revenue"] = revenue

	// Get capacity report
	capacity, err := db.GetCapacityReport(startOfMonth, endOfMonth)
	if err != nil {
		return nil, err
	}

	report["capacity"] = capacity

	// Get popular movies
	popular, err := db.GetPopularMovies(startOfMonth, endOfMonth, 10)
	if err != nil {
		return nil, err
	}

	report["popular_movies"] = popular
	report["period"] = startOfMonth.Format("2006-01-02") + " to " + endOfMonth.Format("2006-01-02")

	return report, nil
}

// Real-time dashboard data
func GetDashboardData() (map[string]interface{}, error) {
	dashboard := make(map[string]interface{})

	// Today's stats
	today := time.Now().Format("2006-01-02")
	dailyStats, err := GetDailyStats(today)
	if err != nil {
		return nil, err
	}
	dashboard["today"] = dailyStats

	// This week's revenue
	weeklyReport, err := GetWeeklyReport()
	if err != nil {
		return nil, err
	}
	dashboard["this_week"] = weeklyReport

	// Popular movies (last 30 days)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	popular, err := db.GetPopularMovies(thirtyDaysAgo, time.Now(), 5)
	if err != nil {
		return nil, err
	}
	dashboard["popular_movies_30_days"] = popular

	// Current active movies
	movies, err := db.GetAllMovies(false)
	if err != nil {
		return nil, err
	}
	dashboard["active_movies_count"] = len(movies)

	return dashboard, nil
}
