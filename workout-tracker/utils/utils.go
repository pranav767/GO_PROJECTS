package utils
// hash the incoming password using bcrypt
import(
	"golang.org/x/crypto/bcrypt"
)

func GenerateHash(password []byte) ([]byte, error) {
	return bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
}

func CompareHashwithPassword(hashedPassword, password []byte) bool {
	return bcrypt.CompareHashAndPassword(hashedPassword, password) == nil
}