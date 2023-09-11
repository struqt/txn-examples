package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/struqt/logging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"examples/mongo/dao"
)

var log logging.Logger

func init() {
	logging.LogConsoleThreshold = -1
	log = logging.NewLogger("")
	dao.Setup(log)
}

func do(ctx context.Context, m dao.Demo) {
	i := bson.M{"name": "Brian Kernighan", "age": 30, "createdAt": time.Now()}
	_, _ = dao.TxnExecute(ctx, m, &dao.ListAuthor{}, dao.ListAuthorDo)
	_, _ = dao.TxnExecute(ctx, m, &dao.PushAuthor{Insert: i}, dao.PushAuthorDo)
	_, _ = dao.TxnExecute(ctx, m, &dao.ListAuthor{}, dao.ListAuthorDo)
}

func main() {
	defer os.Exit(0)
	defer log.Info("Process is ending ...")
	log.Info("Process is starting ...")
	ctx, cancel := context.WithCancel(context.Background())
	defer log.Info("Context is cancelled")
	defer cancel()
	addr, uri := address()
	client, err := open(ctx, uri)
	defer func(client *mongo.Client) { _ = client.Disconnect(context.Background()) }(client)
	if err != nil {
		log.Error(err, "Failed to set up connection pool")
		return
	}
	ping(client, addr)
	log.Info("Connected", "addr", addr)
	m := dao.NewDemo(client)
	do(ctx, m)
}

func ping(client *mongo.Client, addr string) {
	for {
		if cnt, err := dao.TxnPing(client, func(cnt int, delay time.Duration) {
			log.Info("Ping", "count", cnt, "delay~", delay, "target", addr)
		}); err == nil {
			break
		} else {
			log.V(1).Info("Ping", "count", cnt, "err", err)
		}
	}
}

func open(ctx context.Context, uri string) (*mongo.Client, error) {
	clientOptions := options.Client()
	clientOptions.ApplyURI(uri)
	clientOptions.SetReplicaSet("rs0")
	return mongo.Connect(ctx, clientOptions)
}

func host() (addr string) {
	addr = os.Getenv("DB_ADDR_UDS")
	if len(addr) > 0 {
		addr = fmt.Sprintf("%s", addr)
	} else {
		addr = os.Getenv("DB_ADDR_TCP")
		if len(addr) <= 0 {
			addr = "127.0.0.1"
		}
		addr = fmt.Sprintf("%s:27017", addr)
	}
	return
}

func address() (string, string) {
	h := host()
	passwd := os.Getenv("DB_PASSWORD")
	return h, fmt.Sprintf("mongodb://example:%s@%s", url.QueryEscape(passwd), h)
}
