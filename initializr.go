package initializr

import (
	"context"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gota33/initializr/internal/log"
)

var graceful struct {
	sync.Once
	Ctx    context.Context
	Cancel func()
}

func GracefulContext() (context.Context, func()) {
	graceful.Do(func() {
		graceful.Ctx, graceful.Cancel = context.WithCancel(context.Background())

		c := make(chan os.Signal, 2)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-c
			log.Println("Try exit...")
			graceful.Cancel()

			<-c
			log.Println("Force exit")
			os.Exit(0)
		}()
	})
	return graceful.Ctx, graceful.Cancel
}

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

func DeferClose(ctx context.Context, name string, conn io.Closer) {
	go func() {
		<-ctx.Done()

		if conn != nil {
			if err := conn.Close(); err != nil {
				log.Printf("Close connection error %q: %v", name, err)
			} else {
				log.Printf("Closed connection %q", name)
			}
		}
	}()
}
