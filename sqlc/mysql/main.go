package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/struqt/logging"

	"examples/sqlc/mysql/dao"
	"examples/sqlc/mysql/demo"
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
	_, _ = dao.TxnRwExecute(ctx, d, push(), dao.PushAuthorDo)
	_, _ = dao.TxnRoExecute(ctx, d, &dao.ListAuthor{}, dao.ListAuthorDo)
	_, _ = dao.TxnRoExecute(ctx, d, &dao.LastAuthor{}, dao.LastAuthorDo)
	defer wg.Done()
}

func push() (do *dao.PushAuthor) {
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
	driver, uri, target := address()
	if err, db, clo = open(driver, uri); err != nil {
		log.Error(err, "")
		return
	}
	defer log.Info("Connection pool is closed")
	defer clo()
	defer log.Info("Connection cache is closed")
	for {
		if _, err = dao.TxnPing(ctx, db, func(cnt int, interval time.Duration) {
			log.Info("Ping", "count", cnt, "interval", interval, "target", target)
		}); err == nil {
			break
		}
	}
	m := dao.NewDemo(db)
	defer func() { _ = m.Close() }()
	run(ctx, func(i int32, wg *sync.WaitGroup) { tick(ctx, m, i, wg) })
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

func open(driver string, uri string) (error, *sql.DB, func()) {
	db, err := sql.Open(driver, uri)
	if err != nil {
		log.Error(err, "")
		return err, nil, nil
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(2)
	db.SetConnMaxIdleTime(300 * time.Second)
	return nil, db, func() {
		if db != nil {
			_ = db.Close()
		}
	}
}

func host() (addr string) {
	addr = os.Getenv("DB_ADDR_UDS")
	if len(addr) > 0 {
		addr = fmt.Sprintf("unix(%s)", addr)
	} else {
		addr = os.Getenv("DB_ADDR_TCP")
		if len(addr) <= 0 {
			addr = "127.0.0.1"
		}
		addr = fmt.Sprintf("tcp(%s)", addr)
	}
	return
}

func address() (string, string, string) {
	h := host()
	passwd := os.Getenv("DB_PASSWORD")
	return "mysql", fmt.Sprintf("example:%s@%s/example?charset=utf8", passwd, h), h
}
