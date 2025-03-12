package util

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword membuat hash dari password menggunakan bcrypt
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("gagal melakukan hash password: %w", err)
	}
	return string(hashedPassword), nil
}

// CheckPassword memeriksa apakah password sesuai dengan hash yang tersimpan
func CheckPassword(password string, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
