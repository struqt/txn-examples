package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/struqt/logging"
	"github.com/struqt/txn"
	"go.mongodb.org/mongo-driver/bson"

	"examples/mongo/dao"
)

func init() {
	logging.LogConsoleThreshold = 2
	slog.SetDefault(logging.DefaultLogger)
}

func main() {
	slog.Info("Process is starting ...")
	defer os.Exit(0)
	defer slog.Info("Process is ending ...")
	ctx, cancel := context.WithCancel(context.Background())
	defer slog.Info("Context is cancelled")
	defer cancel()
	defer dao.TearDown()
	dao.Setup()
	do(ctx, 0)
	txn.RunTicker(ctx, 300*time.Millisecond, 5, do)
}

func do(ctx context.Context, tick int32) {
	mod := dao.Demo()
	slog.Info(fmt.Sprintf("tick %d", tick))
	i := map[string]any{"name": "Brian Kernighan", "age": 30, "createdAt": time.Now()}
	_, _ = dao.TxnExecute(ctx, mod, &dao.ListAuthor{}, dao.ListAuthorDo)
	_, _ = dao.TxnExecute(ctx, mod, &dao.PushAuthor{Insert: i}, dao.PushAuthorDo)
	_, _ = dao.TxnExecute(ctx, mod, &dao.ListAuthor{}, dao.ListAuthorDo)
	_, _ = stat(ctx)
}

func stat(ctx context.Context) (int64, error) {
	mod := dao.Demo()
	type doer = dao.DemoDoer[int64]
	d, err := dao.TxnExecute(ctx, mod, &doer{}, func(ctx context.Context, do *doer) error {
		client := do.Client()
		collection := client.Database("demo").Collection("authors")
		total, err := collection.CountDocuments(ctx, bson.D{})
		if err != nil {
			return err
		}
		do.Result = total
		slog.With("T", do.Title()).Info(":", "result", do.Result)
		return nil
	}, txn.WithTitle("Txn`AdHocStat"))
	return d.Result, err
}
