package util

func Coalesce[T comparable](items ...T) T {
	var result T
	for _, item := range items {
		if item != result {
			return item
		}
	}
	return result
}
