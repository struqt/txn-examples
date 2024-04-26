package dao

import (
	"context"
	"github.com/struqt/txn"
	t "github.com/struqt/txn/txn_mongo"
	"log/slog"
)

type (
	Txn         = txn.Txn
	TxnOptions  = t.Options
	TxnBeginner = t.Beginner
	TxnModule   = t.Module
)

type TxnDoer t.Doer[TxnBeginner]
type TxnDoerBase[Result any] struct {
	t.DoerBase[TxnBeginner]
	Result Result
}

type TxnModuleBase struct {
	t.ModuleBase
}

func (do *TxnDoerBase[_]) BeginTxn(ctx context.Context, db TxnBeginner) (Txn, error) {
	if w, err := TxnBegin(ctx, db, do.Options()); err != nil {
		return nil, err
	} else {
		do.SetClient(db)
		return w, nil
	}
}

func NewTxnModule(db TxnBeginner) TxnModule {
	i := &TxnModuleBase{}
	i.Mutate(t.WithBeginner(db))
	return i
}

func TxnPing(beginner TxnBeginner, count txn.PingCount) (int, error) {
	return t.Ping(beginner, 3, count)
}

func TxnBegin(ctx context.Context, db TxnBeginner, options TxnOptions) (t.RawTxn, error) {
	return t.BeginTxn(ctx, db, options)
}

func TxnExecute[Doer TxnDoer](
	ctx context.Context, mod TxnModule, do Doer,
	fn txn.DoFunc[TxnOptions, TxnBeginner, Doer], setters ...txn.DoerFieldSetter,
) (Doer, error) {
	ctx = context.WithValue(ctx, "logger", slog.Default())
	return t.Execute[TxnBeginner](ctx, mod, do, fn, setters...)
}
