package redis

import (
	"log"

	"github.com/go-redis/redis/v7"
	"github.com/gota33/initializr"
)

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
