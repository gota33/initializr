package example

import (
	"context"
	"io"
)

func main() {

}

type MockProvider struct {
}

func (p MockProvider) Hash() string {
	// TODO implement me
	panic("implement me")
}

func (p MockProvider) Provide(ctx context.Context) (connection io.Closer, err error) {
	// TODO implement me
	panic("implement me")
}
