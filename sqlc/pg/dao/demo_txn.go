package dao

import (
	"context"
	"database/sql"

	"github.com/struqt/txn/txn_sql"

	"examples/sqlc/pg/demo"
)

type DemoQueries = *demo.Queries

type DemoDoer = txn_sql.SqlDoer[DemoQueries]

type DemoDoerBase struct {
	txn_sql.SqlDoerBase[DemoQueries]
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

type Demo = Dao[DemoQueries]

func NewDemo(db *sql.DB) Demo {
	i := &daoBase[DemoQueries]{}
	i.db = db
	i.cacheNew = func(ctx context.Context, db txn_sql.SqlBeginner) (DemoQueries, error) {
		return demo.Prepare(ctx, db)
	}
	return i
}
