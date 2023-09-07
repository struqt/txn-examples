package dao

import (
	"context"
	"sync"

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
type TxnOptions = txn_sql.SqlOptions
type TxnBeginner = txn_sql.SqlBeginner
type TxnStmt txn_sql.SqlStmt

type TxnDoer[Stmt TxnStmt] txn_sql.SqlDoer[Stmt]
type TxnDoerBase[Stmt TxnStmt] struct {
	txn_sql.SqlDoerBase[Stmt]
}

type TxnModule[Stmt TxnStmt] txn_sql.SqlModule[Stmt]
type TxnModuleBase[Stmt TxnStmt] struct {
	txn_sql.SqlModuleBase[Stmt]
}

func TxnPing(ctx context.Context, beginner TxnBeginner, count txn.PingCount) (int, error) {
	return txn_sql.SqlPing(ctx, beginner, 4, count)
}

func TxnBegin(ctx context.Context, db TxnBeginner, options TxnOptions) (TxnSql, error) {
	return txn_sql.SqlBeginTxn(ctx, db, options)
}

func TxnRwExecute[Stmt TxnStmt, Doer TxnDoer[Stmt]](
	ctx context.Context, mod TxnModule[Stmt], do Doer,
	fn txn.DoFunc[TxnOptions, TxnBeginner, Doer],
) (Doer, error) {
	return txn_sql.SqlRwExecute[Stmt, Doer](ctx, log, mod, do, fn)
}

func TxnRoExecute[Stmt TxnStmt, Doer TxnDoer[Stmt]](
	ctx context.Context, mod TxnModule[Stmt], do Doer,
	fn txn.DoFunc[TxnOptions, TxnBeginner, Doer],
) (Doer, error) {
	return txn_sql.SqlRoExecute[Stmt](ctx, log, mod, do, fn)
}
