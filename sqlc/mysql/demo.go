package main

import (
	"context"
	"fmt"
	"reflect"

	"examples/sqlc/mysql/demo"
)

type FetchLastAuthorDoer struct {
	DemoDoerBase
	id int64
}

func FetchLastAuthorDo(ctx context.Context, do *FetchLastAuthorDoer) error {
	log := log.WithName(do.Title())
	stat, err := do.query.StatAuthor(ctx)
	if err != nil {
		return err
	}
	log.V(2).Info("", "stat", stat)
	if stat.Size <= 0 {
		return nil
	}
	if id, ok := stat.MaxID.(int64); ok {
		fetched, err := do.query.GetAuthor(ctx, id)
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
	insert   demo.CreateAuthorParams
	inserted int64
}

func PushAuthorDo(ctx context.Context, do *PushAuthorDoer) error {
	log := log.WithName(do.Title())
	var err error
	inserted, err := do.query.CreateAuthor(ctx, do.insert)
	if err != nil {
		return err
	}
	do.inserted, _ = inserted.LastInsertId()
	log.V(2).Info("", "inserted", do.inserted)
	fetched, err := do.query.GetAuthor(ctx, do.inserted)
	if err != nil {
		return err
	}
	log.V(2).Info("", "equals", reflect.DeepEqual(do.inserted, fetched.ID))
	count := 1
	for {
		if count > 10 {
			break
		}
		stat, err := do.query.StatAuthor(ctx)
		if err != nil {
			return err
		}
		log.V(2).Info("", "stat", stat)
		if stat.Size <= 5 {
			break
		}
		if id, ok := stat.MinID.(int64); ok {
			if err = do.query.DeleteAuthor(ctx, id); err != nil {
				return err
			}
			count++
		} else {
			return fmt.Errorf("the value is not of type int64")
		}
	}
	authors, err := do.query.ListAuthors(ctx)
	if err != nil {
		return err
	}
	log.V(2).Info("", "list", len(authors))
	//panic("fake panic")
	//return fmt.Errorf("fake error")
	return nil
}
