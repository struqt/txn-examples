package main

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/struqt/logging"
	"github.com/struqt/txn"

	"examples/sqlc/sqlite/dao"
	"examples/sqlc/sqlite/dao/demo"
)

//go:embed schema.sql
var ddl string

var log logging.Logger

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
	dao.Setup(log, ddl)
	do(ctx, 0)
	txn.RunTicker(ctx, 300*time.Millisecond, 5, do)
}

func do(ctx context.Context, count int32) {
	d := dao.Demo()
	log.V(0).Info(fmt.Sprintf("tick %d", count))
	_, _ = dao.TxnRwExecute(ctx, d, push(), dao.PushAuthorDo)
	_, _ = dao.TxnRoExecute(ctx, d, &dao.ListAuthor{}, dao.ListAuthorDo)
	_, _ = dao.TxnRoExecute(ctx, d, &dao.LastAuthor{}, dao.LastAuthorDo)
}

func push() (do *dao.PushAuthor) {
	do = &dao.PushAuthor{}
	do.Insert = demo.CreateAuthorParams{
		Name: "Brian Kernighan",
		Bio: sql.NullString{
			Valid:  true,
			String: "Co-author of The C Programming Language",
		}}
	return
}

//
//func run() error {
//	ctx := context.Background()
//
//	db, err := sql.Open("sqlite3", ":memory:")
//	if err != nil {
//		return err
//	}
//
//	// create tables
//	if _, err := db.ExecContext(ctx, ddl); err != nil {
//		return err
//	}
//
//	//queries := tutorial.New(db)
//	//
//	//// list all authors
//	//authors, err := queries.ListAuthors(ctx)
//	//if err != nil {
//	//	return err
//	//}
//	//log.Println(authors)
//	//
//	//// create an author
//	//insertedAuthor, err := queries.CreateAuthor(ctx, tutorial.CreateAuthorParams{
//	//	Name: "Brian Kernighan",
//	//	Bio:  sql.NullString{String: "Co-author of The C Programming Language and The Go Programming Language", Valid: true},
//	//})
//	//if err != nil {
//	//	return err
//	//}
//	//log.Println(insertedAuthor)
//	//
//	//// get the author we just inserted
//	//fetchedAuthor, err := queries.GetAuthor(ctx, insertedAuthor.ID)
//	//if err != nil {
//	//	return err
//	//}
//	//
//	//// prints true
//	//log.Println(reflect.DeepEqual(insertedAuthor, fetchedAuthor))
//
//	return nil
//}
//
//func main() {
//	if err := run(); err != nil {
//		return
//	}
//}
