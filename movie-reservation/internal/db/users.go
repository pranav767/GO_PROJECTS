package db

import (
	"database/sql"
	"errors"
	"log"
	"movie-reservation/model"
	"movie-reservation/utils"
	"os"
)

func CreateUser(username string, password_hash string, role string) (int64, error) {
	if role == "" {
		role = "user"
	}
	// Insert with explicit empty strings for nullable columns to avoid scanning NULL later
	result, err := db.Exec("INSERT INTO users (username, password_hash, role, email, full_name) VALUES (?,?,?,?,?)", username, password_hash, role, "", "")
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func GetUserByUserName(username string) (*model.User, error) {
	var user model.User
	err := db.QueryRow(
		"SELECT id, username, password_hash, role, COALESCE(email,''), COALESCE(full_name,''), created_at, updated_at FROM users WHERE username = ?", username,
	).Scan(&user.ID, &user.Username, &user.Password, &user.Role, &user.Email, &user.FullName, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	return &user, nil
}

func GetUserByID(userID int64) (*model.User, error) {
	var user model.User
	err := db.QueryRow("SELECT id, username, password_hash, role, COALESCE(email,''), COALESCE(full_name,''), created_at, updated_at FROM users WHERE id = ?", userID).
		Scan(&user.ID, &user.Username, &user.Password, &user.Role, &user.Email, &user.FullName, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func UpdateUserRole(userID int64, role string) error {
	_, err := db.Exec("UPDATE users SET role = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", role, userID)
	return err
}

func GetAllUsers() ([]model.User, error) {
	query := "SELECT id, username, role, COALESCE(email,''), COALESCE(full_name,''), created_at, updated_at FROM users ORDER BY created_at DESC"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		err := rows.Scan(&user.ID, &user.Username, &user.Role, &user.Email, &user.FullName, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

// SyncAdminPasswordFromEnv resets the admin user's password if ADMIN_DEFAULT_PASSWORD is set
// and differs from the current stored hash. This runs at startup to ensure the documented
// credentials match expectations without requiring a manual SQL update. It is idempotent.
func SyncAdminPasswordFromEnv() {
	pw := os.Getenv("ADMIN_DEFAULT_PASSWORD")
	if pw == "" { // nothing to do
		return
	}
	admin, err := GetUserByUserName("admin")
	if err != nil || admin == nil {
		return
	}
	if utils.CompareHashwithPassword([]byte(admin.Password), []byte(pw)) {
		return // already matches desired password
	}
	hash, err := utils.GenerateHash([]byte(pw))
	if err != nil {
		log.Printf("Admin password sync: failed to generate hash: %v", err)
		return
	}
	if _, err := db.Exec("UPDATE users SET password_hash = ? WHERE id = ?", string(hash), admin.ID); err != nil {
		log.Printf("Admin password sync: update failed: %v", err)
		return
	}
	log.Printf("Admin password updated from ADMIN_DEFAULT_PASSWORD environment variable")
}
