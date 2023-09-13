package dao

import (
	"database/sql"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/struqt/logging"
)

var (
	once    sync.Once
	log     logging.Logger
	demoMod DemoModule
)

func Setup(logger logging.Logger) {
	once.Do(func() {
		log = logger
		initDemo()
	})
}

func TearDown() {
	if Demo() != nil {
		_ = Demo().Close()
	}
	if Demo().Beginner() != nil {
		_ = Demo().Beginner().Close()
	}
}

func Demo() DemoModule {
	return demoMod
}

func initDemo() {
	driver, uri, addr := address()
	var (
		err error
		db  *sql.DB
	)
	if err, db = open(driver, uri); err != nil {
		log.Error(err, "Failed to set up connection pool")
		return
	}
	ping(db, addr)
	demoMod = NewDemo(db)
}

func ping(db TxnBeginner, addr string) {
	for {
		if cnt, err := TxnPing(db, func(cnt int, delay time.Duration) {
			log.Info("Ping", "count", cnt, "delay~", delay, "target", addr)
		}); err == nil {
			break
		} else {
			log.V(1).Info("Ping", "count", cnt, "err", err)
		}
	}
}

func open(driver string, uri string) (error, *sql.DB) {
	db, err := sql.Open(driver, uri)
	if err != nil {
		log.Error(err, "")
		return err, nil
	}
	db.SetMaxOpenConns(2)
	db.SetMaxIdleConns(2)
	db.SetConnMaxIdleTime(600 * time.Second)
	return nil, db
}

func host() (addr string) {
	addr = os.Getenv("DB_ADDR_UDS")
	if len(addr) > 0 {
	} else {
		addr = os.Getenv("DB_ADDR_TCP")
		if len(addr) <= 0 {
			addr = "127.0.0.1"
		}
	}
	return
}

func address() (string, string, string) {
	addr := host()
	passwd := os.Getenv("DB_PASSWORD")
	uri := fmt.Sprintf("sslmode=disable dbname=example user=example password=%s host=%s", passwd, addr)
	return "postgres", uri, addr
}
