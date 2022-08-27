package initializr

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"time"

	"github.com/tidwall/gjson"
)

type Configuration interface {
	Scan(key string, target any) error
	MustScan(key string, target any, provide func() any)
	GetString(key, defaultValue string) string
	GetDuration(key string, defaultValue time.Duration) time.Duration
	GetNumber(key string, defaultValue json.Number) json.Number
	GetInt64(key string, defaultValue int64) int64
	GetFloat64(key string, defaultValue float64) float64
	GetBoolean(key string, defaultValue bool) bool
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

func (n jsonNode) GetString(key, defaultValue string) string {
	if node, ok := n.get(key); ok {
		return node.String()
	}
	return defaultValue
}

func (n jsonNode) GetDuration(key string, defaultValue time.Duration) time.Duration {
	if node, ok := n.get(key); ok {
		if dur, err := time.ParseDuration(node.String()); err == nil {
			return dur
		} else {
			log.Printf("JsonNode.GetDuration(%s): %v", key, err)
		}
	}
	return defaultValue
}

func (n jsonNode) GetNumber(key string, defaultValue json.Number) json.Number {
	if node, ok := n.get(key); ok {
		return json.Number(node.String())
	}
	return defaultValue
}

func (n jsonNode) GetInt64(key string, defaultValue int64) int64 {
	if node, ok := n.get(key); ok {
		return node.Int()
	}
	return defaultValue
}

func (n jsonNode) GetFloat64(key string, defaultValue float64) float64 {
	if node, ok := n.get(key); ok {
		return node.Float()
	}
	return defaultValue
}

func (n jsonNode) GetBoolean(key string, defaultValue bool) bool {
	if node, ok := n.get(key); ok {
		return node.Bool()
	}
	return defaultValue
}
