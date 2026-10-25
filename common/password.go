package common

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func DoPasswordsMatch(plainText string, hash string) (bool, error) {
	if len(plainText) == 0 || len(hash) == 0 {
		return true, errors.New("empty password provided")
	}
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plainText))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
		}
		return false, err
	}
	return true, err
}
