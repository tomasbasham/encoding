package encoding

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
		return marshalForm(m)
	}

	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return []byte{}, nil
	}

	return marshal(rv)
}

func marshal(v reflect.Value) ([]byte, error) {
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return []byte{}, nil
		}
		v = v.Elem()
	}
	if isStructValue(v) {
		return marshalValue(v, marshalStruct)
	}
	if isMapValue(v) {
		return marshalValue(v, marshalMap)
	}
	return marshalPrimitive(v)
}

func isStructValue(v reflect.Value) bool {
	return v.Kind() == reflect.Struct ||
		(v.Kind() == reflect.Pointer && v.Elem().Kind() == reflect.Struct)
}

func isMapValue(v reflect.Value) bool {
	return v.Kind() == reflect.Map ||
		(v.Kind() == reflect.Pointer && v.Elem().Kind() == reflect.Map)
}

func isCompositeValue(v reflect.Value) bool {
	return isStructValue(v) || isMapValue(v)
}

func marshalForm(m Marshaler) ([]byte, error) {
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

type marshalerFunc func(v reflect.Value) (url.Values, error)

func marshalValue(v reflect.Value, fn marshalerFunc) ([]byte, error) {
	rv := reflect.Indirect(v)
	values, err := fn(rv)
	if err != nil {
		return nil, fmt.Errorf("form: failed to marshal: %w", err)
	}
	return []byte(values.Encode()), nil
}

func marshalStruct(v reflect.Value) (url.Values, error) {
	tags := tags(v)
	data := url.Values{}
	for i := range v.Type().NumField() {
		tag := tags[i]
		if tag.Ignore {
			continue
		}
		key := tag.Name
		if key == "" {
			continue
		}
		fv := v.Field(i)
		if tag.Omit && isEmptyValue(fv) {
			continue
		}
		if val, err := get(fv); err == nil && len(val) > 0 {
			data[key] = val
		}
	}
	return data, nil
}

func marshalMap(v reflect.Value) (url.Values, error) {
	data := url.Values{}

	// Validate map key type - only string keys are supported
	if v.Type().Key().Kind() != reflect.String {
		return nil, fmt.Errorf("form: unsupported map key type: %v", v.Type().Key())
	}

	for _, key := range v.MapKeys() {
		keyStr := key.String()
		mapVal := v.MapIndex(key)
		if isEmptyValue(mapVal) {
			continue
		}
		if val, err := get(mapVal); err == nil && len(val) > 0 {
			data[keyStr] = val
		}
	}
	return data, nil
}

func assertMarshaler(v reflect.Value) (Marshaler, bool) {
	if v.CanAddr() {
		if m, ok := v.Addr().Interface().(Marshaler); ok {
			return m, true
		}
	}
	if m, ok := v.Interface().(Marshaler); ok {
		return m, true
	}
	return nil, false
}

func get(v reflect.Value) ([]string, error) {
	if v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return nil, nil
		}
		v = v.Elem()
	}
	if v.Kind() == reflect.Slice {
		return getSlice(v)
	}
	if m, ok := assertMarshaler(v); ok {
		b, err := m.MarshalForm()
		if err != nil {
			return nil, err
		}
		return []string{string(b)}, nil
	}
	if isCompositeValue(v) {
		values, err := marshal(v)
		if err != nil {
			return nil, err
		}
		return []string{string(values)}, nil
	}
	return []string{getScalar(v)}, nil
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
		reflect.Interface, reflect.Pointer:
		return v.IsZero()
	}
	return false
}
