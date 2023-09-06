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

type Txn = txn.Txn

type TxnSql = *txn_sql.SqlTxn

type TxnBeginner = txn_sql.SqlBeginner

type TxnOptions = txn_sql.SqlOptions

type TxnStmt interface {
	comparable
	io.Closer
}

type TxnDoer[Stmt TxnStmt] interface {
	txn_sql.SqlDoer[Stmt]
}

type TxnDoerBase[Stmt TxnStmt] struct {
	txn_sql.SqlDoerBase[Stmt]
}

func TxnBegin(ctx context.Context, db TxnBeginner, options TxnOptions) (TxnSql, error) {
	return txn_sql.SqlBeginTxn(ctx, db, options)
}

type Dao[Stmt TxnStmt] interface {
	io.Closer
	prepare(ctx context.Context, do TxnDoer[Stmt]) error
	beginner() TxnBeginner
}

type daoBase[Stmt TxnStmt] struct {
	mu       sync.Mutex
	db       TxnBeginner
	cache    Stmt
	cacheNew func(context.Context, TxnBeginner) (Stmt, error)
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

func (impl *daoBase[any]) beginner() TxnBeginner {
	return impl.db
}

func (impl *daoBase[any]) prepare(ctx context.Context, do TxnDoer[any]) (err error) {
	impl.mu.Lock()
	defer impl.mu.Unlock()
	var empty any
	if impl.cache == empty {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Second*3)
		defer cancel()
		impl.cache, err = impl.cacheNew(ctx, impl.db)
		if err != nil {
			do.SetStmt(empty)
			return
		}
	}
	do.SetStmt(impl.cache)
	return
}

func Execute[Stmt TxnStmt, Doer TxnDoer[Stmt]](
	ctx context.Context, dao Dao[Stmt], do Doer, fn txn.DoFunc[TxnOptions, TxnBeginner, Doer]) (Doer, error) {
	do.SetReadWrite(title(dao, do))
	return exec(ctx, dao, do, fn)
}

func ExecuteRo[Stmt TxnStmt, Doer TxnDoer[Stmt]](
	ctx context.Context, dao Dao[Stmt], do Doer, fn txn.DoFunc[TxnOptions, TxnBeginner, Doer]) (Doer, error) {
	do.SetReadOnly(title(dao, do))
	return exec(ctx, dao, do, fn)
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
	log.V(2).Info(" ~", "state", "Preparing")
	var x, err error
	var pings = 0
	var retries = -1
	t0 := time.Now()
retry:
	retries++
	if retries > doer.MaxRetry() && doer.MaxRetry() > 0 {
		if err != nil {
			log.Error(err, "", "retries", retries, "pings", pings)
		}
		return doer, err
	}
	err = dao.prepare(ctx, doer)
	if err != nil {
		pings, x = txn_sql.SqlPing[Stmt](ctx, dao.beginner(), doer, func(i time.Duration, cnt int) {
			log.Info("Ping", "retries", retries, "pings", cnt, "interval", i)
		})
		if x == nil && pings <= 1 {
			log.Error(err, "", "retries", retries, "pings", pings)
			return doer, err
		}
		log.Info("", "retries", retries, "pings", pings, "err", err)
		goto retry
	}
	t1 := time.Now()
	log.V(2).Info(" ~", "state", "Prepared", "duration", t1.Sub(t0))
	log.V(1).Info(" +")
	if _, err = txn_sql.SqlExecute(ctx, dao.beginner(), doer, fn); err == nil {
		log.V(1).Info(" +", "duration", time.Now().Sub(t1))
		return doer, nil
	}
	if x := dao.Close(); x != nil {
		log.Error(x, "")
		err = fmt.Errorf("%w [exec] %w [Close]", err, x)
	} else {
		err = fmt.Errorf("%w [exec]", err)
	}
	pings, x = txn_sql.SqlPing[Stmt](ctx, dao.beginner(), doer, func(i time.Duration, cnt int) {
		log.Info("Ping", "retries", retries, "pings", cnt, "interval", i)
	})
	if x == nil && pings <= 1 {
		log.Error(err, "", "retries", retries, "pings", pings)
		return doer, err
	}
	log.Info("", "retries", retries, "pings", pings, "err", err)
	goto retry
}
