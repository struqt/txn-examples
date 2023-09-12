package dao

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/struqt/logging"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	once sync.Once
	log  logging.Logger
	demo TxnModule
)

func Setup(logger logging.Logger) {
	once.Do(func() {
		log = logger
		initDemo()
	})
}

func Demo() TxnModule {
	return demo
}

func initDemo() {
	addr, uri := address()
	client, err := open(context.Background(), uri)
	if err != nil {
		log.Error(err, "Failed to set up connection pool")
		return
	}
	ping(client, addr)
	demo = NewTxnModule(client)
}

func ping(client *mongo.Client, addr string) {
	for {
		if cnt, err := TxnPing(client, func(cnt int, delay time.Duration) {
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
