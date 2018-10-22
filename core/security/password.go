package security

import (
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

//HashAndSalt generates a new bcrypt hash from the provided password
func HashAndSalt(pwd []byte) (string, error) {

	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.DefaultCost)
	if err != nil {
		log.Error(err)
		return "", err
	}

	return string(hash), nil
}

// ComparePassword indicate if the plaintext password is the same as the hashed (stored) one.
func ComparePassword(hashedPwd string, plain string) bool {

	err := bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(plain))
	if err != nil {
		log.Error(err)
		return false
	}
	return true
}
