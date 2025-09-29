package db

import (
	"workout-tracker/model"
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
	err := db.QueryRow("SELECT id, username, password_hash FROM users WHERE username = ? ", username).Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
