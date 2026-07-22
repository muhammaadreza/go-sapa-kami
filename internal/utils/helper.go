package utils

import (
	"strconv"
	"crypto/rand"
	"encoding/base64"
)

// Converts a string to uint
func StringToUint(str string) (uint, error) {
	value, err := strconv.ParseUint(str, 10, 32)
	if err != nil { return 0, err }

	return uint(value), nil
}

// Converts a uint to string
func UintToString(value uint) string {
	return strconv.FormatUint(uint64(value), 10)
}

// Converts an int to string
func IntToString(value int) string {
	return strconv.Itoa(value)
}

// Generate a random state string
func GenerateState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil { return "", err }

	return base64.URLEncoding.EncodeToString(b), nil
}
