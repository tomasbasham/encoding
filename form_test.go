package encoding_test

import (
	"fmt"
	"net/url"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// Custom types for testing
type MyDate time.Time

func (d MyDate) MarshalForm() ([]byte, error) {
	return []byte(time.Time(d).Format("2006.01.02")), nil
}

func (d *MyDate) UnmarshalForm(b []byte) error {
	t, err := time.Parse("2006.01.02", string(b))
	if err != nil {
		return err
	}
	*d = MyDate(t)
	return nil
}

// Test structs
type BasicForm struct {
	Name    string   `form:"name"`
	Aliases []string `form:"aliases"`
	Age     int      `form:"age"`
}

type ComplexForm struct {
	ID        int      `form:"id"`
	Name      string   `form:"name"`
	Aliases   []string `form:"aliases,omitempty"`
	Age       int      `form:"age"`
	CreatedAt MyDate   `form:"created_at"`
	Private   string   `form:"-"`
	Optional  *string  `form:"optional,omitempty"`
}

type IgnoredFieldsForm struct {
	Public  string `form:"public"`
	Private string `form:"-"`
	Ignored string `form:",ignore"`
	NoTag   string
	Empty   string `form:""`
	Omitted string `form:",omitempty"`
	Complex MyDate `form:"complex,omitempty"`
}

func diff[T any](a, b T) string {
	if diff := cmp.Diff(a, b, cmpopts.EquateComparable(MyDate{})); diff != "" {
		return fmt.Sprintf("(-want +got):\n%s", diff)
	}
	return ""
}

func pointerTo[T any](v T) *T {
	return &v
}

func valuesToBytes(values url.Values) []byte {
	return []byte(values.Encode())
}
