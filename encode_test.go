package encoding_test

import (
	"net/url"
	"testing"
	"time"

	"github.com/tomasbasham/encoding"
)

func TestMarshal_Primitives(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     any
		want      []byte
		wantErr   bool
		wantPanic bool
	}{
		{
			name:  "bool true",
			input: true,
			want:  []byte("true"),
		},
		{
			name:  "bool false",
			input: false,
			want:  []byte("false"),
		},
		{
			name:  "string",
			input: "hello world",
			want:  []byte("hello+world"),
		},
		{
			name:  "string with special chars",
			input: "hello & world?",
			want:  []byte("hello+%26+world%3F"),
		},
		{
			name:  "int",
			input: 42,
			want:  []byte("42"),
		},
		{
			name:  "negative int",
			input: -42,
			want:  []byte("-42"),
		},
		{
			name:  "int8",
			input: int8(8),
			want:  []byte("8"),
		},
		{
			name:  "int16",
			input: int16(16),
			want:  []byte("16"),
		},
		{
			name:  "int32 (rune)",
			input: 'A',
			want:  []byte("65"),
		},
		{
			name:  "int64",
			input: int64(64),
			want:  []byte("64"),
		},
		{
			name:  "uint",
			input: uint(42),
			want:  []byte("42"),
		},
		{
			name:  "uint8 (byte)",
			input: byte(8),
			want:  []byte("8"),
		},
		{
			name:  "uint16",
			input: uint16(16),
			want:  []byte("16"),
		},
		{
			name:  "uint32",
			input: uint32(32),
			want:  []byte("32"),
		},
		{
			name:  "uint64",
			input: uint64(64),
			want:  []byte("64"),
		},
		{
			name:  "float32",
			input: float32(3.14),
			want:  []byte("3.14"),
		},
		{
			name:  "float64",
			input: 3.14159,
			want:  []byte("3.14159"),
		},
		{
			name:      "complex64",
			input:     complex64(1),
			wantPanic: true,
		},
		{
			name:      "complex128",
			input:     complex128(1),
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				r := recover()
				if (r != nil) != tt.wantPanic {
					t.Errorf("Marshal() panic = %v, wantPanic %v", r, tt.wantPanic)
				}
			}()

			got, err := encoding.Marshal(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if diff := diff(tt.want, got); diff != "" {
					t.Errorf("Marshal() mismatch %s", diff)
				}
			}
		})
	}
}

