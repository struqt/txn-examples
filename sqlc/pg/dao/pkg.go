package dao

import (
	"github.com/struqt/logging"
	"sync"
)

var (
	once sync.Once
	log  logging.Logger
)

func Setup(logger logging.Logger) {
	once.Do(func() {
		log = logger
	})
}
