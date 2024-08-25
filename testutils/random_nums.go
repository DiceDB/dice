package testutils

import (
	"math/rand/v2"
)

func GenerateRandomString(length int, charset string) *string {
	outputBytes := make([]byte, length)
	for i := range outputBytes {
		outputBytes[i] = charset[rand.IntN(len(charset))]
	}
	outputStr := string(outputBytes)
	return &outputStr
}
