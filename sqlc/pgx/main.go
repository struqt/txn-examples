package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/struqt/logging"
	"github.com/struqt/txn"

	"examples/sqlc/pgx/dao"
	"examples/sqlc/pgx/dao/demo"
)

var log *slog.Logger

func init() {
	logging.LogConsoleThreshold = -128
	log = logging.NewLogger("")
}

func main() {
	log.Info("Process is starting ...")
	defer os.Exit(0)
	defer log.Info("Process is ending ...")
	ctx, cancel := context.WithCancel(context.Background())
	defer log.Info("Context is cancelled")
	defer cancel()
	defer dao.TearDown()
	dao.Setup(log)
	do(ctx, 0)
	txn.RunTicker(ctx, 300*time.Millisecond, 5, do)
}

func do(ctx context.Context, count int32) {
	mod := dao.Demo()
	log.Info(fmt.Sprintf("tick %d", count))
	_, _ = stat(ctx, mod)
	_, _ = dao.TxnRwExecute(ctx, mod, push(), dao.PushAuthorDo)
	_, _ = dao.TxnRoExecute(ctx, mod, &dao.DemoDoer[[]demo.Author]{}, dao.ListAuthorDo, txn.WithTitle("TxnRo`List"))
	_, _ = dao.TxnRoExecute(ctx, mod, &dao.LastAuthor{}, dao.LastAuthorDo)
}

func stat(ctx context.Context, mod dao.DemoModule) (*demo.StatAuthorRow, error) {
	type _doer = dao.DemoDoer[demo.StatAuthorRow]
	do, err := dao.TxnRoExecute(ctx, mod, &_doer{}, func(ctx context.Context, do *_doer) error {
		if result, err := do.Stmt().StatAuthor(ctx); err != nil {
			return err
		} else {
			do.Result = result
			log.With("T", do.Title()).Info("     :", "result", do.Result)
			return nil
		}
	}, txn.WithTitle("TxnRo`Stat"))
	return &do.Result, err
}

func push() *dao.PushAuthor {
	do := &dao.PushAuthor{}
	do.Insert = demo.CreateAuthorParams{
		Name: "Brian Kernighan",
		Bio: pgtype.Text{
			Valid:  true,
			String: "Co-author of The C Programming Language",
		},
	}
	return do
}
