package dao

import (
	"context"
	"fmt"
	"io"
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

type Statements interface {
	comparable
	io.Closer
}

type Dao[Stmt Statements] interface {
	io.Closer
	prepare(ctx context.Context, do txn_sql.SqlDoer[Stmt]) error
	beginner() txn_sql.SqlBeginner
}

type daoBase[Stmt Statements] struct {
	mu       sync.Mutex
	db       txn_sql.SqlBeginner
	cache    Stmt
	cacheNew func(context.Context, txn_sql.SqlBeginner) (Stmt, error)
}

func (impl *daoBase[any]) Close() error {
	impl.mu.Lock()
	defer impl.mu.Unlock()
	var empty any
	if impl.cache != empty {
		defer func() { impl.cache = empty }()
		return impl.cache.Close()
	}
	return nil
}

func (impl *daoBase[any]) beginner() txn_sql.SqlBeginner {
	return impl.db
}

func (impl *daoBase[any]) prepare(ctx context.Context, do txn_sql.SqlDoer[any]) (err error) {
	impl.mu.Lock()
	defer impl.mu.Unlock()
	var empty any
	if impl.cache == empty {
		t0 := time.Now()
		log.V(2).Info("Preparing ...")
		if do.Timeout() > time.Millisecond {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, do.Timeout())
			defer log.V(2).Info("Prepared, Canceled")
			defer cancel()
		}
		impl.cache, err = impl.cacheNew(ctx, impl.db)
		log.V(2).Info("Prepared", "duration", time.Now().Sub(t0))
		if err != nil {
			log.Error(err, "failed to prepare transaction")
			return
		}
	}
	do.SetStmt(impl.cache)
	return
}

func ExecuteRo[Stmt Statements, Doer txn_sql.SqlDoer[Stmt]](
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

func ExecuteRw[Stmt Statements, Doer txn_sql.SqlDoer[Stmt]](
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

func Execute[Stmt Statements, Doer txn_sql.SqlDoer[Stmt]](
	ctx context.Context, dt Dao[Stmt],
	do Doer, fn txn.DoFunc[txn.Txn, txn_sql.SqlBeginner, Doer],
) (Doer, error) {
	log := log.WithName(do.Title())
	if err := dt.prepare(ctx, do); err != nil {
		log.Error(err, "")
		return do, err
	}
	t0 := time.Now()
	log.V(1).Info("+")
	if doer, err := txn_sql.SqlExecute(ctx, dt.beginner(), do, fn); err != nil {
		log.Error(err, "")
		if x := dt.Close(); x != nil {
			log.Error(x, "")
			return doer, fmt.Errorf("%w [SqlExecute] %w [Close]", err, x)
		} else {
			return doer, fmt.Errorf("%w [SqlExecute]", err)
		}
	} else {
		log.V(1).Info("-", "cost", time.Now().Sub(t0))
		return doer, nil
	}
}
