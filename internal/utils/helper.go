package utils

import (
	"strconv"
	"golang.org/x/crypto/bcrypt"
)

// Convert Functions
// String to Uint
func StringToUint(str string) (uint, error) {
	value, err := strconv.ParseUint(str, 10, 32)
	if err != nil { return 0, err }
	return uint(value), nil
}

// String to Int
func StringToInt(str string) (int, error) {
	value, err := strconv.Atoi(str)
	if err != nil { return 0, err }
	return value, nil
}

// Uint to String
func UintToString(value uint) string {
	return strconv.FormatUint(uint64(value), 10)
}

// Int to String
func IntToString(value int) string {
	return strconv.Itoa(value)
}

// Hashing Functions
// Hash Password
func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil { return "", err }
	return string(hashed), nil
}

// Check Password
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
