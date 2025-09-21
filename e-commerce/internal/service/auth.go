package service

import(
	"e-commerce/internal/db"
	"e-commerce/utils"
	"errors"
)

func RegisterUser(username, password string) error {

	// Check if user exists
	existing, _ := db.GetUserByUsername(username)
	if existing != nil {
		return errors.New("user already exists")
	}
	hash, err := utils.GenerateHash([]byte(password))
	if err != nil {
		return err
	}

	_, err = db.CreateUser(username, string(hash))
	return err
}

// AuthenticateUser checks credentials and returns true if valid
func AuthenticateUser(username, password string) (bool, error) {
    user, err := db.GetUserByUsername(username)
    if err != nil {
        return false, errors.New("user not found")
    }
    if !utils.CompareHash([]byte(user.PasswordHash), []byte(password)) {
        return false, errors.New("invalid password")
    }
    return true, nil
}
