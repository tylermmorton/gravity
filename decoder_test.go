package gravity_test

import (
	"github.com/tylermmorton/gravity"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func Test_Decode(t *testing.T) {
	t.Run("successfully decodes a json body", func(t *testing.T) {
		var (
			req *http.Request
		)

		req = httptest.NewRequest(
			http.MethodPost,
			"/",
			strings.NewReader(`{"name":"test"}`),
		)
		req.Header.Set("Content-Type", "application/json")

		type testType struct {
			Name string `json:"name"`
		}

		obj, err := gravity.DecodeRequest[testType](req)
		if err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}

		if !reflect.DeepEqual(obj, &testType{
			Name: "test",
		}) {
			t.Fatalf("unexpected object: %v", obj)
		}
	})
}
