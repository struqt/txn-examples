package dao

import (
	"context"
	"github.com/struqt/txn"
	t "github.com/struqt/txn/txn_mongo"
)

type Txn = txn.Txn
type TxnImpl = *t.Txn
type TxnOptions = t.Options
type TxnBeginner = t.Beginner
type TxnStmt t.StmtHolder

type TxnDoer[Stmt TxnStmt] t.Doer[Stmt]
type TxnDoerBase[Stmt TxnStmt, Result any] struct {
	t.DoerBase[Stmt]
	Result Result
}

type TxnModule[Stmt TxnStmt] t.Module[Stmt]
type TxnModuleBase[Stmt TxnStmt] struct {
	t.ModuleBase[Stmt]
}

func TxnPing(beginner TxnBeginner, count txn.PingCount) (int, error) {
	return t.Ping(beginner, 3, count)
}

func TxnBegin(ctx context.Context, db TxnBeginner, options TxnOptions) (TxnImpl, error) {
	return t.BeginTxn(ctx, db, options)
}

func TxnExecute[Stmt TxnStmt, Doer TxnDoer[Stmt]](
	ctx context.Context, mod TxnModule[Stmt], do Doer,
	fn txn.DoFunc[TxnOptions, TxnBeginner, Doer], setters ...txn.DoerFieldSetter,
) (Doer, error) {
	c := context.WithValue(ctx, "logger", log)
	return t.Execute[Stmt](c, mod, do, fn, setters...)
}
