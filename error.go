package gravity

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse is the data structure used to render an error to the response body.
type ErrorResponse struct {
	Error      error
	StackTrace string
}

func writeErrorResponse(wr http.ResponseWriter, req *http.Request, err error, stack []byte) error {
	var res = ErrorResponse{
		Error:      err,
		StackTrace: string(stack),
	}

	switch req.Header.Get("Accept") {
	case "application/json":
		return json.NewEncoder(wr).Encode(&res)
	}

	http.Error(wr, err.Error(), http.StatusInternalServerError)
	return nil
}
