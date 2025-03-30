package encoding // import "github.com/tomasbasham/encoding/form"

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

// Marshaler is the interface implemented by types that can marshal themselves
// into a form description.
type Marshaler interface {
	MarshalForm() ([]byte, error)
}

// Marshal returns the form encoding of v.
func Marshal(v any) ([]byte, error) {
	if m, ok := v.(Marshaler); ok {
		return marshalMarshaler(m)
	}

	rv := reflect.ValueOf(v)
	if isStructValue(rv) {
		return marshalStruct(rv)
	}

	return marshalPrimitive(rv)
}

func marshalMarshaler(m Marshaler) ([]byte, error) {
	b, err := m.MarshalForm()
	if err != nil {
		return nil, fmt.Errorf("form: failed to marshal: %w", err)
	}
	values, err := url.ParseQuery(string(b))
	if err != nil {
		return nil, fmt.Errorf("form: invalid form data: %w", err)
	}
	return []byte(values.Encode()), nil
}

func isStructValue(v reflect.Value) bool {
	return v.Kind() == reflect.Struct ||
		(v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct)
}

func marshalPrimitive(v reflect.Value) ([]byte, error) {
	values, err := get(v)
	if err != nil {
		return nil, fmt.Errorf("form: failed to marshal: %w", err)
	}
	if len(values) == 1 {
		return []byte(url.QueryEscape(values[0])), nil
	} else if len(values) > 1 {
		result := make([]string, len(values))
		for i, val := range values {
			result[i] = url.QueryEscape(val)
		}
		return []byte(strings.Join(result, "&")), nil
	}
	return []byte{}, nil
}

func marshalStruct(v reflect.Value) ([]byte, error) {
	values, err := marshal(v)
	if err != nil {
		return nil, fmt.Errorf("form: failed to marshal: %w", err)
	}
	return []byte(values.Encode()), nil
}

func marshal(v reflect.Value) (url.Values, error) {
	tags := tags(v)
	rv := reflect.Indirect(v)
	data := url.Values{}
	for i := range rv.Type().NumField() {
		fv := rv.Field(i)
		tag := tags[i]
		if tag.Ignore {
			continue
		}
		key := tag.Name
		if key == "" {
			continue
		}
		if tag.Omit && isEmptyValue(fv) {
			continue
		}
		if val, err := get(fv); err == nil && len(val) > 0 {
			data[key] = val
		}
	}
	return data, nil
}

func assertMarshaler(fv reflect.Value) (Marshaler, bool) {
	if fv.CanAddr() {
		if m, ok := fv.Addr().Interface().(Marshaler); ok {
			return m, true
		}
	}
	if m, ok := fv.Interface().(Marshaler); ok {
		return m, true
	}
	return nil, false
}

func get(fv reflect.Value) ([]string, error) {
	if fv.Kind() == reflect.Ptr {
		if fv.IsNil() {
			return nil, nil
		}
		fv = fv.Elem()
	}
	if fv.Kind() == reflect.Slice {
		return getSlice(fv)
	}
	if m, ok := assertMarshaler(fv); ok {
		b, err := m.MarshalForm()
		if err != nil {
			return nil, err
		}
		return []string{string(b)}, nil
	}
	return []string{getScalar(fv)}, nil
}

func getSlice(v reflect.Value) ([]string, error) {
	values := make([]string, v.Len())
	for i := range v.Len() {
		elem := v.Index(i)
		if m, ok := assertMarshaler(elem); ok {
			b, err := m.MarshalForm()
			if err != nil {
				return nil, err
			}
			values[i] = string(b)
			continue
		}
		values[i] = getScalar(elem)
	}
	return values, nil
}

func getScalar(v reflect.Value) string {
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, v.Type().Bits())
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	default:
		panic("form: unsupported type: " + v.Type().String())
	}
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.Interface, reflect.Ptr:
		return v.IsZero()
	}
	return false
}
