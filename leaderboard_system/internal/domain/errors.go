package domain

import "errors"

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrUserExists       = errors.New("user already exists")
	ErrInvalidPassword  = errors.New("invalid password")
	ErrGameNotFound     = errors.New("game not found")
	ErrRedisUnavailable = errors.New("redis unavailable")
)
