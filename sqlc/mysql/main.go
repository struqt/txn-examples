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

func tick(ctx context.Context, d dao.Demo, count int32) {
	log.V(0).Info(fmt.Sprintf("tick %d", count))
	_, _ = dao.ExecuteRw(ctx, d, do1(), dao.PushAuthorDo)
	_, _ = dao.ExecuteRo(ctx, d, &dao.ListAuthor{}, dao.ListAuthorDo)
	_, _ = dao.ExecuteRo(ctx, d, &dao.LastAuthor{}, dao.LastAuthorDo)
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
	defer d.Clear()
	run(ctx, func(i int32) { tick(ctx, d, i) })
}

func run(ctx context.Context, tick func(int32)) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
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
				go tick(count.Load())
			}
		}
	}(&wg)
	wg.Wait()
}

func address() string {
	var addr string
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
	return addr
}

func open() (error, *sql.DB, func()) {
	dsn := fmt.Sprintf("example:%s@%s/example?charset=utf8", os.Getenv("DB_PASSWORD"), address())
	db, err := sql.Open("mysql", dsn)
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
