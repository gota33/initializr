package initializr

import (
	"context"
)

func Run(ctx context.Context, start func() error, stops ...func()) (err error) {
	chErr := make(chan error, 1)

	go func() {
		if err := start(); err != nil {
			chErr <- err
		}
	}()

	select {
	case err = <-chErr:
	case <-ctx.Done():
		for _, stop := range stops {
			stop()
		}
	}
	return
}

type IService interface {
	Start() (err error)
	Stop()
}

func RunService(ctx context.Context, srv IService) error {
	return Run(ctx, srv.Start, srv.Stop)
}
