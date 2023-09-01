package dao

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/struqt/txn/txn_sql"

	"examples/sqlc/mysql/demo"
)

type Demo = Dao[*demo.Queries]

type DemoDoer = txn_sql.SqlDoer[*demo.Queries]

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

func NewDemo(db *sql.DB) Demo {
	return &DemoImpl{db: db}
}

type DemoImpl struct {
	db    txn_sql.SqlBeginner
	mu    sync.Mutex
	cache *demo.Queries
}

func (impl *DemoImpl) beginner() txn_sql.SqlBeginner {
	return impl.db
}

func (impl *DemoImpl) prepare(ctx context.Context, do DemoDoer) (err error) {
	impl.mu.Lock()
	defer impl.mu.Unlock()
	if impl.cache == nil {
		t0 := time.Now()
		log.V(2).Info("Preparing ...")
		if do.Timeout() > time.Millisecond {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, do.Timeout())
			defer log.V(2).Info("Prepared, Canceled")
			defer cancel()
		}
		impl.cache, err = demo.Prepare(ctx, impl.db)
		log.V(2).Info("Prepared", "duration", time.Now().Sub(t0))
		if err != nil {
			log.Error(err, "failed to prepare transaction")
			return
		}
	}
	do.SetStmt(impl.cache)
	return
}

func (impl *DemoImpl) Clear() {
	impl.mu.Lock()
	defer impl.mu.Unlock()
	if impl.cache != nil {
		_ = impl.cache.Close()
		impl.cache = nil
	}
}
