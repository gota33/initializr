package initializr

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"time"

	"github.com/tidwall/gjson"
)

type Configuration interface {
	Scan(key string, target any) error
	MustScan(key string, target any, provide func() any)
	Get(key string) Configuration
	Exists() bool
	AsString(defaultValue string) string
	AsDuration(defaultValue time.Duration) time.Duration
	AsInt64(defaultValue int64) int64
	AsFloat64(defaultValue float64) float64
	AsBoolean(defaultValue bool) bool
	AsMap(defaultValue map[string]Configuration) map[string]Configuration
	AsArray(defaultValue []Configuration) []Configuration
	AsUrlValues(defaultValue url.Values) url.Values
}

func FromJson(reader io.Reader) (res Configuration, err error) {
	var data []byte
	if data, err = io.ReadAll(reader); err != nil {
		return
	}
	res = jsonNode(gjson.ParseBytes(data))
	return
}

func FromJsonRemoteCtx(ctx context.Context, url string) (res Configuration, err error) {
	var (
		req  *http.Request
		resp *http.Response
	)
	if req, err = http.NewRequestWithContext(ctx, http.MethodGet, url, nil); err != nil {
		return
	}
	if resp, err = http.DefaultClient.Do(req); err != nil {
		return
	}
	defer func() { _ = resp.Body.Close() }()
	return FromJson(resp.Body)
}

func FromJsonRemote(url string) (res Configuration, err error) {
	return FromJsonRemoteCtx(context.Background(), url)
}

type jsonNode gjson.Result

func (n jsonNode) get(key string) (gjson.Result, bool) {
	node := gjson.Result(n).Get(key)
	return node, node.Exists()
}

func (n jsonNode) raw() gjson.Result {
	return gjson.Result(n)
}

func (n jsonNode) Get(key string) Configuration {
	if node, ok := n.get(key); ok {
		return jsonNode(node)
	}
	return nil
}

func (n jsonNode) Exists() bool {
	return n.raw().Exists()
}

func (n jsonNode) Scan(key string, target interface{}) (err error) {
	if node, ok := n.get(key); ok {
		return json.Unmarshal([]byte(node.Raw), target)
	} else {
		return fmt.Errorf("resource key not exists: %s", key)
	}
}

func (n jsonNode) MustScan(key string, target any, provide func() any) {
	if err := n.Scan(key, target); err != nil {
		if provide != nil {
			v := reflect.ValueOf(provide())
			reflect.ValueOf(target).Set(v)
		} else {
			log.Panicf("JsonNode.MustScan(%s): %v", key, err)
		}
	}
}

func (n jsonNode) AsString(defaultValue string) string {
	if n.Exists() {
		return n.raw().String()
	}
	return defaultValue
}

func (n jsonNode) AsDuration(defaultValue time.Duration) time.Duration {
	if n.Exists() {
		if dur, err := time.ParseDuration(n.raw().String()); err == nil {
			return dur
		} else {
			log.Printf("JsonNode.GetDuration: %v", err)
		}
	}
	return defaultValue
}

func (n jsonNode) AsInt64(defaultValue int64) int64 {
	if n.Exists() {
		return n.raw().Int()
	}
	return defaultValue
}

func (n jsonNode) AsFloat64(defaultValue float64) float64 {
	if n.Exists() {
		return n.raw().Float()
	}
	return defaultValue
}

func (n jsonNode) AsBoolean(defaultValue bool) bool {
	if n.Exists() {
		return n.raw().Bool()
	}
	return defaultValue
}

func (n jsonNode) AsMap(defaultValue map[string]Configuration) map[string]Configuration {
	if n.Exists() {
		m0 := n.raw().Map()
		m1 := make(map[string]Configuration, len(m0))
		for k, v := range m0 {
			m1[k] = jsonNode(v)
		}
		return m1
	}
	return defaultValue
}

func (n jsonNode) AsArray(defaultValue []Configuration) []Configuration {
	if n.Exists() {
		arr0 := n.raw().Array()
		arr1 := make([]Configuration, len(arr0))
		for i, v := range arr0 {
			arr1[i] = jsonNode(v)
		}
		return arr1
	}
	return defaultValue
}

func (n jsonNode) AsUrlValues(defaultValue url.Values) url.Values {
	if n.raw().Exists() {
		m0 := n.raw().Map()
		m1 := make(url.Values, len(m0))
		for k, v := range m0 {
			if str := v.String(); str != "" {
				m1.Set(k, str)
			}
		}
		return m1
	}
	return defaultValue
}
