package dao

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	once    sync.Once
	log     *slog.Logger
	demoMod DemoModule
)

func Setup(logger *slog.Logger) {
	once.Do(func() {
		log = logger
		initDemo()
	})
}

func TearDown() {
	if Demo() != nil && Demo().Beginner() != nil {
		Demo().Beginner().Close()
		log.Info("Pgx Pool is closed.")
	}
}

func Demo() DemoModule {
	return demoMod
}

func initDemo() {
	addr, uri := address()
	pool, err := open(context.Background(), uri)
	if err != nil {
		log.Error(err.Error(), "Failed to set up connection pool")
		return
	}
	ping(pool, addr)
	demoMod = NewDemo(pool)
}

func ping(pool TxnBeginner, addr string) {
	for {
		if cnt, err := TxnPing(pool, func(cnt int, delay time.Duration) {
			log.Info("Ping", "count", cnt, "delay~", delay, "target", addr)
		}); err == nil {
			break
		} else {
			log.Info("Ping", "count", cnt, "err", err)
		}
	}
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
