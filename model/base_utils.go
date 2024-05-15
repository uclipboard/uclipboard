package model

import (
	"math/rand"
	"time"
)

var seed = time.Now().Unix()

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandString(size int) string {
	randGen := rand.New(rand.NewSource(seed))
	// Define the characters that can be used in the random string

	// Create a byte slice to hold the random characters
	randomBytes := make([]byte, size)

	// Fill the byte slice with random characters from the charset
	for i := range randomBytes {
		randomBytes[i] = charset[randGen.Intn(len(charset))]
	}

	// Convert the byte slice to a string
	randomString := string(randomBytes)

	// Print the random string
	return randomString
}
