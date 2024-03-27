package gravity

import (
	"context"
	"net/http"
)

type Request interface{}
type Response interface{}

type Handler[T Request, K Response] interface {
	Handle(ctx context.Context, req *T) (*K, error)
}

type Controller[T Request, K Response] interface {
	http.Handler
}

func New[T Request, K Response](h Handler[T, K]) Controller[T, K] {
	ctl := &controllerImpl[T, K]{
		handler: h,
	}

	if errorBoundary, ok := h.(ErrorBoundary); ok {
		ctl.errorBoundary = errorBoundary
	}

	if panicBoundary, ok := h.(PanicBoundary); ok {
		ctl.panicBoundary = panicBoundary
	}

	return ctl
}

func NewHandlerFunc[T Request, K Response](h Handler[T, K]) http.HandlerFunc {
	ctl := New[T, K](h)
	return func(wr http.ResponseWriter, req *http.Request) {
		ctl.ServeHTTP(wr, req)
	}
}

type controllerImpl[T Request, K Response] struct {
	handler Handler[T, K]

	errorBoundary ErrorBoundary
	panicBoundary PanicBoundary
}
