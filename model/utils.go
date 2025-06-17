package model

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	randSeedSrc = rand.NewSource(time.Now().Unix())
	randGen     = rand.New(randSeedSrc)
	randMutex   = sync.Mutex{}
)

func RandDecIntString(size int) string {
	randMutex.Lock()
	defer randMutex.Unlock()
	var sb strings.Builder
	sb.Grow(size) // Pre-allocate memory to avoid reallocations

	for i := 0; i < size; i++ {
		sb.WriteByte(byte(randGen.Intn(10) + '0'))
	}

	return sb.String()
}

func RandString(size int) string {
	randMutex.Lock()
	defer randMutex.Unlock()
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	var sb strings.Builder
	sb.Grow(size) // Pre-allocate memory to avoid reallocations

	for i := 0; i < size; i++ {
		sb.WriteByte(charset[randGen.Intn(len(charset))])
	}

	return sb.String()
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

func MaxLimitExpoGrowthAlgo(logger *logrus.Entry, initDelay, maxDelay time.Duration, F func() bool) {
	currentDelay := initDelay
	for i := 0; ; i++ {
		logger.Debugf("MaxLimitExpoGrowthAlgo attempt %d. Waiting for %v before retrying.", i, currentDelay)

		time.Sleep(currentDelay)

		if !F() {
			currentDelay *= 2
			if currentDelay > maxDelay {
				currentDelay = maxDelay
			}
		} else {
			return
		}
	}
}
