package util

import "strconv"

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
