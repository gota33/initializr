package oss

import (
	"net/http"

	driver "github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/gota33/initializr"
	"github.com/gota33/initializr/internal"
)

type Options struct {
	Id       string `json:"id"`
	Secret   string `json:"secret"`
	Endpoint string `json:"endpoint"`
}

type Provider func() (*driver.Client, func())

func MustNew(res initializr.Resource, key string, defaultProvider Provider) (client *driver.Client, shutdown func()) {
	client, shutdown, err := New(res, key)
	if err != nil {
		internal.OnError("OSS", err, defaultProvider, &client, &shutdown)
	}
	return
}

func New(src initializr.Resource, key string) (client *driver.Client, close func(), err error) {
	var opt Options
	if err = src.Scan(key, &opt); err != nil {
		return
	}
	client, err = driver.New(opt.Endpoint, opt.Id, opt.Secret, driver.HTTPClient(http.DefaultClient))
	close = func() {}
	return
}
