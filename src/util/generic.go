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

func StringToIntOrNull(text string) *int {
	result, err := strconv.Atoi(text)
	if err != nil {
		return nil
	}
	return &result
}

func StringToInt64OrDefault(text string, dflt int64) int64 {
	result, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return dflt
	}
	return result
}

func StringToBoolOrDefault(text string, dflt bool) bool {
	result, err := strconv.ParseBool(text)
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
