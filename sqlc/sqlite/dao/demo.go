package dao

import (
	"context"
	"fmt"
	"reflect"

	"examples/sqlc/sqlite/dao/demo"
)

type DemoQueries = *demo.Queries

type DemoDoer[Result any] struct {
	TxnDoerBase[DemoQueries, Result]
}

func (do *DemoDoer[R]) BeginTxn(ctx context.Context, db TxnBeginner) (Txn, error) {
	if w, err := TxnBegin(ctx, db, do.Options()); err != nil {
		return nil, err
	} else {
		if do.Stmt() == nil {
			do.SetStmt(demo.New(db))
		}
		do.SetStmt(do.Stmt().WithTx(w.Raw()))
		return w, nil
	}
}

type DemoModule = TxnModule[DemoQueries]

func NewDemo(db TxnBeginner) DemoModule {
	mod := &TxnModuleBase[DemoQueries]{}
	mod.Init(db, func(ctx context.Context, db TxnBeginner) (DemoQueries, error) {
		return demo.Prepare(ctx, db)
	})
	return mod
}

type ListAuthor struct {
	DemoDoer[DemoQueries]
	len int
}

func ListAuthorDo(ctx context.Context, do *ListAuthor) error {
	log := log.WithName(do.Title())
	authors, err := do.Stmt().ListAuthors(ctx)
	if err != nil {
		return err
	}
	do.len = len(authors)
	log.V(2).Info(" :", "len", do.len)
	return nil
}

type LastAuthor struct {
	DemoDoer[DemoQueries]
	id int64
}

func LastAuthorDo(ctx context.Context, do *LastAuthor) error {
	log := log.WithName(do.Title())
	stat, err := do.Stmt().StatAuthor(ctx)
	if err != nil {
		return err
	}
	log.V(2).Info(" :", "stat", stat)
	if stat.Count <= 0 {
		return nil
	}
	if id, ok := stat.Max.(int64); ok {
		fetched, err := do.Stmt().GetAuthor(ctx, id)
		do.id = id
		if err != nil {
			return err
		}
		log.V(2).Info(" :", "fetched.id", fetched.ID, "name", fetched.Name, "bio", fetched.Bio.String)
	} else {
		return fmt.Errorf("the value is not of type int64")
	}
	//panic("fake panic")
	//return fmt.Errorf("fake error")
	return nil
}

type PushAuthor struct {
	DemoDoer[DemoQueries]
	Insert   demo.CreateAuthorParams
	inserted int64
}

func PushAuthorDo(ctx context.Context, do *PushAuthor) error {
	log := log.WithName(do.Title())
	var err error
	inserted, err := do.Stmt().CreateAuthor(ctx, do.Insert)
	if err != nil {
		return err
	}
	do.inserted = inserted.ID
	log.V(2).Info(":", "inserted.id", inserted.ID, "name", inserted.Name, "bio", inserted.Bio.String)
	fetched, err := do.Stmt().GetAuthor(ctx, inserted.ID)
	if err != nil {
		return err
	}
	log.V(2).Info(":", "equals", reflect.DeepEqual(inserted, fetched))
	count := 1
	for {
		if count > 10 {
			break
		}
		stat, err := do.Stmt().StatAuthor(ctx)
		if err != nil {
			return err
		}
		log.V(2).Info(":", "stat", stat)
		if stat.Count <= 5 {
			break
		}
		if id, ok := stat.Min.(int64); ok {
			if err = do.Stmt().DeleteAuthor(ctx, id); err != nil {
				return err
			}
			count++
		} else {
			return fmt.Errorf("the value is not of type int64")
		}
	}
	//panic("fake panic")
	//return fmt.Errorf("fake error")
	return nil
}