func TestMarshal_CompositeTypes(t *testing.T) {
	t.Parallel()

	baseTime := time.Date(2025, 2, 8, 0, 0, 0, 0, time.UTC)
	optionalVal := "optional_value"

	tests := []struct {
		name      string
		input     any
		want      []byte
		wantErr   bool
		wantPanic bool
	}{
		{
			name:      "chan int",
			input:     make(chan int),
			wantPanic: true,
		},
		{
			name:      "func",
			input:     func() {},
			wantPanic: true,
		},
		{
			name:  "pointer to int",
			input: pointerTo(42),
			want:  []byte("42"),
		},
		{
			name:  "nil pointer",
			input: (*int)(nil),
			want:  []byte(""),
		},
		{
			name:  "nil interface",
			input: nil,
			want:  []byte(""),
		},
		{
			name:  "slice of ints",
			input: []int{1, 2, 3},
			want:  []byte("1&2&3"),
		},
		{
			name:  "slice of strings",
			input: []string{"a", "b", "c"},
			want:  []byte("a&b&c"),
		},
		{
			name:  "slice with special chars",
			input: []string{"a&b", "c=d", "e?f"},
			want:  []byte("a%26b&c%3Dd&e%3Ff"),
		},
		{
			name:  "empty slice",
			input: []int{},
			want:  []byte(""),
		},
		{
			name:  "nil slice",
			input: []int(nil),
			want:  []byte(""),
		},
		{
			name:  "map with string keys and int values",
			input: map[string]int{"a": 1, "b": 2, "c": 3},
			want:  valuesToBytes(url.Values{"a": {"1"}, "b": {"2"}, "c": {"3"}}),
		},
		{
			name:  "map with string keys and string values",
			input: map[string]string{"a": "x", "b": "y", "c": "z"},
			want:  valuesToBytes(url.Values{"a": {"x"}, "b": {"y"}, "c": {"z"}}),
		},
		{
			name:  "map with string keys and any values",
			input: map[string]any{"a": 1, "b": "string", "c": true, "d": 3.14},
			want:  valuesToBytes(url.Values{"a": {"1"}, "b": {"string"}, "c": {"true"}, "d": {"3.14"}}),
		},
		{
			name:    "map with non-string key",
			input:   map[int]int{1: 1, 2: 2, 3: 3},
			wantErr: true,
		},
		{
			name:  "empty map",
			input: map[string]int{},
			want:  []byte(""),
		},
		{
			name:  "nil map",
			input: map[string]int(nil),
			want:  []byte(""),
		},
		{
			name:  "map with nested struct",
			input: map[string]any{"user": BasicForm{Name: "john", Age: 20, Aliases: []string{"j"}}},
			want:  valuesToBytes(url.Values{"user": {"age=20&aliases=j&name=john"}}),
		},
		{
			name:  "map with nested map",
			input: map[string]any{"data": map[string]int{"x": 1, "y": 2}},
			want:  valuesToBytes(url.Values{"data": {"x=1&y=2"}}),
		},
		{
			name: "basic form",
			input: BasicForm{
				Name:    "john",
				Aliases: []string{"johnny", "jonny"},
				Age:     20,
			},
			want: valuesToBytes(url.Values{
				"age":     {"20"},
				"aliases": {"johnny", "jonny"},
				"name":    {"john"},
			}),
		},
		{
			name: "complex form with custom type",
			input: ComplexForm{
				ID:        1,
				Name:      "jane",
				Aliases:   []string{"janet", "jan"},
				Age:       25,
				CreatedAt: MyDate(baseTime),
				Private:   "hidden",
				Optional:  &optionalVal,
			},
			want: valuesToBytes(url.Values{
				"age":        {"25"},
				"aliases":    {"janet", "jan"},
				"created_at": {"2025.02.08"},
				"id":         {"1"},
				"name":       {"jane"},
				"optional":   {"optional_value"},
			}),
		},
		{
			name: "form with ignored fields",
			input: IgnoredFieldsForm{
				Public:  "visible",
				Private: "hidden",
				Ignored: "skip",
				NoTag:   "value",
				Empty:   "value",
				Omitted: "",
				Complex: MyDate(baseTime),
			},
			want: valuesToBytes(url.Values{
				"Empty":   {"value"},
				"NoTag":   {"value"},
				"complex": {"2025.02.08"},
				"public":  {"visible"},
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			defer func() {
				r := recover()
				if (r != nil) != tt.wantPanic {
					t.Errorf("Marshal() panic = %v, wantPanic %v", r, tt.wantPanic)
				}
			}()

			got, err := encoding.Marshal(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if diff := diff(tt.want, got); diff != "" {
					t.Errorf("Marshal() mismatch %s", diff)
				}
			}
		})
	}
}

func BenchmarkMarshal(b *testing.B) {
	baseTime := time.Date(2025, 2, 8, 0, 0, 0, 0, time.UTC)
	optionalVal := "optional_value"

	benchmarks := []struct {
		name  string
		input any
	}{
		{
			name:  "bool",
			input: true,
		},
		{
			name:  "string",
			input: "hello world",
		},
		{
			name:  "int",
			input: 42,
		},
		{
			name:  "float64",
			input: 3.14159,
		},
		{
			name:  "small int slice",
			input: []int{1, 2, 3, 4, 5},
		},
		{
			name:  "medium int slice",
			input: make([]int, 100),
		},
		{
			name:  "large int slice",
			input: make([]int, 1000),
		},
		{
			name:  "string slice",
			input: []string{"a", "b", "c", "d", "e"},
		},
		{
			name:  "small map",
			input: map[string]int{"a": 1, "b": 2, "c": 3},
		},
		{
			name: "medium map",
			input: func() map[string]int {
				m := make(map[string]int)
				for i := range 100 {
					m[string(rune('a'+i%26))+string(rune('0'+i/26))] = i
				}
				return m
			}(),
		},
		{
			name: "large map",
			input: func() map[string]int {
				m := make(map[string]int)
				for i := range 1000 {
					m[string(rune('a'+i%26))+string(rune('0'+i/26))+string(rune('0'+i/260))] = i
				}
				return m
			}(),
		},
		{
			name: "basic form",
			input: BasicForm{
				Name:    "john",
				Aliases: []string{"johnny", "jonny"},
				Age:     20,
			},
		},
		{
			name: "complex form with custom type",
			input: ComplexForm{
				ID:        1,
				Name:      "jane",
				Aliases:   []string{"janet", "jan"},
				Age:       25,
				CreatedAt: MyDate(baseTime),
				Private:   "hidden",
				Optional:  &optionalVal,
			},
		},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				encoding.Marshal(bm.input)
			}
		})
	}
}
