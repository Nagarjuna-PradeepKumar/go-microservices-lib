package db

import (
	"database/sql"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"os"
	"sync"
)

var once sync.Once
var db *sql.DB

func InitDB(dbSource string, logger log.Logger, name string) *sql.DB {
	var err error
	once.Do(func() {
		db, err = sql.Open("postgres", dbSource)
		if err != nil {
			_ = level.Error(logger).Log("exit", err)
			os.Exit(-1)
		}
		_ = level.Info(logger).Log("msg", "postgres connection established", "connection-for", name)
	})
	return db
}

func CloseDB(db *sql.DB, logger log.Logger, name string) {
	if db != nil {
		err := db.Close()
		if err != nil {
			_ = level.Error(logger).Log("exit", err)
			os.Exit(-1)
		}
		_ = level.Info(logger).Log("msg", "postgres connection closed", "connection-for", name)
	}
}
