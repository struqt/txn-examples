package main

import (
	"context"
	"sync"
	"time"

	"github.com/struqt/txn"
	"github.com/struqt/txn/txn_sql"

	"examples/sqlc/pg/demo"
)

type DemoDoer interface {
	txn_sql.SqlDoer[*demo.Queries]
}

type DemoDoerBase struct {
	txn_sql.SqlDoerBase[*demo.Queries]
}

func (do *DemoDoerBase) BeginTxn(ctx context.Context, db txn_sql.SqlBeginner) (txn_sql.Txn, error) {
	if w, err := txn_sql.SqlBeginTxn(ctx, db, do.Options()); err != nil {
		return nil, err
	} else {
		if do.Stmt() == nil {
			do.SetStmt(demo.New(db))
		}
		do.SetStmt(do.Stmt().WithTx(w.Raw))
		return w, nil
	}
}

type DemoQueries struct {
	db    txn_sql.SqlBeginner
	mu    sync.Mutex
	cache *demo.Queries
}

func (dq *DemoQueries) Clear() {
	dq.mu.Lock()
	defer dq.mu.Unlock()
	if dq.cache != nil {
		_ = dq.cache.Close()
		dq.cache = nil
	}
}

func (dq *DemoQueries) Prepare(ctx context.Context, do DemoDoer) error {
	dq.mu.Lock()
	defer dq.mu.Unlock()
	if dq.cache == nil {
		var err error
		t0 := time.Now()
		log.V(2).Info("Preparing ...")
		dq.cache, err = demo.Prepare(ctx, dq.db)
		log.V(2).Info("Prepared", "duration", time.Now().Sub(t0))
		if err != nil {
			log.Error(err, "failed to prepare transaction")
			return err
		}
	}
	do.SetStmt(dq.cache)
	return nil
}

func DemoExecute[D DemoDoer](
	ctx context.Context, dq *DemoQueries, do D, fn txn.DoFunc[txn.Txn, txn_sql.SqlBeginner, D]) (D, error) {
	if err := dq.Prepare(ctx, do); err != nil {
		return do, err
	}
	t0 := time.Now()
	log.V(2).Info("Executing ...")
	if doer, err := txn_sql.SqlExecute(ctx, dq.db, do, fn); err != nil {
		dq.Clear()
		return doer, err
	} else {
		log.V(2).Info("Executed", "duration", time.Now().Sub(t0))
		return doer, nil
	}
}
