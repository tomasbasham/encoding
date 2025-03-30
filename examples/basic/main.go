package main

import (
	"fmt"
	"net/url"
	"reflect"

	"github.com/tomasbasham/encoding"
)

type Panda struct {
	Name     string  `form:"name"`
	Species  string  `form:"species"`
	Age      int     `form:"age"`
	Children []Panda `form:"children"`
}

func main() {
	// primitives()
	compositeTypes()
}

func primitives() {
	// Booleans
	marshalUnmarshal(true)
	marshalUnmarshal(false)

	// Strings
	marshalUnmarshal("hello world")
	marshalUnmarshal("hello & world?")

	// Integers
	marshalUnmarshal(42)
	marshalUnmarshal(-42)
	marshalUnmarshal(int8(8))
	marshalUnmarshal(int16(16))
	marshalUnmarshal(int32(32))
	marshalUnmarshal(int64(64))
	marshalUnmarshal(uint(8))
	marshalUnmarshal(uint8(8))
	marshalUnmarshal(uint16(16))
	marshalUnmarshal(uint32(32))
	marshalUnmarshal(uint64(64))

	// Floats
	marshalUnmarshal(3.14)
	marshalUnmarshal(float32(3.14))
	marshalUnmarshal(float64(3.14159))

	// Unsupported types
	marshalUnmarshal(complex64(1))
	marshalUnmarshal(complex128(1))

}

func compositeTypes() {
	// Pointers
	marshalUnmarshal(toPointer(42))
	marshalUnmarshal((*int)(nil))

	// Slices
	marshalUnmarshal([]int{1, 2, 3})
	marshalUnmarshal([]string{"a", "b", "c"})
	marshalUnmarshal([]string{"a&b", "c=d", "e?f"})
	marshalUnmarshal([]int{})

	// Maps
	marshalUnmarshal(map[string]int{"a": 1, "b": 2, "c": 3})
	marshalUnmarshal(map[string]any{"a": 2, "b": "string", "c": true, "d": 3.14})
	marshalUnmarshal(map[string]any{"a": 3, "b": map[string]int{"c": 4}})

	// URL Values
	marshalUnmarshal(url.Values{"a": []string{"1", "2"}, "b": []string{"3"}})

	// Structs
	marshalUnmarshal(Panda{Name: "Panda", Species: "Ailuropoda melanoleuca", Age: 5})
	marshalUnmarshal(Panda{Name: "Panda", Species: "Ailuropoda melanoleuca", Age: 5, Children: []Panda{{Name: "Baby Panda"}}})

	// Unsupported types
	marshalUnmarshal(make(chan int))
	marshalUnmarshal(func() {})
}

func marshalUnmarshal[T any](v T) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered in marshalUnmarshal: %v\n", r)
		}
	}()

	val, typeName := getValueAndType(v)
	fmt.Printf("\nMarshaling value: (%s) %+v\n", typeName, val)

	b, err := encoding.Marshal(v)
	if err != nil {
		fmt.Printf("Error marshaling: %v\n", err)
		return
	}

	fmt.Printf("Marshaled value: ([]byte) %s\n", b)

	uv := new(T)
	err = encoding.Unmarshal(b, uv)
	if err != nil {
		fmt.Printf("Error unmarshaling: %v\n", err)
		return
	}

	val, typeName = getValueAndType(uv)
	fmt.Printf("Unmarshaled value: (%s) %+v\n", typeName, val)
}

func toPointer[T any](v T) *T {
	return &v
}

func getValueAndType(v any) (any, string) {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	if !val.IsValid() {
		return nil, "nil"
	}

	// Dereference pointers to get the underlying value.
	typeName := val.Type().String()
	for val.Kind() == reflect.Pointer {
		val = val.Elem()
		if val.IsValid() {
			typeName = val.Type().String()
		}
	}

	if !val.IsValid() {
		return nil, typeName
	}

	return val.Interface(), typeName
}
