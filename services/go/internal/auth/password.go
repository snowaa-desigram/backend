package auth

import "golang.org/x/crypto/bcrypt"

var bcryptCost = 12

// SetBcryptCost — для тестов: cost 12 слишком медленный под -race.
func SetBcryptCost(cost int) { bcryptCost = cost }

func HashPassword(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(h), nil
}

func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
