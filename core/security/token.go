package security

import (
	"crypto/rand"
	"fmt"
)

//GenerateNewToken return a new random token string
func GenerateNewToken() string {
	b := make([]byte, 20)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
