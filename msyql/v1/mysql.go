package mysql

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/gota33/initializr"
)

type Options struct {
	mysql.Config
	Loc          LocationString         // Location for time.Time values
	Timeout      initializr.DurationStr // Dial timeout
	ReadTimeout  initializr.DurationStr // I/O read timeout
	WriteTimeout initializr.DurationStr // I/O write timeout
}

func New(res initializr.Resource, key string, defaultProvider func() (*sql.DB, func())) (db *sql.DB, close func()) {
	onError := func(err error) (db *sql.DB, close func()) {
		if defaultProvider != nil {
			db, close = defaultProvider()
		}
		if db == nil || close == nil {
			log.Panicf("MySQL init error: %s", err)
		} else {
			log.Printf("MySQL use default, cause: %s", err)
		}
		return
	}

	var (
		opts Options
		err  error
	)
	if err = res.Scan(key, &opts); err != nil {
		return onError(err)
	}

	dsn := opts.unwrap().FormatDSN()
	if db, err = sql.Open("mysql", dsn); err != nil {
		return onError(err)
	}

	close = func() {
		if err := db.Close(); err != nil {
			log.Printf("Fail to close DB: %q", key)
		} else {
			log.Printf("Close mysql: %q", key)
		}
	}
	return
}

type LocationString time.Location

func (s *LocationString) UnmarshalJSON(data []byte) (err error) {
	var (
		name string
		loc  *time.Location
	)
	if err = json.Unmarshal(data, &name); err != nil {
		return
	}
	if loc, err = time.LoadLocation(name); err != nil {
		return
	}
	*s = LocationString(*loc)
	return
}

func (o Options) unwrap() *mysql.Config {
	loc := time.Location(o.Loc)
	out := o.Config
	out.Loc = &loc
	out.Timeout = time.Duration(o.Timeout)
	out.ReadTimeout = time.Duration(o.ReadTimeout)
	out.WriteTimeout = time.Duration(o.WriteTimeout)
	return &out
}
