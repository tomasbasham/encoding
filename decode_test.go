package encoding_test

import (
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/tomasbasham/encoding"
)

func TestUnmarshal_Primitives(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   []byte
		target  any
		want    any
		wantErr bool
	}{
		{
			name:   "bool true",
			input:  []byte("true"),
			target: new(bool),
			want:   pointerTo(true),
		},
		{
			name:   "bool false",
			input:  []byte("false"),
			target: new(bool),
			want:   pointerTo(false),
		},
		{
			name:   "string",
			input:  []byte("hello+world"),
			target: new(string),
			want:   pointerTo("hello world"),
		},
		{
			name:   "string with special chars",
			input:  []byte("hello+%26+world%3F"),
			target: new(string),
			want:   pointerTo("hello & world?"),
		},
		{
			name:   "empty string",
			input:  []byte(""),
			target: new(string),
			want:   pointerTo(""),
		},
		{
			name:   "int",
			input:  []byte("42"),
			target: new(int),
			want:   pointerTo(42),
		},
		{
			name:   "negative int",
			input:  []byte("-42"),
			target: new(int),
			want:   pointerTo(-42),
		},
		{
			name:   "int8",
			input:  []byte("8"),
			target: new(int8),
			want:   pointerTo(int8(8)),
		},
		{
			name:   "int16",
			input:  []byte("16"),
			target: new(int16),
			want:   pointerTo(int16(16)),
		},
		{
			name:   "int32 (rune)",
			input:  []byte("65"),
			target: new(int32),
			want:   pointerTo(int32(65)),
		},
		{
			name:   "int64",
			input:  []byte("64"),
			target: new(int64),
			want:   pointerTo(int64(64)),
		},
		{
			name:    "invalid int",
			input:   []byte("notanumber"),
			target:  new(int),
			wantErr: true,
		},
		{
			name:   "uint",
			input:  []byte("42"),
			target: new(uint),
			want:   pointerTo(uint(42)),
		},
		{
			name:   "uint8 (byte)",
			input:  []byte("8"),
			target: new(uint8),
			want:   pointerTo(uint8(8)),
		},
		{
			name:   "uint16",
			input:  []byte("16"),
			target: new(uint16),
			want:   pointerTo(uint16(16)),
		},
		{
			name:   "uint32",
			input:  []byte("32"),
			target: new(uint32),
			want:   pointerTo(uint32(32)),
		},
		{
			name:   "uint64",
			input:  []byte("64"),
			target: new(uint64),
			want:   pointerTo(uint64(64)),
		},
		{
			name:    "invalid uint",
			input:   []byte("notanumber"),
			target:  new(uint),
			wantErr: true,
		},
		{
			name:   "float32",
			input:  []byte("3.14"),
			target: new(float32),
			want:   pointerTo(float32(3.14)),
		},
		{
			name:   "float64",
			input:  []byte("3.14159"),
			target: new(float64),
			want:   pointerTo(3.14159),
		},
		{
			name:    "invalid float",
			input:   []byte("notanumber"),
			target:  new(float64),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := encoding.Unmarshal(tt.input, tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if diff := diff(tt.want, tt.target); diff != "" {
					t.Errorf("Unmarshal() mismatch %s", diff)
				}
			}
		})
	}
}

