package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/struqt/logging"
	"github.com/struqt/txn"

	"examples/sqlc/pgx/dao"
	"examples/sqlc/pgx/dao/demo"
)

const (
	TickCount      = 6
	TickIntervalMs = 500
)

var log logging.Logger

func init() {
	logging.LogConsoleThreshold = -1
	log = logging.NewLogger("")
	dao.Setup(log)
}

func tick(ctx context.Context, mod dao.Demo, count int32, wg *sync.WaitGroup) {
	defer wg.Done()
	log.V(0).Info(fmt.Sprintf("tick %d", count))
	_, _ = stat(ctx, mod)
	_, _ = dao.TxnRwExecute(ctx, mod, push(), dao.PushAuthorDo)
	_, _ = dao.TxnRoExecute(ctx, mod, &dao.DemoDoer[[]demo.Author]{}, dao.ListAuthorDo, txn.WithTitle("TxnRo`List"))
	_, _ = dao.TxnRoExecute(ctx, mod, &dao.LastAuthor{}, dao.LastAuthorDo)
}

func stat(ctx context.Context, mod dao.Demo) (*demo.StatAuthorRow, error) {
	type _doer = dao.DemoDoer[demo.StatAuthorRow]
	do, err := dao.TxnRoExecute(ctx, mod, &_doer{}, func(ctx context.Context, do *_doer) error {
		if result, err := do.Stmt().StatAuthor(ctx); err != nil {
			return err
		} else {
			do.Result = result
			log.WithName(do.Title()).V(2).Info("     :", "result", do.Result)
			return nil
		}
	}, txn.WithTitle("TxnRo`Stat"))
	return &do.Result, err
}

func push() *dao.PushAuthor {
	do := &dao.PushAuthor{}
	do.Insert = demo.CreateAuthorParams{
		Name: "Brian Kernighan",
		Bio: pgtype.Text{
			Valid:  true,
			String: "Co-author of The C Programming Language",
		},
	}
	return do
}

func main() {
	defer os.Exit(0)
	defer log.Info("Process is ending ...")
	ctx, cancel := context.WithCancel(context.Background())
	defer log.Info("Context is cancelled")
	defer cancel()
	addr, uri := address()
	pool, err := open(ctx, uri)
	if err != nil {
		log.Error(err, "Failed to set up connection pool")
		return
	}
	defer func() {
		pool.Close()
		log.Info("Pgx Pool is closed.")
	}()
	var cnt int
	for {
		if cnt, err = dao.TxnPing(pool, func(cnt int, delay time.Duration) {
			log.Info("Ping", "count", cnt, "delay~", delay, "target", addr)
		}); err == nil {
			break
		} else {
			log.V(1).Info("Ping", "count", cnt, "err", err)
		}
	}
	d := dao.NewDemo(pool)
	var count atomic.Int32
	var wg sync.WaitGroup
	wg.Add(TickCount)
	go func(wg *sync.WaitGroup) {
		ticker := time.NewTicker(TickIntervalMs * time.Millisecond)
		defer ticker.Stop()
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
				go tick(ctx, d, count.Load(), wg)
			}
		}
	}(&wg)
	wg.Wait()
}

func open(ctx context.Context, uri string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(uri)
	if err != nil {
		return nil, err
	}
	config.MinConns = 1
	config.MaxConns = 8
	return pgxpool.NewWithConfig(ctx, config)
}

func host() (addr string) {
	addr = os.Getenv("DB_ADDR_UDS")
	if len(addr) > 0 {
		addr = fmt.Sprintf("%s", addr)
	} else {
		addr = os.Getenv("DB_ADDR_TCP")
		if len(addr) <= 0 {
			addr = "127.0.0.1"
		}
		addr = fmt.Sprintf("%s:5432", addr)
	}
	return
}

func address() (string, string) {
	h := host()
	passwd := os.Getenv("DB_PASSWORD")
	return h, fmt.Sprintf("postgres://example:%s@%s/example", url.QueryEscape(passwd), h)
}
