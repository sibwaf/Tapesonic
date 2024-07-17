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
	lock      *sync.Mutex
	itemLocks map[string]*sync.RWMutex
}

type StripedRwMutexToken struct {
	lock *sync.RWMutex
}

func NewStripedRwMutex() *StripedRwMutex {
	return &StripedRwMutex{
		lock:      &sync.Mutex{},
		itemLocks: map[string]*sync.RWMutex{},
	}
}

func (l *StripedRwMutex) LockForReading(id string) *StripedRwMutexToken {
	l.lock.Lock()
	defer l.lock.Unlock()

	itemLock := l.getOrCreateLock(id)
	itemLock.lock.RLock()
	return itemLock
}

func (l *StripedRwMutex) UnlockReader(id string, itemLock *StripedRwMutexToken) {
	l.lock.Lock()
	defer l.lock.Unlock()

	itemLock.lock.RUnlock()

	// no one else can change item lock's state right now so if we're able to
	// write-lock it we can safely delete it because it's not used by anyone;
	// if we failed to get a lock, the current lock's user will deal with it

	if itemLock.lock.TryLock() {
		itemLock.lock.Unlock()
		delete(l.itemLocks, id)
	}
}

func (l *StripedRwMutex) LockForWriting(id string) *StripedRwMutexToken {
	l.lock.Lock()
	defer l.lock.Unlock()

	itemLock := l.getOrCreateLock(id)
	itemLock.lock.Lock()
	return itemLock
}

func (l *StripedRwMutex) TryLockForWriting(id string) *StripedRwMutexToken {
	l.lock.Lock()
	defer l.lock.Unlock()

	itemLock := l.getOrCreateLock(id)
	if !itemLock.lock.TryLock() {
		return nil
	}
	return itemLock
}

func (l *StripedRwMutex) UnlockWriter(id string, itemLock *StripedRwMutexToken) {
	l.lock.Lock()
	defer l.lock.Unlock()

	itemLock.lock.Unlock()

	// see UnlockReader

	if itemLock.lock.TryLock() {
		itemLock.lock.Unlock()
		delete(l.itemLocks, id)
	}
}

func (l *StripedRwMutex) getOrCreateLock(id string) *StripedRwMutexToken {
	itemLock := l.itemLocks[id]
	if itemLock == nil {
		itemLock = &sync.RWMutex{}
		l.itemLocks[id] = itemLock
	}

	return &StripedRwMutexToken{lock: itemLock}
}
