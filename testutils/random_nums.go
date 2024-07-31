package testutils

import (
	"math/rand"
	"time"
)

var randIndex = rand.New(rand.NewSource(time.Now().UnixNano()))

func GenerateRandomString(length int, charset string) *string {
	outputBytes := make([]byte, length)
	for i := range outputBytes {
		outputBytes[i] = charset[randIndex.Intn(len(charset))]
	}
	outputStr := string(outputBytes)
	return &outputStr
}
