package dao

import (
	"context"
	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DemoDoer[Result any] struct {
	TxnDoerBase[Result]
}

type ListAuthor struct {
	DemoDoer[[]bson.M]
	len int
}

func ListAuthorDo(ctx context.Context, do *ListAuthor) error {
	client := do.Client()
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
		do.Result = append(do.Result, doc)
	}
	slog.With("T", do.Title()).Info("|", "len", do.len)
	for _, record := range do.Result {
		slog.With("T", do.Title()).Info("|", "_id", record["_id"], "created_at", record["createdAt"])
	}
	if cur.Err() != nil {
		return cur.Err()
	}
	return nil
}

type PushAuthor struct {
	DemoDoer[int64]
	Insert bson.M
}

func PushAuthorDo(ctx context.Context, do *PushAuthor) error {
	//log := slog.With("T", do.Title())
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
	//slog.With("T", do.Title()).Info("|", "inserted", doc)
	slog.With("T", do.Title()).Info("|", "inserted_id", doc["_id"], "created_at", doc["createdAt"])
	total, err := collection.CountDocuments(ctx, bson.D{})
	if err != nil {
		return err
	}
	slog.With("T", do.Title()).Info("|", "total", total)
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
			slog.With("T", do.Title()).Info("|", "deleted_id", doc["_id"], "created_at", doc["createdAt"])
		}
		count++
		total, err = collection.CountDocuments(ctx, bson.D{})
		if err != nil {
			return err
		}
	}
	slog.With("T", do.Title()).Info("|", "total", total)
	//panic("fake panic")
	//return errors.New("intended for test")
	return nil
}
