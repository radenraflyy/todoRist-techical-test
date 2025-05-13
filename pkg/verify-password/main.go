package verifypassword

import (
	"golang.org/x/crypto/bcrypt"
)

func VerifyPassword(userPassword string, hashedPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(userPassword))
	if err != nil {
		return err
	}
	return nil
}
