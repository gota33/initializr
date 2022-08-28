package resource

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/gota33/initializr/internal/log"
	"github.com/robfig/cron/v3"
)

type Provider interface {
	Hash() string
	Provide(ctx context.Context) (connection io.Closer, err error)
}

type Container interface {
	Get(name string) (c io.Closer)
	Run(ctx context.Context) (err error)
	Stop()
}

type Resource struct {
	Name       string
	Hash       string
	Connection io.Closer
	LastUpdate time.Time
}

func (r *Resource) Close() (_ error) {
	if r.Connection == nil {
		return
	}
	return r.Connection.Close()
}

func (r *Resource) Update(ctx context.Context, p Provider) (err error) {
	var (
		nextHash string
		conn     io.Closer
		prev     = r.Connection
	)
	if nextHash = p.Hash(); nextHash == r.Hash {
		return
	}
	if conn, err = p.Provide(ctx); err != nil {
		return
	}

	log.Printf("Update connection %q: %s => %s", r.Name, r.Hash, nextHash)

	r.Hash = nextHash
	r.Connection = conn
	r.LastUpdate = time.Now()

	if prev != nil {
		go func() {
			if err := prev.Close(); err != nil {
				log.Printf("Fail to close prev connection: %v", err)
			}
		}()
	}
	return
}

type ContainerOptions struct {
	Providers map[string]Provider
	CronExpr  string
}

type container struct {
	providers map[string]Provider
	resources *sync.Map
}

func NewContainer(opts ContainerOptions) Container {
	c := container{
		providers: opts.Providers,
		resources: &sync.Map{},
	}

	if opts.CronExpr == "" {
		return c
	}

	return containerWithCron{
		container: c,
		cronExpr:  opts.CronExpr,
		scheduler: cron.New(cron.WithSeconds()),
	}
}

func (c container) Run(ctx context.Context) (err error) {
	for name, p := range c.providers {
		if err = c.doUpdate(ctx, name, p); err != nil {
			return
		}
	}
	return
}

func (c container) Stop() {
	c.resources.Range(func(name, value any) bool {
		if res, _ := value.(*Resource); res != nil {
			if err := res.Connection.Close(); err != nil {
				log.Printf("Fail to close connection: %s", name)
			}
		}
		return true
	})
}

func (c container) doUpdate(ctx context.Context, name string, p Provider) (err error) {
	value, _ := c.resources.LoadOrStore(name, &Resource{Name: name})
	info, _ := value.(*Resource)
	return info.Update(ctx, p)
}

func (c container) Get(name string) (conn io.Closer) {
	if value, ok := c.resources.Load(name); ok {
		if conn, ok = value.(io.Closer); ok {
			return
		}
	}
	return
}

type containerWithCron struct {
	container
	cronExpr  string
	scheduler *cron.Cron
}

func (c containerWithCron) Run(ctx context.Context) (err error) {
	if err = c.container.Run(ctx); err != nil {
		return
	}
	_, err = c.scheduler.AddFunc(c.cronExpr, func() {
		if err := c.container.Run(ctx); err != nil {
			log.Printf("Fail to reload container: %v", err)
		}
	})
	c.scheduler.Run()
	return
}

func (c containerWithCron) Stop() {
	<-c.scheduler.Stop().Done()
	c.container.Stop()
}
