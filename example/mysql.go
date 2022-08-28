package example

import (
	"context"
	"crypto/md5"
	"database/sql"
	"fmt"
	"io"
	"net/url"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gota33/initializr"
)

type MySQL struct {
	Protocol string
	Host     string
	Database string
	Port     int64
	Username string
	Password string
	MaxIdle  int64
	MaxOpen  int64
	Params   url.Values
}

func NewMySql(c initializr.Configuration) MySQL {
	const (
		keyProtocol = "protocol"
		keyHost     = "host"
		keyDatabase = "database"
		keyPort     = "port"
		keyUsername = "username"
		keyPassword = "password"
		keyMaxIdle  = "maxIdle"
		keyMaxOpen  = "maxOpen"
		keyParams   = "params"
	)
	return MySQL{
		Protocol: c.Get(keyProtocol).AsString("tcp"),
		Host:     c.Get(keyHost).AsString(""),
		Database: c.Get(keyDatabase).AsString(""),
		Port:     c.Get(keyPort).AsInt64(3306),
		Username: c.Get(keyUsername).AsString(""),
		Password: c.Get(keyPassword).AsString(""),
		MaxIdle:  c.Get(keyMaxIdle).AsInt64(0),
		MaxOpen:  c.Get(keyMaxOpen).AsInt64(0),
		Params:   c.Get(keyParams).AsUrlValues(make(url.Values, 0)),
	}
}

func (p MySQL) Hash() string {
	w := md5.New()
	_, _ = fmt.Fprint(w, p.Protocol, p.Host, p.Database, p.Port, p.Username, p.Params, p.MaxIdle, p.MaxOpen)
	for k, v := range p.Params {
		_, _ = fmt.Fprint(w, k, v)
	}
	return fmt.Sprintf("%x", w.Sum(nil))
}

func (p MySQL) Provide(ctx context.Context) (connection io.Closer, err error) {
	var db *sql.DB

	dsn := fmt.Sprintf("%s:%s@%s(%s:%d)/%s?%s",
		p.Username, p.Password, p.Protocol, p.Host, p.Port, p.Database, p.Params.Encode())

	if db, err = sql.Open("mysql", dsn); err != nil {
		return
	}
	if err = db.PingContext(ctx); err != nil {
		return
	}

	if n := p.MaxIdle; n > 0 {
		db.SetMaxIdleConns(int(n))
	}
	if n := p.MaxOpen; n > 0 {
		db.SetMaxOpenConns(int(n))
	}
	return db, nil
}
