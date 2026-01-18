package scheduling

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"tapesonic/config"
	"time"
)

type Looper struct {
	startBarrier *sync.WaitGroup

	running bool
}

func newLooper() *Looper {
	startBarrier := &sync.WaitGroup{}
	startBarrier.Add(1)

	return &Looper{
		startBarrier: startBarrier,
		running:      true,
	}
}

func (looper *Looper) RegisterInterval(name string, interval time.Duration, code func() error) {
	go func() {
		looper.startBarrier.Wait()

		for looper.running {
			slog.Log(context.Background(), config.LevelTrace, fmt.Sprintf("Running background task %s", name))

			err := code()
			if err != nil {
				slog.Error(fmt.Sprintf("Background task %s failed: %s", name, err.Error()))
			} else {
				slog.Log(context.Background(), config.LevelTrace, fmt.Sprintf("Background task %s succeeded", name))
			}

			time.Sleep(interval)
		}
	}()
}

func (looper *Looper) Start() {
	looper.startBarrier.Done()
}

func (looper *Looper) Stop() {
	// todo: proper graceful stopping with completion waiting
	looper.running = false
}
