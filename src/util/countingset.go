package util

type CountingSet[T comparable] struct {
	data map[T]int
	size int
}

func NewCountingSet[T comparable]() *CountingSet[T] {
	return &CountingSet[T]{
		data: make(map[T]int),
		size: 0,
	}
}

func (s *CountingSet[T]) Add(item T) {
	s.data[item] += 1
	s.size += 1
}

func (s *CountingSet[T]) Remove(item T) {
	s.size -= 1
	oldCount := s.data[item]
	if oldCount == 1 {
		delete(s.data, item)
	} else {
		s.data[item] = oldCount - 1
	}
}

func (s *CountingSet[T]) RemoveAll(item T) {
	s.size -= s.data[item]
	delete(s.data, item)
}

func (s *CountingSet[T]) TotalSize() int {
	return s.size
}

func (s *CountingSet[T]) UniqueSize() int {
	return len(s.data)
}

func (s *CountingSet[T]) GetDominatingValue(minPercentage float32) T {
	var result T
	var maxPercentage float32

	for item, count := range s.data {
		percentage := float32(count) / float32(s.size)
		if percentage >= minPercentage && percentage > maxPercentage {
			result = item
			maxPercentage = percentage
		}
	}

	return result
}
