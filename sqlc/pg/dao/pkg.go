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
	prepare(ctx context.Context, do txn_sql.SqlDoer[Stmt]) error
	beginner() txn_sql.SqlBeginner
}

type daoBase[Stmt TxnStmt] struct {
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
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Second*3)
		defer cancel()
		impl.cache, err = impl.cacheNew(ctx, impl.db)
		if err != nil {
			do.SetStmt(empty)
			return
		} else {

		}
	}
	do.SetStmt(impl.cache)
	return
}

func Execute[Stmt TxnStmt, Doer TxnDoer[Stmt]](
	ctx context.Context, dao Dao[Stmt], do Doer, fn txn.DoFunc[TxnOptions, TxnBeginner, Doer],
) (Doer, error) {
	do.SetReadWrite(title(dao, do))
	return exec(ctx, dao, do, fn)
}

func ExecuteRo[Stmt TxnStmt, Doer TxnDoer[Stmt]](
	ctx context.Context, dao Dao[Stmt], do Doer, fn txn.DoFunc[TxnOptions, TxnBeginner, Doer],
) (Doer, error) {
	do.SetReadOnly(title(dao, do))
	return exec(ctx, dao, do, fn)
}

func exec[Stmt TxnStmt, Doer TxnDoer[Stmt]](
	ctx context.Context, dao Dao[Stmt], doer Doer, fn txn.DoFunc[TxnOptions, TxnBeginner, Doer],
) (Doer, error) {
	log := log.WithName(doer.Title())
	retryIntervals := [4]time.Duration{
		time.Second * 1,
		time.Second * 4,
		time.Second * 9,
		time.Second * 16,
	}
	const retryIntervalsLen = len(retryIntervals)
	t0 := time.Now()
	cnt := 0
	log.V(2).Info(" ~", "state", "Preparing")
	for {
		var err error
		if err = dao.prepare(ctx, doer); err == nil {
			break
		}
		if doer.IsReadOnly() && cnt > retryIntervalsLen {
			log.Error(err, "")
			return doer, err
		}
		i := retryIntervals[cnt%retryIntervalsLen]
		cnt++
		log.Info("", "e", err)
		log.Info("Retrying", "count", cnt, "interval", i)
		time.Sleep(i)
	}
	t1 := time.Now()
	log.V(2).Info(" ~", "state", "Prepared", "duration", t1.Sub(t0))
	log.V(1).Info(" +")
	if doer, err := txn_sql.SqlExecute(ctx, dao.beginner(), doer, fn); err != nil {
		log.Error(err, "")
		if x := dao.Close(); x != nil {
			log.Error(x, "")
			return doer, fmt.Errorf("%w [SqlExecute] %w [Close]", err, x)
		} else {
			return doer, fmt.Errorf("%w [SqlExecute]", err)
		}
	} else {
		log.V(1).Info(" +", "duration", time.Now().Sub(t1))
		return doer, nil
	}
}

func title[Stmt TxnStmt, Doer txn_sql.SqlDoer[Stmt]](_ Dao[Stmt], do Doer) string {
	if do.Title() != "" {
		return ""
	}
	t := reflect.TypeOf(do)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}
