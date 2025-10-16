package utils

// hash the incoming password using bcrypt
import (
	"os"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func GenerateHash(password []byte) ([]byte, error) {
	return bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
}

func CompareHashwithPassword(hashedPassword, password []byte) bool {
	return bcrypt.CompareHashAndPassword(hashedPassword, password) == nil
}

// GetEnvInt returns an int from env or default if unset/invalid
func GetEnvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}

// BookingCutoffDuration returns the configured booking cutoff (minutes) as a duration
func BookingCutoffDuration() time.Duration {
	mins := GetEnvInt("BOOKING_CUTOFF_MINUTES", 30)
	if mins < 0 {
		mins = 0
	}
	return time.Duration(mins) * time.Minute
}

// CancelCutoffDuration returns the configured cancellation cutoff (hours) as a duration
func CancelCutoffDuration() time.Duration {
	hrs := GetEnvInt("CANCEL_CUTOFF_HOURS", 2)
	if hrs < 0 {
		hrs = 0
	}
	return time.Duration(hrs) * time.Hour
}
