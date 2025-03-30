package encoding

import (
	"reflect"
	"strings"
)

type tag struct {
	Name   string
	Omit   bool
	Ignore bool
}

func tags(fv reflect.Value) []*tag {
	tt := reflect.Indirect(fv).Type()
	if tt.Kind() != reflect.Struct {
		return []*tag{}
	}

	// Create a slice of tags to store the tags for each field on the struct.  The
	// length of the slice is equal to the number of fields on the struct.
	tags := make([]*tag, tt.NumField())

	// Look for a Field on the struct that matches the key name.
	for i := range tt.NumField() {
		f := tt.Field(i)
		tag := parseTag(f.Tag.Get("form"))
		if !tag.Ignore && tag.Name == "" {
			tag.Name = f.Name
		}
		tags[i] = tag
	}
	return tags
}

func parseTag(str string) *tag {
	if str == "-" {
		return &tag{Ignore: true}
	}

	// Split the tag into parts. Although it should never be the case that a
	// tag contains zero parts, we should handle this case gracefully.
	parts := strings.Split(str, ",")
	if len(parts) == 0 {
		return &tag{Ignore: true}
	}

	t := &tag{}

	// The first part of the tag is the name of the field. If the first part is
	// a hyphen, then the field should be ignored.
	switch n := parts[0]; n {
	case "-":
		t.Ignore = true
	default:
		t.Name = n
	}

	// The remaining parts of the tag are flags that modify the behaviour of the
	// field.
	for _, p := range parts[1:] {
		switch p {
		case "omitempty":
			t.Omit = true
		case "ignore":
			t.Ignore = true
		}
	}

	return t
}
