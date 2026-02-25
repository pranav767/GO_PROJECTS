package db

import (
	"leaderboard_system/model"
)

func CreateUser(username string, password_hash string) (int64, error) {
	result, err := db.Exec("INSERT into users (username, password_hash) VALUES (?,?)", username, password_hash)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func GetUserByUserName(username string) (*model.User, error) {
	var user model.User
	err := db.QueryRow("SELECT id, username, password_hash, role FROM users WHERE username = ? ", username).Scan(&user.ID, &user.Username, &user.Password, &user.Role)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func GetUserByID(userID int64) (*model.User, error) {
	var user model.User
	err := db.QueryRow("SELECT id, username, password_hash, role FROM users WHERE id = ?", userID).Scan(&user.ID, &user.Username, &user.Password, &user.Role)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUserRole updates the role of a user
func UpdateUserRole(userID int64, role string) error {
	_, err := db.Exec("UPDATE users SET role = ? WHERE id = ?", role, userID)
	return err
}