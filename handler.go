package gravity

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"runtime/debug"
)

func (ctl *controllerImpl[T, K]) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
	// defer a panic recoverer and pass panics to the PanicBoundary
	defer func() {
		if err, ok := recover().(error); ok && err != nil {
			ctl.handlePanic(wr, req, err)
			return
		}
	}()

	log.Printf("[Request] (http) %s -> %T\n", req.URL, ctl.handler)

	request, err := DecodeRequest[T](req)
	if err != nil {
		ctl.handleError(wr, req, err)
		return
	}

	if v, ok := interface{}(request).(SelfValidator); ok {
		err = v.Validate(req.Context())
		if err != nil {
			ctl.handleError(wr, req, err)
			return
		}
	}

	response, err := ctl.handler.Handle(req.Context(), request)
	if err != nil {
		ctl.handleError(wr, req, err)
		return
	}

	switch req.Header.Get("Accept") {
	case "application/json":
		err = json.NewEncoder(wr).Encode(response)
	default:
		err = errors.New("mime type provided in Accept header is not supported")
	}
	if err != nil {
		ctl.handleError(wr, req, err)
		return
	}

	wr.WriteHeader(http.StatusOK)
}

func (ctl *controllerImpl[T, K]) handleError(wr http.ResponseWriter, req *http.Request, err error) {
	if ctl.errorBoundary != nil {
		// Calls to ErrorBoundary can return an http.HandlerFunc
		// that can be used to cleanly handle the error. Or not
		h := ctl.errorBoundary.ErrorBoundary(wr, req, err)
		if h != nil {
			log.Printf("[ErrorBoundary] %s -> handled\n", req.URL)
			h(wr, req)
			return
		}
	} else {
		// No ErrorBoundary was implemented in the route module.
		// So your error goes to the PanicBoundary.
		log.Printf("[ErrorBoundary] %s -> not implemented\n", req.URL)
		panic(err)
	}
}

func (ctl *controllerImpl[T, K]) handlePanic(wr http.ResponseWriter, req *http.Request, err error) {
	if ctl.panicBoundary != nil {
		// Calls to PanicBoundary can return an http.HandlerFunc
		// that can be used to cleanly handle the error.
		h := ctl.panicBoundary.PanicBoundary(wr, req, err)
		if h != nil {
			log.Printf("[PanicBoundary] %s -> handled\n", req.URL)
			h(wr, req)
			return
		}
	} else {
		stack := debug.Stack()
		log.Printf("[UncaughtPanic] %s\n-- ERROR --\nUncaught panic in route module %T: %+v\n-- STACK TRACE --\n%s", req.URL, ctl.handler, err, stack)
		err = writeErrorResponse(wr, req, err, stack)
		if err != nil {
			log.Printf("[UncaughtPanic] %s -> failed to write error response: %v\n", req.URL, err)
		}
	}
}
