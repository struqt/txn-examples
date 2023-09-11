package dao

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DemoStmt = TxnBeginner

type DemoDoer[Result any] struct {
	TxnDoerBase[DemoStmt, Result]
}

func (do *DemoDoer[_]) BeginTxn(ctx context.Context, db TxnBeginner) (Txn, error) {
	if w, err := TxnBegin(ctx, db, do.Options()); err != nil {
		return nil, err
	} else {
		do.SetClient(db)
		return w, nil
	}
}

type Demo = TxnModule[DemoStmt]

func NewDemo(db TxnBeginner) Demo {
	i := &TxnModuleBase[DemoStmt]{}
	i.Init(db)
	return i
}

type PushAuthor struct {
	DemoDoer[int64]
	Insert bson.M
}

func PushAuthorDo(ctx context.Context, do *PushAuthor) error {
	log := log.WithName(do.Title()).V(2)
	client := do.Client()
	//options.Database()
	collection := client.Database("demo").Collection("authors")
	res, err := collection.InsertOne(ctx, do.Insert)
	if err != nil {
		return err
	}
	newDoc := collection.FindOne(ctx, bson.M{"_id": res.InsertedID})
	var doc bson.M
	err = newDoc.Decode(&doc)
	if err != nil {
		return err
	}
	//log.Info("|", "inserted", doc)
	log.Info("|", "inserted_id", doc["_id"], "created_at", doc["createdAt"])

	total, err := collection.CountDocuments(ctx, bson.D{})
	if err != nil {
		return err
	}
	log.Info("|", "total", total)
	count := 1
	for total > 10 {
		if count > 10 {
			break
		}
		opts := options.FindOneAndDelete().SetSort(bson.M{"createdAt": 1})
		deleted := collection.FindOneAndDelete(ctx, bson.D{}, opts)
		if deleted.Err() != nil {
			return deleted.Err()
		}
		var doc bson.M
		if err = deleted.Decode(&doc); err == nil {
			log.Info("|", "deleted_id", doc["_id"], "created_at", doc["createdAt"])
		}
		count++
		total, err = collection.CountDocuments(ctx, bson.D{})
		if err != nil {
			return err
		}
	}
	log.Info("|", "total", total)
	//panic("fake panic")
	//return errors.New("intended for test")
	return nil
}

type ListAuthor struct {
	DemoDoer[any]
	len int
}

func ListAuthorDo(ctx context.Context, do *ListAuthor) error {
	log := log.WithName(do.Title()).V(2)
	client := do.Client()
	//options.Database()
	collection := client.Database("demo").Collection("authors")
	findOptions := options.Find()
	findOptions.SetSort(bson.M{"_id": -1})
	findOptions.SetLimit(10)
	cur, err := collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return err
	}
	defer func(cur *mongo.Cursor) { _ = cur.Close(context.Background()) }(cur)
	for cur.Next(ctx) {
		var doc bson.M
		err := cur.Decode(&doc)
		if err != nil {
			return err
		}
		do.len++
		log.Info("|", "_id", doc["_id"], "created_at", doc["createdAt"])
	}
	if cur.Err() != nil {
		return cur.Err()
	}
	log.V(2).Info("|", "len", do.len)
	return nil
}
