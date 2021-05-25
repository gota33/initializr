package redis

import (
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/gota33/initializr"
)

func New(src initializr.Resource, key string, defaultProvider func() (*redis.Client, func())) (rc *redis.Client, close func()) {
	onError := func(err error) (db *redis.Client, close func()) {
		if defaultProvider != nil {
			db, close = defaultProvider()
		}
		if db == nil || close == nil {
			log.Panicf("Redis init error: %s", err)
		} else {
			log.Printf("Redis use default, cause: %s", err)
		}
		return
	}

	var (
		opt redis.Options
		err error
	)
	if err = src.Scan(key, &opt); err != nil {
		return onError(err)
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
