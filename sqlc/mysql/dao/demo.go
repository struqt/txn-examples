package dao

import (
	"context"
	"fmt"
	"reflect"

	"examples/sqlc/mysql/demo"
)

type DemoStmt = *demo.Queries

type DemoDoer[Result any] struct {
	TxnDoerBase[DemoStmt, Result]
}

func (do *DemoDoer[_]) BeginTxn(ctx context.Context, db TxnBeginner) (Txn, error) {
	if w, err := TxnBegin(ctx, db, do.Options()); err != nil {
		return nil, err
	} else {
		if do.Stmt() == nil {
			do.SetStmt(demo.New(db))
		}
		do.SetStmt(do.Stmt().WithTx(w.Raw))
		return w, nil
	}
}

type Demo = TxnModule[DemoStmt]

func NewDemo(db TxnBeginner) Demo {
	i := &TxnModuleBase[DemoStmt]{}
	i.Init(db, func(ctx context.Context, db TxnBeginner) (DemoStmt, error) {
		return demo.Prepare(ctx, db)
	})
	return i
}

type ListAuthor struct {
	DemoDoer[any]
	len int
}

func ListAuthorDo(ctx context.Context, do *ListAuthor) error {
	log := log.WithName(do.Title())
	authors, err := do.Stmt().ListAuthors(ctx)
	if err != nil {
		return err
	}
	do.len = len(authors)
	log.V(2).Info(" |", "len", do.len)
	return nil
}

type LastAuthor struct {
	DemoDoer[any]
	id int64
}

func LastAuthorDo(ctx context.Context, do *LastAuthor) error {
	log := log.WithName(do.Title())
	stat, err := do.Stmt().StatAuthor(ctx)
	if err != nil {
		return err
	}
	log.V(2).Info(" |", "stat", stat)
	if stat.Size <= 0 {
		return nil
	}
	if id, ok := stat.MaxID.(int64); ok {
		fetched, err := do.Stmt().GetAuthor(ctx, id)
		do.id = id
		if err != nil {
			return err
		}
		log.V(2).Info(" |", "fetched.id", fetched.ID, "name", fetched.Name, "bio", fetched.Bio.String)
	} else {
		return fmt.Errorf("the value is not of type int64")
	}
	//panic("fake panic")
	//return fmt.Errorf("fake error")
	return nil
}

type PushAuthor struct {
	DemoDoer[any]
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
	do.inserted, _ = inserted.LastInsertId()
	log.V(2).Info("|", "inserted", do.inserted)
	fetched, err := do.Stmt().GetAuthor(ctx, do.inserted)
	if err != nil {
		return err
	}
	log.V(2).Info("|", "equals", reflect.DeepEqual(do.inserted, fetched.ID))
	count := 1
	for {
		if count > 10 {
			break
		}
		stat, err := do.Stmt().StatAuthor(ctx)
		if err != nil {
			return err
		}
		log.V(2).Info("|", "stat", stat)
		if stat.Size <= 10 {
			break
		}
		if id, ok := stat.MinID.(int64); ok {
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
