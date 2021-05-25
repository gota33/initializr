package initializr

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	DefaultVersion = "dev"
)

var (
	Version  = DefaultVersion
	LogExtra = map[string]string{"version": Version}

	graceful struct {
		sync.Once
		Ctx    context.Context
		Cancel func()
	}
)

func IsDev() bool { return Version == DefaultVersion }

type DurationStr time.Duration

func (v *DurationStr) UnmarshalJSON(data []byte) (err error) {
	var (
		str string
		dur time.Duration
	)
	if err = json.Unmarshal(data, &str); err != nil {
		return
	}
	if dur, err = time.ParseDuration(str); err != nil {
		return
	}
	*v = DurationStr(dur)
	return
}

type Int64Value int64

func (v *Int64Value) UnmarshalJSON(data []byte) (err error) {
	var (
		num json.Number
		d   int64
	)
	if err = json.Unmarshal(data, &num); err != nil {
		return
	}
	if d, err = num.Int64(); err != nil {
		return
	}
	*v = Int64Value(d)
	return
}

type Float64Value float64

func (v *Float64Value) UnmarshalJSON(data []byte) (err error) {
	var (
		num json.Number
		d   float64
	)
	if err = json.Unmarshal(data, &num); err != nil {
		return
	}
	if d, err = num.Float64(); err != nil {
		return
	}
	*v = Float64Value(d)
	return
}

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
	var raw MapResource
	err = json.NewDecoder(reader).Decode(&raw)
	res = raw
	return
}

type MapResource map[string]interface{}

func (r MapResource) GetString(key string, defaultValue string) (v string) {
	if err := r.Scan(key, &v); err != nil {
		r.logError(key, defaultValue, err)
		v = defaultValue
	}
	return
}

func (r MapResource) GetDuration(key string, defaultValue time.Duration) (v time.Duration) {
	r.MustScan(key, (*DurationStr)(&v), func() interface{} { return defaultValue })
	return
}

func (r MapResource) GetNumber(key string, defaultValue json.Number) (v json.Number) {
	r.MustScan(key, &v, func() interface{} { return defaultValue })
	return
}

func (r MapResource) GetInt64(key string, defaultValue int64) (v int64) {
	r.MustScan(key, (*Int64Value)(&v), func() interface{} { return defaultValue })
	return
}

func (r MapResource) GetFloat64(key string, defaultValue float64) (v float64) {
	r.MustScan(key, (*Float64Value)(&v), func() interface{} { return defaultValue })
	return
}

func (r MapResource) GetBoolean(key string, defaultValue bool) (v bool) {
	r.MustScan(key, &v, func() interface{} { return defaultValue })
	return
}

func (r MapResource) MustScan(key string, target interface{}, provider func() interface{}) {
	err := r.Scan(key, target)
	if err == nil {
		return
	}

	if provider == nil {
		log.Panicf("MustScan panic: %v", err)
	}

	defaultValue := provider()
	r.logError(key, defaultValue, err)

	d := reflect.ValueOf(target).Elem()
	v := reflect.ValueOf(defaultValue)
	d.Set(v)
}

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

func (r MapResource) logError(key string, defaultValue interface{}, err error) {
	log.Printf("Use default %s %+v, because %s", key, defaultValue, err)
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
