package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/struqt/logging"
	"github.com/struqt/txn/txn_pgx"

	"examples/sqlc/pgx/demo"
)

var log = logging.NewLogger("")

func main() {
	defer os.Exit(0)
	defer log.Info("Process is ending ...")
	ctx, cancel := context.WithCancel(context.Background())
	defer log.Info("Context is cancelled")
	defer cancel()
	ds := fmt.Sprintf("postgres://example:%s@%s:5432/example",
		url.QueryEscape(os.Getenv("DB_PASSWORD")),
		os.Getenv("DB_HOST"),
	)
	pool, err := newPgxPool(ctx, ds)
	if err != nil {
		log.Error(err, "Failed to set up connection pool")
		return
	}
	defer func() {
		pool.Close()
		log.Info("Pgx Pool is closed.")
	}()
	var count atomic.Int32
	var wg sync.WaitGroup
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer func() { wg.Done() }()
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
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
				go tick(ctx, pool, count.Load())
			}
		}
	}(&wg)
	wg.Wait()
}

func tick(ctx context.Context, pool txn_pgx.PgxBeginner, count int32) {
	t0 := time.Now()
	if doer, err := txn_pgx.PgxExecute(ctx, pool, push(), PushAuthorDo); err != nil {
		log.Error(err, "tick", "count", count, "txn", doer.Title())
	} else {
		log.V(1).Info("", "title", doer.Title(), "duration", time.Now().Sub(t0), "inserted", doer.inserted)
	}
	t1 := time.Now()
	if doer, err := txn_pgx.PgxExecute(ctx, pool, fetch(), FetchLastAuthorDo); err != nil {
		log.Error(err, "tick", "count", count, "txn", doer.Title())
	} else {
		log.V(1).Info("", "title", doer.Title(), "duration", time.Now().Sub(t1), "id", doer.id)
	}
	log.Info("tick", "count", count)
}

func fetch() *FetchLastAuthorDoer {
	do := &FetchLastAuthorDoer{}
	do.SetRethrowPanic(false)
	do.SetTitle("Txn.FetchLastAuthor")
	do.SetTimeout(200 * time.Millisecond)
	do.SetOptions(&pgx.TxOptions{
		IsoLevel:       pgx.ReadCommitted,
		AccessMode:     pgx.ReadOnly,
		DeferrableMode: pgx.NotDeferrable,
		BeginQuery:     "",
	})
	return do
}

func push() *PushAuthorDoer {
	do := &PushAuthorDoer{}
	do.SetRethrowPanic(false)
	do.SetTitle("Txn.PushAuthor")
	do.SetTimeout(250 * time.Millisecond)
	do.SetOptions(&pgx.TxOptions{
		IsoLevel:       pgx.ReadCommitted,
		AccessMode:     pgx.ReadWrite,
		DeferrableMode: pgx.NotDeferrable,
		BeginQuery:     "",
	})
	do.insert = demo.CreateAuthorParams{
		Name: "Brian Kernighan",
		Bio: pgtype.Text{
			Valid:  true,
			String: "Co-author of The C Programming Language",
		},
	}
	return do
}

func newPgxPool(ctx context.Context, uri string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(uri)
	if err != nil {
		return nil, err
	}
	config.MinConns = 1
	config.MaxConns = 8
	return pgxpool.NewWithConfig(ctx, config)
}
