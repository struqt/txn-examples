package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/struqt/logging"

	"examples/mongo/dao"
)

var log logging.Logger

func init() {
	logging.LogConsoleThreshold = -128
	log = logging.NewLogger("")
}

func do(ctx context.Context, mod dao.TxnModule, tick int32) {
	log.Info(fmt.Sprintf("tick %d", tick))
	i := map[string]any{"name": "Brian Kernighan", "age": 30, "createdAt": time.Now()}
	_, _ = dao.TxnExecute(ctx, mod, &dao.ListAuthor{}, dao.ListAuthorDo)
	_, _ = dao.TxnExecute(ctx, mod, &dao.PushAuthor{Insert: i}, dao.PushAuthorDo)
	_, _ = dao.TxnExecute(ctx, mod, &dao.ListAuthor{}, dao.ListAuthorDo)
}

func main() {
	defer os.Exit(0)
	defer log.Info("Process is ending ...")
	log.Info("Process is starting ...")
	ctx, cancel := context.WithCancel(context.Background())
	defer log.Info("Context is cancelled")
	defer cancel()
	defer func() {
		if dao.Demo() != nil && dao.Demo().Beginner() != nil {
			_ = dao.Demo().Beginner().Disconnect(ctx)
		}
	}()
	dao.Setup(log)
	do(ctx, dao.Demo(), 0)
	run(ctx, func(tick int32) { do(ctx, dao.Demo(), tick) })
}

func run(ctx context.Context, tick func(int32)) {
	const (
		TickCount      = 5
		TickIntervalMs = 300
	)
	var wg sync.WaitGroup
	wg.Add(TickCount)
	go func(wg *sync.WaitGroup) {
		ticker := time.NewTicker(TickIntervalMs * time.Millisecond)
		defer ticker.Stop()
		var count atomic.Int32
		for {
			select {
			case <-ctx.Done():
				log.Info("Demo Ticker is stopping ...")
				return
			case <-ticker.C:
				count.Add(1)
				if count.Load() > TickCount {
					return
				}
				go func() {
					defer wg.Done()
					tick(count.Load())
				}()
			}
		}
	}(&wg)
	wg.Wait()
}
