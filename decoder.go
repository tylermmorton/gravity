package gravity

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// DecodeRequest an HTTP request into the provided struct
func DecodeRequest[T Request](req *http.Request) (*T, error) {
	var data = new(T)

	typ := reflect.TypeOf(data)
	if typ == nil {
		return nil, fmt.Errorf("invalid decode type: nil")
	}

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("invalid decode type: %v", typ.Kind())
	}

	err := decodeRequest(req, typ, data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func decodeRequest(r *http.Request, t reflect.Type, data interface{}) error {
	body, err := decodeStruct(r, t, data)
	if err != nil {
		return err
	}

	if !body {
		err := decodeBody(r, data)
		if err != nil {
			return err
		}
	}

	return nil
}

func decodeStruct(r *http.Request, t reflect.Type, data interface{}) (bool, error) {
	query := r.URL.Query()
	//vars := mux.Vars(r)
	body := false
	for i := 0; i < t.NumField(); i++ {
		typ := t.Field(i)
		field := reflect.ValueOf(data).Elem().Field(i)

		if typ.Type.Kind() == reflect.Struct {
			var err error
			if body, err = decodeStruct(r, typ.Type, field.Addr().Interface()); err != nil {
				return body, err
			}
		}

		if queryTag := typ.Tag.Get("query"); queryTag != "" {
			if err := decodeQuery(field, typ.Type, query, queryTag); err != nil {
				return body, err
			}
		}

		//if pathTag := typ.Tag.Get("path"); pathTag != "" {
		//	if err := decodePath(field, typ.Type, vars, pathTag); err != nil {
		//		return body, err
		//	}
		//}

		bodyTag := typ.Tag.Get("body")
		if bodyTag != "" {
			body = true
			if err := decodeBody(r, field.Addr().Interface()); err != nil {
				return body, err
			}
		}
	}
	return body, nil
}

func decodeQuery(field reflect.Value, typ reflect.Type, query url.Values, tag string) error {
	parts := strings.Split(tag, ",")
	if query.Has(parts[0]) {
		if field.Kind() == reflect.Slice {
			var explode bool
			for _, p := range parts[1:] {
				if p == "explode" {
					explode = true
				}
			}

			var value []string
			if explode {
				value = query[parts[0]]
			} else {
				value = strings.Split(query.Get(parts[0]), ",")
			}

			if err := resolveValues(field, typ, value); err != nil {
				return err
			}
			return nil
		}
		if err := resolveValue(field, typ, query.Get(parts[0])); err != nil {
			return err
		}
	}
	return nil
}

func decodePath(field reflect.Value, typ reflect.Type, vars map[string]string, tag string) error {
	if path, ok := vars[tag]; ok {
		if err := resolveValue(field, typ, path); err != nil {
			return err
		}
	}
	return nil
}

func decodeBody(r *http.Request, data interface{}) error {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	if len(b) == 0 {
		return nil
	}

	switch r.Header.Get("Content-Type") {
	case "application/json":
		err := json.Unmarshal(b, &data)
		if err != nil {
			return err
		}
	}

	return nil
}

// resolveValues iterates over string values to resolve a slice value on the field
func resolveValues(field reflect.Value, typ reflect.Type, values []string) error {
	r := reflect.MakeSlice(typ, len(values), len(values))
	for i, value := range values {
		if err := resolveValue(r.Index(i), typ, value); err != nil {
			return err
		}
	}
	field.Set(reflect.ValueOf(r.Interface()))
	return nil
}

// resolveValue resolves and sets the string value to appropriate type on the field
func resolveValue(field reflect.Value, typ reflect.Type, value string) error {
	if field.Kind() == reflect.Pointer {
		v, err := resolve(reflect.New(typ.Elem()).Elem().Interface(), value)
		if err != nil {
			return err
		}

		field.Set(reflect.New(typ.Elem()))
		field.Elem().Set(reflect.ValueOf(v))
		return nil
	}
	v, err := resolve(field.Interface(), value)
	if err != nil {
		return err
	}
	field.Set(reflect.ValueOf(v))
	return nil
}

// resolve the string value to the proper type and return the value
func resolve(t interface{}, v string) (interface{}, error) {
	switch t.(type) {
	case string:
		return v, nil
	case bool:
		return strconv.ParseBool(v)
	case time.Time:
		return time.Parse(time.RFC3339, v)
	case time.Duration:
		return time.ParseDuration(v)
	case int:
		i, err := strconv.ParseInt(v, 10, 32)
		return int(i), err
	case int64:
		return strconv.ParseInt(v, 10, 64)
	case int32:
		i, err := strconv.ParseInt(v, 10, 32)
		return int32(i), err
	case int16:
		i, err := strconv.ParseInt(v, 10, 16)
		return int16(i), err
	case int8:
		i, err := strconv.ParseInt(v, 10, 8)
		return int8(i), err
	case float64:
		return strconv.ParseFloat(v, 64)
	case float32:
		i, err := strconv.ParseFloat(v, 32)
		return float32(i), err
	case uint:
		i, err := strconv.ParseUint(v, 10, 32)
		return uint(i), err
	case uint64:
		return strconv.ParseUint(v, 10, 64)
	case uint32:
		i, err := strconv.ParseUint(v, 10, 32)
		return uint32(i), err
	case uint16:
		i, err := strconv.ParseUint(v, 10, 16)
		return uint16(i), err
	case uint8:
		i, err := strconv.ParseUint(v, 10, 8)
		return uint8(i), err
	case complex128:
		return strconv.ParseComplex(v, 128)
	case complex64:
		i, err := strconv.ParseComplex(v, 64)
		return complex64(i), err
	default:
		return nil, fmt.Errorf("unsupported type: %v", reflect.TypeOf(t))
	}
}
