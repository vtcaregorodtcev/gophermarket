package helpers

import (
	"golang.org/x/crypto/bcrypt"
)

const SALT = "salt"

func CredentialsHash(username, password string) (string, error) {
	data := []byte(username + password + SALT)

	hash, err := bcrypt.GenerateFromPassword(data, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}
