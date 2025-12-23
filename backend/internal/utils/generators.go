package utils

import (
	"math/rand"

	"github.com/google/uuid"
)

// generates user id using uuid
func GenerateUserID() string {
	id := uuid.New()
	return id.String()
}

// generates random 6 character room code
func GenerateRoomCode() string {
	const characters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	b := make([]byte, 6)
	for i := range b {
		b[i] = characters[rand.Intn(len(characters))]
	}

	return string(b)
}
