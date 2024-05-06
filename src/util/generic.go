package util

import (
	"math/rand"
	"strconv"
)

const (
	alphabetLowerLatin = "abcdefghijklmnopqrstuvwxyz"
	alphabetUpperLatin = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	alphabetDigits     = "0123456789"
)

func Coalesce[T comparable](items ...T) T {
	var result T
	for _, item := range items {
		if item != result {
			return item
		}
	}
	return result
}

func StringToIntOrDefault(text string, dflt int) int {
	result, err := strconv.Atoi(text)
	if err != nil {
		return dflt
	}
	return result
}

func GenerateRandomString(length int) string {
	alphabet := alphabetLowerLatin + alphabetUpperLatin + alphabetDigits

	result := ""
	for i := 0; i < length; i += 1 {
		alphabetIndex := rand.Int31n(int32(len(alphabet)))
		result += string(alphabet[alphabetIndex])
	}
	return result
}
