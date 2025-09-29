package service

import (
	"crypto/rand"
	"encoding/hex"
)

// generateRandomHex generates a random hexadecimal string of specified length
func generateRandomHex(length int) string {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "0000000000000000000000000000000000000000000000000000000000000000"
	}
	return hex.EncodeToString(bytes)
}