package foo

import (
	"context"
	"net/http"
)

type service interface {
	Foo(ctx context.Context) (interface{}, error)
}

type fooHandler struct {
	foo service
}

func New(fooClient service) *fooHandler {
	return &fooHandler{
		foo: fooClient,
	}
}

func (f *fooHandler) Handle(w http.ResponseWriter, req *http.Request) {

	_, _ = f.foo.Foo(context.Background())

}
