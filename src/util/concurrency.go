package util

import (
	"context"

	"golang.org/x/sync/errgroup"
)

func ParallelMap[T any, R any](items []T, mapper func(item T) (R, error)) ([]R, error) {
	return ParallelMapContext(context.Background(), items, mapper)
}

func ParallelMapContext[T any, R any](ctx context.Context, items []T, mapper func(item T) (R, error)) ([]R, error) {
	wg, _ := errgroup.WithContext(ctx)

	results := make([]R, len(items))

	for i, item := range items {
		i, item := i, item
		wg.Go(func() error {
			mapped, err := mapper(item)
			results[i] = mapped
			return err
		})
	}

	return results, wg.Wait()
}
