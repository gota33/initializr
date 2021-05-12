package mysql

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/go-sql-driver/mysql"
	"initializr"
)

type Options struct {
	mysql.Config
	Loc          LocationString // Location for time.Time values
	Timeout      DurationString // Dial timeout
	ReadTimeout  DurationString // I/O read timeout
	WriteTimeout DurationString // I/O write timeout
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
			log.Printf("Fail to close DB: %q", key)
		} else {
			log.Printf("Close mysql: %q", key)
		}
	}
	return
}

type DurationString time.Duration

func (d *DurationString) UnmarshalJSON(data []byte) (err error) {
	var (
		value string
		dur   time.Duration
	)
	if err = json.Unmarshal(data, &value); err != nil {
		return
	}
	if dur, err = time.ParseDuration(value); err != nil {
		return
	}
	*d = DurationString(dur)
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
