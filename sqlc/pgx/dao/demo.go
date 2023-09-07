package dao

import (
	"context"
	"fmt"
	"reflect"

	"examples/sqlc/pgx/dao/demo"
)

type DemoQueries = *demo.Queries

type DemoDoer[Result any] struct {
	TxnDoerBase[DemoQueries, Result]
}

func (do *DemoDoer[R]) BeginTxn(ctx context.Context, db TxnBeginner) (Txn, error) {
	if w, err := TxnBegin(ctx, db, do.Options()); err != nil {
		return nil, err
	} else {
		do.SetStmt(demo.New(w.Raw))
		return w, nil
	}
}

type Demo = TxnModule[DemoQueries]

func NewDemo(db TxnBeginner) Demo {
	i := &TxnModuleBase[DemoQueries]{}
	i.Init(db)
	return i
}

func ListAuthorDo(ctx context.Context, do *DemoDoer[[]demo.Author]) error {
	log := log.WithName(do.Title())
	authors, err := do.Stmt().ListAuthors(ctx)
	if err != nil {
		return err
	}
	do.Result = authors
	log.V(2).Info(" :", "len", len(authors))
	return nil
}

type LastAuthor struct {
	DemoDoer[demo.Author]
	Id int64
}

func LastAuthorDo(ctx context.Context, do *LastAuthor) error {
	log := log.WithName(do.Title())
	stat, err := do.Stmt().StatAuthor(ctx)
	if err != nil {
		return err
	}
	log.V(2).Info(" :", "stat", stat)
	if stat.Size <= 0 {
		return nil
	}
	if id, ok := stat.MaxID.(int64); ok {
		fetched, err := do.Stmt().GetAuthor(ctx, id)
		do.Id = id
		do.Result = fetched
		if err != nil {
			return err
		}
		log.V(2).Info(" :", "fetched", fetched)
	} else {
		return fmt.Errorf("the value is not of type int64")
	}
	//panic("fake panic")
	//return fmt.Errorf("fake error")
	return nil
}

type PushAuthor struct {
	DemoDoer[int64]
	Insert demo.CreateAuthorParams
}

func PushAuthorDo(ctx context.Context, do *PushAuthor) error {
	log := log.WithName(do.Title())
	var err error
	inserted, err := do.Stmt().CreateAuthor(ctx, do.Insert)
	if err != nil {
		return err
	}
	do.Result = inserted.ID
	log.V(2).Info(" :", "inserted", inserted)
	fetched, err := do.Stmt().GetAuthor(ctx, inserted.ID)
	if err != nil {
		return err
	}
	log.V(2).Info(" :", "equals", reflect.DeepEqual(inserted, fetched))
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