func TestUnmarshal_CompositeTypes(t *testing.T) {
	t.Parallel()

	baseTime := time.Date(2025, 2, 8, 0, 0, 0, 0, time.UTC)
	optionalVal := "optional_value"

	tests := []struct {
		name    string
		input   []byte
		target  any
		want    any
		wantErr bool
	}{
		{
			name:   "pointer to int",
			input:  []byte("42"),
			target: new(*int),
			want:   pointerTo(pointerTo(42)),
		},
		{
			name:   "nil pointer",
			input:  []byte(""),
			target: new(*int),
			want:   new(*int),
		},
		{
			name:   "slice of ints",
			input:  []byte("1&2&3"),
			target: new([]int),
			want:   &[]int{1, 2, 3},
		},
		{
			name:   "slice of strings",
			input:  []byte("a&b&c"),
			target: new([]string),
			want:   &[]string{"a", "b", "c"},
		},
		{
			name:   "slice with special chars",
			input:  []byte("a%26b&c%3Dd&e%3Ff"),
			target: new([]string),
			want:   &[]string{"a&b", "c=d", "e?f"},
		},
		{
			name:   "empty slice",
			input:  []byte(""),
			target: new([]int),
			want:   &[]int{},
		},
		{
			name:   "map with string keys and int values",
			input:  []byte("a=1&b=2&c=3"),
			target: new(map[string]int),
			want:   &map[string]int{"a": 1, "b": 2, "c": 3},
		},
		{
			name:   "map with string keys and string values",
			input:  []byte("a=x&b=y&c=z"),
			target: new(map[string]string),
			want:   &map[string]string{"a": "x", "b": "y", "c": "z"},
		},
		{
			name:    "map with non-string key",
			input:   []byte("1=1&2=2&3=3"),
			target:  new(map[int]int),
			wantErr: true,
		},
		{
			name:   "empty map",
			input:  []byte(""),
			target: new(map[string]int),
			want:   &map[string]int{},
		},
		{
			name: "basic form",
			input: valuesToBytes(url.Values{
				"age":     {"20"},
				"aliases": {"johnny", "jonny"},
				"name":    {"john"},
			}),
			target: &BasicForm{},
			want: &BasicForm{
				Name:    "john",
				Aliases: []string{"johnny", "jonny"},
				Age:     20,
			},
		},
		{
			name: "complex form with custom type",
			input: valuesToBytes(url.Values{
				"age":        {"25"},
				"aliases":    {"janet", "jan"},
				"created_at": {"2025.02.08"},
				"id":         {"1"},
				"name":       {"jane"},
				"optional":   {"optional_value"},
			}),
			target: &ComplexForm{},
			want: &ComplexForm{
				ID:        1,
				Name:      "jane",
				Aliases:   []string{"janet", "jan"},
				Age:       25,
				CreatedAt: MyDate(baseTime),
				Optional:  &optionalVal,
			},
		},
		{
			name:    "non-pointer target",
			input:   []byte{},
			target:  BasicForm{},
			wantErr: true,
		},
		{
			name:    "nil target",
			input:   []byte{},
			target:  nil,
			wantErr: true,
		},
		{
			name: "ignored fields",
			input: valuesToBytes(url.Values{
				"Empty":   {"value"},
				"NoTag":   {"value"},
				"Omitted": {"present"},
				"complex": {"2025.02.08"},
				"ignored": {"skip"},
				"private": {"hidden"},
				"public":  {"visible"},
			}),
			target: &IgnoredFieldsForm{},
			want: &IgnoredFieldsForm{
				Public:  "visible",
				NoTag:   "value",
				Empty:   "value",
				Omitted: "present",
				Complex: MyDate(baseTime),
			},
		},
		{
			name: "invalid type conversion",
			input: valuesToBytes(url.Values{
				"age": {"not_a_number"},
			}),
			target:  &BasicForm{},
			wantErr: true,
		},
		{
			name: "invalid custom type format",
			input: valuesToBytes(url.Values{
				"created_at": {"invalid_date"},
			}),
			target:  &ComplexForm{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := encoding.Unmarshal(tt.input, tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if diff := diff(tt.want, tt.target); diff != "" {
					t.Errorf("Unmarshal() mismatch %s", diff)
				}
			}
		})
	}
}

func BenchmarkUnmarshal(b *testing.B) {
	benchmarks := []struct {
		name   string
		input  []byte
		target any
	}{
		{
			name:   "bool",
			input:  []byte("true"),
			target: new(bool),
		},
		{
			name:   "string",
			input:  []byte("hello+world"),
			target: new(string),
		},
		{
			name:   "int",
			input:  []byte("42"),
			target: new(int),
		},
		{
			name:   "float64",
			input:  []byte("3.14159"),
			target: new(float64),
		},
		{
			name:   "small int slice",
			input:  []byte("1&2&3&4&5"),
			target: new([]int),
		},
		{
			name:   "medium int slice",
			input:  generateIntSliceBytes(100),
			target: new([]int),
		},
		{
			name:   "large int slice",
			input:  generateIntSliceBytes(1000),
			target: new([]int),
		},
		{
			name:   "string slice",
			input:  []byte("a&b&c&d&e"),
			target: new([]string),
		},
		{
			name: "small map",
			input: valuesToBytes(url.Values{
				"a": {"1"},
				"b": {"2"},
				"c": {"3"},
			}),
			target: new(map[string]int),
		},
		{
			name:   "medium map",
			input:  generateMapBytes(100),
			target: new(map[string]int),
		},
		{
			name:   "large map",
			input:  generateMapBytes(1000),
			target: new(map[string]int),
		},
		{
			name: "basic form",
			input: valuesToBytes(url.Values{
				"age":     {"20"},
				"aliases": {"johnny", "jonny"},
				"name":    {"john"},
			}),
			target: &BasicForm{},
		},
		{
			name: "complex form",
			input: valuesToBytes(url.Values{
				"age":        {"25"},
				"aliases":    {"janet", "jan"},
				"created_at": {"2025.02.08"},
				"id":         {"1"},
				"name":       {"jane"},
				"optional":   {"optional_value"},
			}),
			target: &ComplexForm{},
		},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				encoding.Unmarshal(bm.input, bm.target)
			}
		})
	}
}

func generateIntSliceBytes(count int) []byte {
	var values []string
	for i := range count {
		values = append(values, strconv.Itoa(i))
	}
	return []byte(strings.Join(values, "&"))
}

func generateMapBytes(count int) []byte {
	values := url.Values{}
	for i := range count {
		key := string(rune('a'+i%26)) + string(rune('0'+i/26)) + string(rune('0'+i/260))
		values[key] = []string{strconv.Itoa(i)}
	}
	return valuesToBytes(values)
}
