package main

import (
	"fmt"

	form "github.com/tomasbasham/encoding"
)

type Panda struct {
	Name string `form:"name"`
	Age  int    `form:"age"`
}

func main() {
	// Primitive types
	bo := false
	b := must(form.Marshal(bo))
	fmt.Println(string(b))

	s := "Alice"
	b = must(form.Marshal(s))
	fmt.Println(string(b))

	i := 5
	b = must(form.Marshal(i))
	fmt.Println(string(b))

	i8 := int8(5)
	b = must(form.Marshal(i8))
	fmt.Println(string(b))

	u := uint(5)
	b = must(form.Marshal(u))
	fmt.Println(string(b))

	u8 := uint8(5)
	b = must(form.Marshal(u8))
	fmt.Println(string(b))

	by := byte(5)
	b = must(form.Marshal(by))
	fmt.Println(string(b))

	r := rune(5)
	b = must(form.Marshal(r))
	fmt.Println(string(b))

	f := 5.0
	b = must(form.Marshal(f))
	fmt.Println(string(b))

	// c := complex(5, 5)
	// b = must(form.Marshal(c))
	// fmt.Println(string(b))

	// Composite types
	a := []int{1, 2, 3}
	b = must(form.Marshal(a))
	fmt.Println(string(b))

	m := map[string]int{"one": 1, "two": 2, "three": 3}
	b = must(form.Marshal(m))
	fmt.Println(string(b))

	p := Panda{Name: "Alice", Age: 5}
	b = must(form.Marshal(p))
	fmt.Println(string(b))

	// t := chan int(nil)
	// b = must(form.Marshal(t))
	// fmt.Println(string(b))
}

func must(b []byte, err error) []byte {
	if err != nil {
		panic(err)
	}
	return b
}
