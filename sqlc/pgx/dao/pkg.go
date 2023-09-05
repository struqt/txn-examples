package dao

import (
	"context"
	"fmt"
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

type Statements = any

type TxnDoer[Stmt Statements] interface {
	txn_pgx.PgxDoer[Stmt]
}

type TxnDoerBase[Stmt Statements] struct {
	txn_pgx.PgxDoerBase[Stmt]
}

func TxnBegin(ctx context.Context, db TxnBeginner, options TxnOptions) (TxnPgx, error) {
	return txn_pgx.PgxBeginTxn(ctx, db, options)
}

type Dao[Stmt Statements] interface {
	beginner() TxnBeginner
}

type daoBase[Stmt Statements] struct {
	db TxnBeginner
}

func (impl *daoBase[any]) beginner() TxnBeginner {
	return impl.db
}

func Execute[Stmt Statements, Doer TxnDoer[Stmt]](
	ctx context.Context, dao Dao[Stmt], doer Doer, fn txn.DoFunc[TxnOptions, TxnBeginner, Doer]) (Doer, error) {
	doer.SetReadWrite(title(dao, doer))
	return exec(ctx, dao, doer, fn)
}

func ExecuteRo[Stmt Statements, Doer TxnDoer[Stmt]](
	ctx context.Context, dao Dao[Stmt], doer Doer, fn txn.DoFunc[TxnOptions, TxnBeginner, Doer]) (Doer, error) {
	doer.SetReadOnly(title(dao, doer))
	return exec(ctx, dao, doer, fn)
}

func exec[Stmt Statements, Doer TxnDoer[Stmt]](
	ctx context.Context, dao Dao[Stmt], doer Doer, fn txn.DoFunc[TxnOptions, TxnBeginner, Doer]) (Doer, error) {
	log := log.WithName(doer.Title())
	t1 := time.Now()
	log.V(1).Info("  +")
	if doer, err := txn_pgx.PgxExecute(ctx, dao.beginner(), doer, fn); err != nil {
		log.Error(err, "")
		return doer, fmt.Errorf("%w [PgxExecute]", err)
	} else {
		log.V(1).Info("  +", "duration", time.Now().Sub(t1))
		return doer, nil
	}
}

func title[Stmt Statements, Doer TxnDoer[Stmt]](_ Dao[Stmt], do Doer) string {
	if do.Title() != "" {
		return ""
	}
	t := reflect.TypeOf(do)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}
