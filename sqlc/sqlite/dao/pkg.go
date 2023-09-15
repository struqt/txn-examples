package dao

import (
	"context"
	"database/sql"
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

func Setup(logger logging.Logger, ddl string) {
	once.Do(func() {
		log = logger
		initDemo(ddl)
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

func initDemo(ddl string) {
	driver, uri := "sqlite3", os.Getenv("DB_ADDR_PATH")
	var (
		err error
		db  *sql.DB
	)
	if err, db = open(driver, uri); err != nil {
		log.Error(err, "Failed to set up connection pool")
		return
	}
	if _, err := db.ExecContext(context.Background(), ddl); err != nil {
		log.Error(err, "Failed to set up DDL")
		return
	}
	ping(db, uri)
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
