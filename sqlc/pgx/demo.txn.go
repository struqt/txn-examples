package main

import (
	"context"

	"github.com/struqt/txn/txn_pgx"

	"examples/sqlc/pgx/demo"
)

type DemoDoerBase struct {
	txn_pgx.PgxDoerBase[*demo.Queries]
}

func (do *DemoDoerBase) BeginTxn(ctx context.Context, db txn_pgx.PgxBeginner) (txn_pgx.Txn, error) {
	if w, err := txn_pgx.PgxBeginTxn(ctx, db, do.Options()); err != nil {
		return nil, err
	} else {
		do.SetStmt(demo.New(w.Raw))
		return w, nil
	}
}
