package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/lib/pq"
	"github.com/struqt/logging"

	"examples/sqlc/pg/dao"
	"examples/sqlc/pg/demo"
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

func tick(ctx context.Context, d dao.Demo, count int32, wg *sync.WaitGroup) {
	log.V(0).Info(fmt.Sprintf("tick %d", count))
	_, _ = dao.Execute(ctx, d, do1(), dao.PushAuthorDo)
	_, _ = dao.ExecuteRo(ctx, d, &dao.ListAuthor{}, dao.ListAuthorDo)
	_, _ = dao.ExecuteRo(ctx, d, &dao.LastAuthor{}, dao.LastAuthorDo)
	defer wg.Done()
}

func do1() (do *dao.PushAuthor) {
	do = &dao.PushAuthor{}
	do.Insert = demo.CreateAuthorParams{
		Name: "Brian Kernighan",
		Bio: sql.NullString{
			Valid:  true,
			String: "Co-author of The C Programming Language",
		}}
	return
}

func main() {
	log.Info("Process is starting ...")
	defer os.Exit(0)
	defer log.Info("Process is ending ...")
	ctx, cancel := context.WithCancel(context.Background())
	defer log.Info("Context is cancelled")
	defer cancel()
	var (
		err error
		db  *sql.DB
		clo func()
	)
	if err, db, clo = open(); err != nil {
		log.Error(err, "")
		return
	}
	defer log.Info("Connection pool is closed")
	defer clo()
	defer log.Info("Connection cache is closed")
	d := dao.NewDemo(db)
	defer func() { _ = d.Close() }()
	run(ctx, func(i int32, wg *sync.WaitGroup) { tick(ctx, d, i, wg) })
}

func run(ctx context.Context, tick func(int32, *sync.WaitGroup)) {
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
				go tick(count.Load(), wg)
			}
		}
	}(&wg)
	wg.Wait()
}

func open() (error, *sql.DB, func()) {
	db, err := sql.Open(address())
	if err != nil {
		log.Error(err, "")
		return err, nil, nil
	}
	db.SetMaxOpenConns(2)
	db.SetMaxIdleConns(2)
	db.SetConnMaxIdleTime(600 * time.Second)
	return nil, db, func() {
		if db != nil {
			_ = db.Close()
		}
	}
}

func address() (string, string) {
	var addr string
	addr = os.Getenv("DB_ADDR_UDS")
	if len(addr) > 0 {

	} else {
		addr = os.Getenv("DB_ADDR_TCP")
		if len(addr) <= 0 {
			addr = "127.0.0.1"
		}

	}
	passwd := os.Getenv("DB_PASSWORD")
	return "postgres", fmt.Sprintf("sslmode=disable dbname=example user=example password=%s host=%s", passwd, addr)
}
