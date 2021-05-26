package initializr

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gota33/initializr/internal"
)

const (
	DefaultVersionKey = "version"
	DefaultVersion    = "dev"
	DefaultServiceKey = "service"
	DefaultService    = "app"
)

var (
	Version    = DefaultVersion
	VersionKey = DefaultVersionKey
	Service    = DefaultService
	ServiceKey = DefaultServiceKey

	graceful struct {
		sync.Once
		Ctx    context.Context
		Cancel func()
	}
)

//goland:noinspection GoBoolExpressions
func IsDev() bool { return Version == DefaultVersion }

type Resource interface {
	Scan(key string, target interface{}) error
	MustScan(key string, target interface{}, provide func() interface{})
	GetString(key, defaultValue string) string
	GetDuration(key string, defaultValue time.Duration) time.Duration
	GetNumber(key string, defaultValue json.Number) json.Number
	GetInt64(key string, defaultValue int64) int64
	GetFloat64(key string, defaultValue float64) float64
	GetBoolean(key string, defaultValue bool) bool
}

func FromJsonRemote(url string) (res Resource, err error) {
	var resp *http.Response
	if resp, err = http.Get(url); err != nil {
		return
	}
	defer func() { _ = resp.Body.Close() }()
	return FromJson(resp.Body)
}

func FromJson(reader io.Reader) (res Resource, err error) {
	var mr internal.MapResource
	err = json.NewDecoder(reader).Decode(&mr)
	res = mr
	return
}

func GracefulContext() (context.Context, func()) {
	graceful.Do(func() {
		graceful.Ctx, graceful.Cancel = context.WithCancel(context.Background())

		c := make(chan os.Signal, 2)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-c
			fmt.Println("Try exit...")
			graceful.Cancel()

			<-c
			fmt.Println("Force exit")
			os.Exit(0)
		}()
	})
	return graceful.Ctx, graceful.Cancel
}

func Run(ctx context.Context, start func() error, stops ...func()) (err error) {
	chErr := make(chan error, 1)

	go func() {
		if err := start(); err != nil {
			chErr <- err
		}
	}()

	select {
	case err = <-chErr:
	case <-ctx.Done():
		for _, stop := range stops {
			stop()
		}
	}
	return
}
