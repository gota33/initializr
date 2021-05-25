package logrus

import (
	"encoding/json"
	"log"

	sls "github.com/GotaX/logrus-aliyun-log-hook"
	"github.com/gota33/initializr"
	"github.com/sirupsen/logrus"
)

type Options struct {
	Level    LevelString `json:"level"`
	Endpoint string      `json:"endpoint"`
	Key      string      `json:"key"`
	Secret   string      `json:"secret"`
	Project  string      `json:"project"`
	Name     string      `json:"name"`
	Topic    string      `json:"topic"`
	Extra    []string    `json:"extra"`
	Default  bool        `json:"default"`
	Async    bool        `json:"async"`
	Color    bool        `json:"color"`
}

func New(res initializr.Resource, key string, defaultProvider func() (*logrus.Logger, func())) (logger *logrus.Logger, close func()) {
	onError := func(err error) (logger *logrus.Logger, close func()) {
		if defaultProvider != nil {
			logger, close = defaultProvider()
		}
		if logger == nil || close == nil {
			log.Panicf("Logrus init error: %s", err)
		} else {
			log.Printf("Logrus use default, cause: %s", err)
		}
		return
	}

	var (
		opt  Options
		hook *sls.Hook
		err  error
	)

	if !initializr.IsDev() {
		if err = res.Scan(key, &opt); err != nil {
			return onError(err)
		}

		c := sls.Config{
			Endpoint:     opt.Endpoint,
			AccessKey:    opt.Key,
			AccessSecret: opt.Secret,
			Project:      opt.Project,
			Store:        opt.Name,
			Topic:        opt.Topic,
			Extra:        initializr.LogExtra,
		}

		if hook, err = sls.New(c); err != nil {
			return onError(err)
		}

		close = func() {
			if err := hook.Close(); err != nil {
				log.Printf("Fail to close sls: %q", key)
			}
		}
	} else {
		close = func() {}
	}

	if opt.Default {
		logger = logrus.StandardLogger()
	} else {
		logger = logrus.New()
	}

	if hook != nil {
		logger.AddHook(hook)
	}

	logger.SetLevel(logrus.Level(opt.Level))
	logger.SetFormatter(&logrus.TextFormatter{
		ForceColors:     opt.Color,
		FullTimestamp:   true,
		TimestampFormat: "15:04:05",
	})
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
