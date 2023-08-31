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

	"examples/sqlc/pg/demo"
)

var log = logging.NewLogger("")

func main() {
	defer os.Exit(0)
	defer log.Info("Process is ending ...")
	ctx, cancel := context.WithCancel(context.Background())
	defer log.Info("Context is cancelled")
	defer cancel()
	run(ctx)
}

func run(ctx context.Context) {
	var (
		dqc *DemoQueries
		err error
		clo []func()
	)
	if err, dqc, clo = cache(); err != nil {
		log.Error(err, "")
		return
	}
	defer log.Info("Connection pool is closed")
	defer func() {
		for _, c := range clo {
			c()
		}
	}()
	var wg sync.WaitGroup
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		var count atomic.Int32
		for {
			select {
			case <-ctx.Done():
				log.Info("Demo Ticker is stopping ...")
				return
			case <-ticker.C:
				count.Add(1)
				if count.Load() > 3 {
					return
				}
				go tick(ctx, dqc, count.Load())
			}
		}
	}(&wg)
	wg.Wait()
}

func tick(ctx context.Context, dq *DemoQueries, count int32) {
	if do, err := DemoExecute(ctx, dq, push(), PushAuthorDo); err != nil {
		log.Error(err, "tick", "count", count, "txn", do.Title())
	} else {
		log.V(1).Info("tick", "title", do.Title(), "inserted", do.inserted)
	}
	if do, err := DemoExecute(ctx, dq, fetch(), FetchLastAuthorDo); err != nil {
		log.Error(err, "tick", "count", count, "txn", do.Title())
	} else {
		log.V(1).Info("tick", "title", do.Title(), "id", do.id)
	}
	log.Info("tick", "count", count)
}

func fetch() *FetchLastAuthorDoer {
	do := &FetchLastAuthorDoer{}
	do.SetRethrowPanic(false)
	do.SetTitle("Txn.FetchLastAuthor")
	do.SetTimeout(200 * time.Millisecond)
	do.SetOptions(&sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  true,
	})
	return do
}

func push() *PushAuthorDoer {
	do := &PushAuthorDoer{
		Insert: demo.CreateAuthorParams{
			Name: "Brian Kernighan",
			Bio: sql.NullString{
				Valid:  true,
				String: "Co-author of The C Programming Language",
			},
		},
	}
	do.SetRethrowPanic(false)
	do.SetTitle("Txn.PushAuthor")
	do.SetTimeout(250 * time.Millisecond)
	do.SetOptions(&sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	})
	return do
}

func cache() (error, *DemoQueries, []func()) {
	dsn := fmt.Sprintf("sslmode=disable dbname=example host=%s user=example password=%s",
		os.Getenv("DB_HOST"), os.Getenv("DB_PASSWORD"))
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Error(err, "")
		return err, nil, nil
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxIdleTime(600 * time.Second)
	dq := &DemoQueries{db: db}
	return nil, dq,
		[]func(){
			func() { dq.Clear() },
			func() {
				if dq.db != nil {
					_ = dq.db.Close()
				}
			},
		}
}
