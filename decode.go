package encoding

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

// An InvalidUnmarshalError describes an invalid argument passed to [Unmarshal].
// (The argument to [Unmarshal] must be a non-nil pointer.)
type InvalidUnmarshalError struct {
	Type reflect.Type
}

func (e *InvalidUnmarshalError) Error() string {
	if e.Type == nil {
		return "form: Unmarshal(nil)"
	}

	if e.Type.Kind() != reflect.Pointer {
		return "form: Unmarshal(non-pointer " + e.Type.String() + ")"
	}
	return "form: Unmarshal(nil " + e.Type.String() + ")"
}

// Unmarshaler is the interface implemented by types that can unmarshal a form
// description of themselves. The input can be assumed to be a valid encoding of
// a form value. UnmarshalForm must copy the form data if it wishes to retain
// the data after returning.
type Unmarshaler interface {
	UnmarshalForm([]byte) error
}

// Unmarshal parses the form data and stores the result in the value pointed to
// by v. If v is nil or not a pointer, Unmarshal returns an InvalidValueError.
func Unmarshal(data []byte, v any) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Pointer || val.IsNil() {
		return &InvalidUnmarshalError{reflect.TypeOf(v)}
	}
	if u, ok := v.(Unmarshaler); ok {
		return u.UnmarshalForm(data)
	}

	return unmarshal(data, val)
}

func unmarshal(data []byte, v reflect.Value) error {
	if !isCompositePointer(v) {
		return unmarshalPrimitive(data, v)
	}

	values, err := url.ParseQuery(string(data))
	if err != nil {
		return fmt.Errorf("form: invalid form data: %w", err)
	}

	if isStructPointer(v) {
		return unmarshalValue(values, v, unmarshalStruct)
	}
	if isMapPointer(v) {
		return unmarshalValue(values, v, unmarshalMap)
	}
	return nil
}

func isStructPointer(v reflect.Value) bool {
	return v.Kind() == reflect.Pointer && v.Elem().Kind() == reflect.Struct
}

func isMapPointer(v reflect.Value) bool {
	return v.Kind() == reflect.Pointer && v.Elem().Kind() == reflect.Map
}

func isCompositePointer(v reflect.Value) bool {
	return isStructPointer(v) || isMapPointer(v)
}

func unmarshalPrimitive(data []byte, v reflect.Value) error {
	if len(data) == 0 {
		return nil
	}

	unescaped, err := url.QueryUnescape(string(data))
	if err != nil {
		return fmt.Errorf("form: invalid form data: %w", err)
	}

	var allValues []string
	if strings.Contains(unescaped, "&") {
		parts := strings.Split(string(data), "&")
		allValues = make([]string, len(parts))
		for _, part := range parts {
			val, err := url.QueryUnescape(part)
			if err != nil {
				return fmt.Errorf("form: invalid form data: %w", err)
			}
			allValues = append(allValues, val)
		}
	} else {
		allValues = []string{unescaped}
	}

	return set(v.Elem(), allValues)
}

type unmarshalerFunc func(url.Values, reflect.Value) error

func unmarshalValue(data url.Values, v reflect.Value, fn unmarshalerFunc) error {
	rv := reflect.Indirect(v)
	err := fn(data, rv)
	if err != nil {
		return fmt.Errorf("form: failed to unmarshal: %w", err)
	}
	return nil
}

func unmarshalStruct(data url.Values, v reflect.Value) error {
	tags := tags(v)
	for i := range v.Type().NumField() {
		fv := v.Field(i)
		if !fv.CanSet() {
			continue
		}
		tag := tags[i]
		if tag.Ignore {
			continue
		}
		key := tag.Name
		if key == "-" {
			continue
		}
		if val, ok := data[key]; ok {
			if err := set(fv, val); err != nil {
				return fmt.Errorf("form: failed to set field %s: %w", tag.Name, err)
			}
		}
	}
	return nil
}

func unmarshalMap(data url.Values, v reflect.Value) error {
	if v.IsNil() {
		v.Set(reflect.MakeMap(v.Type()))
	}

	if v.Type().Key().Kind() != reflect.String {
		return fmt.Errorf("form: unsupported map key type: %v", v.Type().Key())
	}

	elemType := v.Type().Elem()
	for key, values := range data {
		elemValue := reflect.New(elemType).Elem()
		if err := set(elemValue, values); err != nil {
			return fmt.Errorf("form: failed to set map value for key %s: %w", key, err)
		}

		v.SetMapIndex(reflect.ValueOf(key), elemValue)
	}

	return nil
}

func assertUnmarshaler(v reflect.Value) (Unmarshaler, bool) {
	if v.CanAddr() {
		if u, ok := v.Addr().Interface().(Unmarshaler); ok {
			return u, true
		}
	}
	if u, ok := v.Interface().(Unmarshaler); ok {
		return u, true
	}
	return nil, false
}

func set(fv reflect.Value, val []string) error {
	if fv.Kind() == reflect.Ptr {
		if fv.IsNil() {
			fv.Set(reflect.New(fv.Type().Elem()))
		}
		fv = fv.Elem()
	}
	if fv.Kind() == reflect.Slice {
		return setSlice(fv, val)
	}
	if len(val) == 0 {
		return nil
	}
	if u, ok := assertUnmarshaler(fv); ok {
		return u.UnmarshalForm([]byte(val[0]))
	}
	return setScalar(fv, val[0])
}

func setSlice(fv reflect.Value, val []string) error {
	if fv.IsNil() || fv.Len() != len(val) {
		fv.Set(reflect.MakeSlice(fv.Type(), len(val), len(val)))
	}

	for i, v := range val {
		elem := fv.Index(i)
		if u, ok := assertUnmarshaler(elem); ok {
			if err := u.UnmarshalForm([]byte(v)); err != nil {
				return fmt.Errorf("failed to set slice element %d: %w", i, err)
			}
			continue
		}
		if err := setScalar(elem, v); err != nil {
			return fmt.Errorf("failed to set slice element %d: %w", i, err)
		}
	}
	return nil
}

func setScalar(v reflect.Value, val string) error {
	switch v.Kind() {
	case reflect.String:
		v.SetString(val)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := parseInt(val, v.Type().Bits())
		if err != nil {
			return fmt.Errorf("setScalar: %w", err)
		}
		v.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := parseUint(val, v.Type().Bits())
		if err != nil {
			return fmt.Errorf("setScalar: %w", err)
		}
		v.SetUint(i)
	case reflect.Float32, reflect.Float64:
		f, err := parseFloat(val, v.Type().Bits())
		if err != nil {
			return fmt.Errorf("setScalar: %w", err)
		}
		v.SetFloat(f)
	case reflect.Bool:
		b, err := parseBool(val)
		if err != nil {
			return fmt.Errorf("setScalar: %w", err)
		}
		v.SetBool(b)
	}
	return nil
}

func parseInt(s string, bitSize int) (int64, error) {
	return strconv.ParseInt(s, 10, bitSize)
}

func parseUint(s string, bitSize int) (uint64, error) {
	return strconv.ParseUint(s, 10, bitSize)
}

func parseFloat(s string, bitSize int) (float64, error) {
	return strconv.ParseFloat(s, bitSize)
}

func parseBool(s string) (bool, error) {
	return strconv.ParseBool(s)
}
