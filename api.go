package gravity

import (
	"context"
	"net/http"
)

type SelfValidator interface {
	Validate(ctx context.Context) error
}

// ErrorBoundary handles all errors returned by read and write operations in a .
type ErrorBoundary interface {
	ErrorBoundary(wr http.ResponseWriter, req *http.Request, err error) http.HandlerFunc
}

// PanicBoundary is a panic recovery handler. It catches any unhandled errors.
//
// If a handler is not returned to redirect the request, a stack trace is printed
// to the server logs.
type PanicBoundary interface {
	PanicBoundary(wr http.ResponseWriter, req *http.Request, err error) http.HandlerFunc
}
