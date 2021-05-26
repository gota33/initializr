package logrus

import (
	"encoding/json"
	"fmt"
	"log"

	sls "github.com/GotaX/logrus-aliyun-log-hook"
	"github.com/gota33/initializr"
	"github.com/gota33/initializr/internal"
	"github.com/sirupsen/logrus"
)

var Extra = map[string]interface{}{
	initializr.ServiceKey: &initializr.Service,
	initializr.VersionKey: &initializr.Version,
}

type Options struct {
	Level    LevelString `json:"level"`
	Color    bool        `json:"color"`
	Default  bool        `json:"default"`
	Endpoint string      `json:"endpoint"`
	Key      string      `json:"key"`
	Secret   string      `json:"secret"`
	Project  string      `json:"project"`
	Name     string      `json:"name"`
	Topic    string      `json:"topic"`
}

type Provider func() (logger *logrus.Logger, shutdown func())

func MustNew(res initializr.Resource, key string, defaultProvider Provider) (logger *logrus.Logger, shutdown func()) {
	logger, shutdown, err := New(res, key)
	if err != nil {
		internal.OnError("Logrus", err, defaultProvider, &logger, &shutdown)
	}
	return
}

func New(res initializr.Resource, key string) (logger *logrus.Logger, shutdown func(), err error) {
	var opt Options
	if err = res.Scan(key, &opt); err != nil {
		return
	}

	if opt.Default {
		logger = logrus.StandardLogger()
	} else {
		logger = logrus.New()
	}

	logger.SetLevel(logrus.Level(opt.Level))
	logger.SetFormatter(&logrus.TextFormatter{
		ForceColors:     opt.Color,
		FullTimestamp:   true,
		TimestampFormat: "15:04:05",
	})

	shutdown = func() {}

	if !initializr.IsDev() {
		extra := make(map[string]string, len(Extra))
		for k, v := range Extra {
			extra[k] = fmt.Sprintf("%s", v)
		}

		c := sls.Config{
			Endpoint:     opt.Endpoint,
			AccessKey:    opt.Key,
			AccessSecret: opt.Secret,
			Project:      opt.Project,
			Store:        opt.Name,
			Topic:        opt.Topic,
			Extra:        extra,
		}

		var hook *sls.Hook
		if hook, err = sls.New(c); err != nil {
			return
		}

		shutdown = func() {
			if err := hook.Close(); err != nil {
				log.Printf("Fail to close sls: %q", key)
			}
		}

		logger.AddHook(hook)
	}
	return
}

type LevelString logrus.Level

func (l *LevelString) UnmarshalJSON(data []byte) (err error) {
	var (
		str   string
		level logrus.Level
	)
	if err = json.Unmarshal(data, &str); err != nil {
		return
	}
	if level, err = logrus.ParseLevel(str); err != nil {
		return
	}

	*l = LevelString(level)
	return
}
