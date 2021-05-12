package initializr

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

var (
	Version  = "dev"
	LogExtra = map[string]string{"version": Version}

	graceful struct {
		sync.Once
		Ctx    context.Context
		Cancel func()
	}
)

type Resource interface {
	Scan(key string, target interface{}) error
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
	var raw MapResource
	err = json.NewDecoder(reader).Decode(&raw)
	res = raw
	return
}

type MapResource map[string]interface{}

func (r MapResource) Scan(key string, target interface{}) (err error) {
	var (
		ok    bool
		value interface{}
		data  []byte
	)
	if value, ok = r.get(strings.Split(key, ".")); !ok {
		return fmt.Errorf("cofig key not found: %s", key)
	}
	if data, err = json.Marshal(value); err != nil {
		return
	}
	return json.Unmarshal(data, &target)
}

func (r MapResource) get(sections []string) (out interface{}, ok bool) {
	switch len(sections) {
	case 0:
		return
	case 1:
		out, ok = r[sections[0]]
	default:
		head, tail := sections[0], sections[1:]
		if sub, ok0 := r[head]; ok0 {
			if v, ok1 := sub.(map[string]interface{}); ok1 {
				return MapResource(v).get(tail)
			}
		}
	}
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
