package dao

import (
	"context"
	"reflect"
	"sync"
	"time"

	"github.com/struqt/logging"
	"github.com/struqt/txn"
	"github.com/struqt/txn/txn_sql"
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

type Dao[Stmt any] interface {
	Clear()
	prepare(ctx context.Context, do txn_sql.SqlDoer[Stmt]) error
	beginner() txn_sql.SqlBeginner
}

func ExecuteRo[Stmt any, Doer txn_sql.SqlDoer[Stmt]](
	ctx context.Context, dao Dao[Stmt],
	do Doer, fn txn.DoFunc[txn.Txn, txn_sql.SqlBeginner, Doer],
) (Doer, error) {
	t := reflect.TypeOf(do)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	title := t.Name()
	do.ReadOnly(title)
	return Execute(ctx, dao, do, fn)
}

func ExecuteRw[Stmt any, Doer txn_sql.SqlDoer[Stmt]](
	ctx context.Context, dao Dao[Stmt],
	do Doer, fn txn.DoFunc[txn.Txn, txn_sql.SqlBeginner, Doer],
) (Doer, error) {
	t := reflect.TypeOf(do)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	title := t.Name()
	do.ReadWrite(title)
	return Execute(ctx, dao, do, fn)
}

func Execute[Stmt any, Doer txn_sql.SqlDoer[Stmt]](
	ctx context.Context, dt Dao[Stmt],
	do Doer, fn txn.DoFunc[txn.Txn, txn_sql.SqlBeginner, Doer],
) (Doer, error) {
	log := log.WithName(do.Title())
	if err := dt.prepare(ctx, do); err != nil {
		log.Error(err, "")
		return do, err
	}
	t0 := time.Now()
	log.V(1).Info(" +")
	if doer, err := txn_sql.SqlExecute(ctx, dt.beginner(), do, fn); err != nil {
		log.Error(err, "")
		dt.Clear()
		return doer, err
	} else {
		log.V(1).Info(" -", "cost", time.Now().Sub(t0))
		return doer, nil
	}
}
