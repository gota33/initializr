package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"time"
)

func OnError(name string, err error, provider interface{}, targets ...interface{}) {
	const maxStack = 100

	log.Printf("%q use provider, cause: %s", name, err)

	prefix := fmt.Sprintf("%q init error: ", name)
	logger := log.New(os.Stderr, prefix, log.LstdFlags|log.Lmsgprefix)

	if reflect.ValueOf(provider).IsNil() {
		logger.Panic("provider is nil")
	}

	tvs := make([]reflect.Value, len(targets))
	for i, t := range targets {
		tvs[i] = reflect.ValueOf(t).Elem()
	}
	values := reflect.ValueOf(provider).Call(nil)

	for i, v := range values {
		// Check nil
		if v.IsZero() {
			logger.Panic("provide nil value")
		}

		// Find assignable value
		tv := tvs[i]
		cv := v
		ok := false

		for j := 0; j < maxStack; j++ {
			ct, tvt := cv.Type(), tv.Type()
			if ok = ct.AssignableTo(tvt); ok || ct.Kind() != reflect.Interface {
				break
			}
			cv = v.Elem()
		}

		// Do assign
		if ok {
			tv.Set(cv)
		} else {
			logger.Panicf("can't assign %q to %q", cv.Type(), tv.Type())
		}
	}
}

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

type MapResource map[string]interface{}

func (r MapResource) GetString(key string, defaultValue string) (v string) {
	r.MustScan(key, &v, func() interface{} { return defaultValue })
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

	OnError(key, err, provider, target)
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
