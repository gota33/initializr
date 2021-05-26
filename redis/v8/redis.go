package redis

import (
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/gota33/initializr"
	"github.com/gota33/initializr/internal"
)

type Provider func() (*redis.Client, func())

func MustNew(src initializr.Resource, key string, defaultProvider Provider) (rc *redis.Client, shutdown func()) {
	rc, shutdown, err := New(src, key)
	if err != nil {
		internal.OnError("Redis", err, defaultProvider, &rc, &shutdown)
	}
	return
}

func New(src initializr.Resource, key string) (rc *redis.Client, close func(), err error) {
	var opt redis.Options
	if err = src.Scan(key, &opt); err != nil {
		return
	}

	rc = redis.NewClient(&opt)
	close = func() {
		if err := rc.Close(); err != nil {
			log.Printf("Fail to close redis: %q", key)
		} else {
			log.Printf("Close redis: %q", key)
		}
	}
	return
}
