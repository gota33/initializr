package initializr

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/gota33/initializr/internal/log"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	log.SetOutput(os.Stderr)
	os.Exit(m.Run())
}

func TestA(t *testing.T) {
	providers := map[string]Provider{
		"demo": MockProvider{},
	}
	c := NewContainer(ContainerOptions{
		Providers: providers,
		CronExpr:  "*/5 * * * * *",
	})

	go func() {
		err := c.Run(context.Background())
		require.NoError(t, err)
	}()

	time.Sleep(30 * time.Second)
	c.Stop()
}

type MockProvider struct{}

func (p MockProvider) Hash() string {
	return fmt.Sprintf("%x", md5.Sum([]byte(strconv.Itoa(rand.Int()))))
}

func (p MockProvider) Provide(ctx context.Context) (conn io.Closer, err error) {
	time.Sleep(time.Second)
	conn = MockConnection{}
	log.Printf("Provide connection")
	return
}

type MockConnection struct{}

func (c MockConnection) Close() (_ error) {
	log.Printf("Close connection")
	return
}
