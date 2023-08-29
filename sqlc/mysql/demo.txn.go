package main

import (
	"context"

	"github.com/struqt/txn/txn_sql"

	"examples/sqlc/mysql/demo"
)

type DemoDoerBase struct {
	txn_sql.SqlDoerBase
	query *demo.Queries
}

func (d *DemoDoerBase) BeginTxn(ctx context.Context, db txn_sql.SqlBeginner) (txn_sql.Txn, error) {
	if w, err := txn_sql.SqlBeginTxn(ctx, db, d.Options()); err != nil {
		return nil, err
	} else {
		if d.query == nil {
			d.query = demo.New(db)
		}
		d.query = d.query.WithTx(w.Raw)
		return w, nil
	}
}
