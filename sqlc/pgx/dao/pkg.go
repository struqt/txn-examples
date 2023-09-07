package dao

import (
	"context"
	"sync"

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
type TxnOptions = txn_pgx.PgxOptions
type TxnBeginner = txn_pgx.PgxBeginner
type TxnStmt txn_pgx.PgxStmt

type TxnDoer[Stmt TxnStmt] txn_pgx.PgxDoer[Stmt]
type TxnDoerBase[Stmt TxnStmt, R any] struct {
	txn_pgx.PgxDoerBase[Stmt]
	Result R
}

type TxnModule[Stmt TxnStmt] txn_pgx.PgxModule[Stmt]
type TxnModuleBase[Stmt TxnStmt] struct {
	txn_pgx.PgxModuleBase[Stmt]
}

func TxnPing(ctx context.Context, beginner TxnBeginner, count txn.PingCount) (int, error) {
	return txn_pgx.PgxPing(ctx, beginner, 4, count)
}

func TxnBegin(ctx context.Context, db TxnBeginner, options TxnOptions) (TxnPgx, error) {
	return txn_pgx.PgxBeginTxn(ctx, db, options)
}

func TxnRwExecute[Stmt TxnStmt, Doer TxnDoer[Stmt]](
	ctx context.Context, mod TxnModule[Stmt], do Doer,
	fn txn.DoFunc[TxnOptions, TxnBeginner, Doer],
) (Doer, error) {
	return txn_pgx.PgxRwExecute[Stmt, Doer](ctx, log, mod, do, fn)
}

func TxnRoExecute[Stmt TxnStmt, Doer TxnDoer[Stmt]](
	ctx context.Context, mod TxnModule[Stmt], do Doer,
	fn txn.DoFunc[TxnOptions, TxnBeginner, Doer],
) (Doer, error) {
	return txn_pgx.PgxRoExecute[Stmt](ctx, log, mod, do, fn)
}
