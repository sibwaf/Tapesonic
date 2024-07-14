package util

import (
	"fmt"
	"math"
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

func TakeIf[T any](value *T, condition bool) *T {
	if condition {
		return value
	} else {
		return nil
	}
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

func StringToInt64OrNull(text string) *int64 {
	result, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return nil
	}
	return &result
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

func FormatBytes(size int64) string {
	return FormatBytesWithMagnitude(size, size)
}

func FormatBytesWithMagnitude(size int64, magnitude int64) string {
	if magnitude < 0 {
		magnitude = -magnitude
	}

	if magnitude < 1024 {
		return fmt.Sprintf("%d B", size)
	}

	sign := float64(1)
	if size < 0 {
		sign = -1
	}

	magnitude /= 1024
	result := math.Abs(float64(size)) / 1024
	suffix := "KiB"

	if magnitude >= 1024 {
		result /= 1024
		magnitude /= 1024
		suffix = "MiB"
	}
	if magnitude >= 1024 {
		result /= 1024
		magnitude /= 1024
		suffix = "GiB"
	}

	return fmt.Sprintf("%.2f %s", result*sign, suffix)
}
