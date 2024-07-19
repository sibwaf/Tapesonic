package util

import (
	"context"
	"sync"

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

type StripedRwMutex struct {
	lock        *sync.Mutex
	lockCounter int64
	itemLocks   map[string]*StripedRwMutexToken
}

type StripedRwMutexToken struct {
	serial int64
	usages int
	lock   *sync.RWMutex
}

func NewStripedRwMutex() *StripedRwMutex {
	return &StripedRwMutex{
		lock:        &sync.Mutex{},
		lockCounter: 0,
		itemLocks:   map[string]*StripedRwMutexToken{},
	}
}

func (l *StripedRwMutex) LockForReading(id string) *StripedRwMutexToken {
	itemLock := l.acquireLock(id)
	itemLock.lock.RLock()
	return itemLock
}

func (l *StripedRwMutex) UnlockReader(id string, itemLock *StripedRwMutexToken) {
	itemLock.lock.RUnlock()
	l.releaseLock(id, itemLock)
}

func (l *StripedRwMutex) LockForWriting(id string) *StripedRwMutexToken {
	itemLock := l.acquireLock(id)
	itemLock.lock.Lock()
	return itemLock
}

func (l *StripedRwMutex) TryLockForWriting(id string) *StripedRwMutexToken {
	itemLock := l.acquireLock(id)
	if !itemLock.lock.TryLock() {
		l.releaseLock(id, itemLock)
		return nil
	}

	return itemLock
}

func (l *StripedRwMutex) UnlockWriter(id string, itemLock *StripedRwMutexToken) {
	itemLock.lock.Unlock()
	l.releaseLock(id, itemLock)
}

func (l *StripedRwMutex) acquireLock(id string) *StripedRwMutexToken {
	l.lock.Lock()
	defer l.lock.Unlock()

	itemLock := l.itemLocks[id]
	if itemLock == nil {
		l.lockCounter += 1

		itemLock = &StripedRwMutexToken{
			serial: l.lockCounter,
			usages: 0,
			lock:   &sync.RWMutex{},
		}
		l.itemLocks[id] = itemLock
	}

	itemLock.usages += 1

	return itemLock
}

func (l *StripedRwMutex) releaseLock(id string, itemLock *StripedRwMutexToken) {
	l.lock.Lock()
	defer l.lock.Unlock()

	itemLock.usages -= 1

	currentLock := l.itemLocks[id]
	if currentLock.serial != itemLock.serial {
		return
	}

	if currentLock.usages == 0 {
		delete(l.itemLocks, id)
	}
}
