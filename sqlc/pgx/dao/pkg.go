package dao

import (
	"context"
	"reflect"
	"sync"
	"time"

	"github.com/struqt/logging"
	"github.com/struqt/txn"
	"github.com/struqt/txn/txn_pgx"
)

var (
	once sync.Once
	log  logging.Logger
)

func Setup(logger logging.Logger) {
	once.Do(func() {
		log = logger
	})
}

type Txn = txn.Txn

type TxnPgx = *txn_pgx.PgxTxn

type TxnBeginner = txn_pgx.PgxBeginner

type TxnOptions = txn_pgx.PgxOptions

type TxnStmt = any

type TxnDoer[Stmt TxnStmt] interface {
	txn_pgx.PgxDoer[Stmt]
}

type TxnDoerBase[Stmt TxnStmt] struct {
	txn_pgx.PgxDoerBase[Stmt]
}

func TxnBegin(ctx context.Context, db TxnBeginner, options TxnOptions) (TxnPgx, error) {
	return txn_pgx.PgxBeginTxn(ctx, db, options)
}

type Dao[Stmt TxnStmt] interface {
	beginner() TxnBeginner
}

type daoBase[Stmt TxnStmt] struct {
	db TxnBeginner
}

func (d *daoBase[_]) beginner() TxnBeginner {
	return d.db
}

func Execute[Stmt TxnStmt, Doer TxnDoer[Stmt]](
	ctx context.Context, dao Dao[Stmt], doer Doer, fn txn.DoFunc[TxnOptions, TxnBeginner, Doer]) (Doer, error) {
	doer.SetReadWrite(title(dao, doer))
	return exec(ctx, dao, doer, fn)
}

func ExecuteRo[Stmt TxnStmt, Doer TxnDoer[Stmt]](
	ctx context.Context, dao Dao[Stmt], doer Doer, fn txn.DoFunc[TxnOptions, TxnBeginner, Doer]) (Doer, error) {
	doer.SetReadOnly(title(dao, doer))
	return exec(ctx, dao, doer, fn)
}

func title[Stmt TxnStmt, Doer TxnDoer[Stmt]](_ Dao[Stmt], do Doer) string {
	if do.Title() != "" {
		return ""
	}
	t := reflect.TypeOf(do)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

func exec[Stmt TxnStmt, Doer TxnDoer[Stmt]](
	ctx context.Context, dao Dao[Stmt], doer Doer, fn txn.DoFunc[TxnOptions, TxnBeginner, Doer]) (Doer, error) {
	log := log.WithName(doer.Title())
	log.V(1).Info("  +")
	var x, err error
	var pings int
	var retries = -1
	t1 := time.Now()
retry:
	retries++
	if retries > doer.MaxRetry() && doer.MaxRetry() > 0 {
		if err != nil {
			log.Error(err, "", "retries", retries, "pings", pings)
		}
		return doer, err
	}
	if doer, err = txn_pgx.PgxExecute(ctx, dao.beginner(), doer, fn); err == nil {
		log.V(1).Info("  +", "duration", time.Now().Sub(t1))
		return doer, nil
	}

	pings, x = txn_pgx.PgxPing[Stmt](ctx, dao.beginner(), doer, func(i time.Duration, cnt int) {
		log.Info("PgxPing", "retries", retries, "pings", cnt, "interval", i)
	})
	if x == nil && pings <= 1 {
		log.Error(err, "", "retries", retries, "pings", pings)
		return doer, err
	}
	log.Info("", "retries", retries, "pings", pings, "err", err)
	goto retry
}
