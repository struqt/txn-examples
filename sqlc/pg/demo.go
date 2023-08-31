package main

import (
	"context"
	"fmt"
	"reflect"

	"examples/sqlc/pg/demo"
)

type FetchLastAuthorDoer struct {
	DemoDoerBase
	id int64
}

func FetchLastAuthorDo(ctx context.Context, do *FetchLastAuthorDoer) error {
	log := log.WithName(do.Title())
	stat, err := do.Stmt().StatAuthor(ctx)
	if err != nil {
		return err
	}
	log.V(2).Info("", "stat", stat)
	if stat.Size <= 0 {
		return nil
	}
	if id, ok := stat.MaxID.(int64); ok {
		fetched, err := do.Stmt().GetAuthor(ctx, id)
		do.id = id
		if err != nil {
			return err
		}
		log.V(2).Info("", "fetched.id", fetched.ID, "name", fetched.Name, "bio", fetched.Bio.String)
	} else {
		return fmt.Errorf("the value is not of type int64")
	}
	//panic("fake panic")
	//return fmt.Errorf("fake error")
	return nil
}

type PushAuthorDoer struct {
	DemoDoerBase
	Insert   demo.CreateAuthorParams
	inserted int64
}

func PushAuthorDo(ctx context.Context, do *PushAuthorDoer) error {
	log := log.WithName(do.Title())
	var err error
	inserted, err := do.Stmt().CreateAuthor(ctx, do.Insert)
	if err != nil {
		return err
	}
	do.inserted = inserted.ID
	log.V(2).Info("", "inserted.id", inserted.ID, "name", inserted.Name, "bio", inserted.Bio.String)
	fetched, err := do.Stmt().GetAuthor(ctx, inserted.ID)
	if err != nil {
		return err
	}
	log.V(2).Info("", "equals", reflect.DeepEqual(inserted, fetched))
	count := 1
	for {
		if count > 10 {
			break
		}
		stat, err := do.Stmt().StatAuthor(ctx)
		if err != nil {
			return err
		}
		log.V(2).Info("", "stat", stat)
		if stat.Size <= 5 {
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
	authors, err := do.Stmt().ListAuthors(ctx)
	if err != nil {
		return err
	}
	log.V(2).Info("", "list", len(authors))
	//panic("fake panic")
	//return fmt.Errorf("fake error")
	return nil
}
