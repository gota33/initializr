package mysql

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/gota33/initializr"
	"github.com/gota33/initializr/internal"
)

type Options struct {
	mysql.Config
	Loc          LocationString       // Location for time.Time values
	Timeout      internal.DurationStr // Dial timeout
	ReadTimeout  internal.DurationStr // I/O read timeout
	WriteTimeout internal.DurationStr // I/O write timeout
}

type Provider func() (*sql.DB, func())

func MustNew(res initializr.Resource, key string, defaultProvider Provider) (db *sql.DB, shutdown func()) {
	db, shutdown, err := New(res, key)
	if err != nil {
		internal.OnError("MySQL", err, defaultProvider, &db, &shutdown)
	}
	return
}

func New(res initializr.Resource, key string) (db *sql.DB, close func(), err error) {
	var opts Options
	if err = res.Scan(key, &opts); err != nil {
		return
	}

	dsn := opts.unwrap().FormatDSN()
	if db, err = sql.Open("mysql", dsn); err != nil {
		return
	}

	close = func() {
		if err := db.Close(); err != nil {
			log.Printf("Fail to close MySQL: %q", key)
		} else {
			log.Printf("Close MySQL: %q", key)
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
