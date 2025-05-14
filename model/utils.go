package model

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

var seedSrc = rand.NewSource(time.Now().Unix())
var randGen = rand.New(seedSrc)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandString(size int) string {
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

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

const encryptSalt = "uclipboard:%s"

func TokenEncrypt(token string) string {
	// encrypt token with 3 pheases md5
	md5_1 := GetMD5Hash(fmt.Sprintf(encryptSalt, token))
	md5_2 := GetMD5Hash(fmt.Sprintf(encryptSalt, md5_1))
	md5_3 := GetMD5Hash(fmt.Sprintf(encryptSalt, md5_2))

	return md5_3
}

// get the current executable file's directory
func ExDir() string {
	exPath, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return filepath.Dir(exPath)
}

func ExPath() string {
	exPath, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return exPath
}
