package encoding_test

import (
	"bytes"
	"net/url"
	"strings"
	"testing"

	"github.com/tomasbasham/encoding"
)

func TestDecoder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		target  any
		want    any
		wantErr bool
	}{
		{
			name:   "valid query string",
			input:  "name=john&age=20&aliases=johnny&aliases=jonny",
			target: &BasicForm{},
			want: &BasicForm{
				Name:    "john",
				Age:     20,
				Aliases: []string{"johnny", "jonny"},
			},
		},
		{
			name:    "invalid query string",
			input:   "%%%",
			target:  &BasicForm{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			decoder := encoding.NewDecoder(strings.NewReader(tt.input))
			err := decoder.Decode(tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if diff := diff(tt.want, tt.target); diff != "" {
					t.Errorf("Decode() mismatch %s", diff)
				}
			}
		})
	}
}

func TestEncoder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   any
		want    []byte
		wantErr bool
	}{
		{
			name: "basic form",
			input: &BasicForm{
				Name:    "john",
				Age:     20,
				Aliases: []string{"johnny", "jonny"},
			},
			want: valuesToBytes(url.Values{
				"age":     {"20"},
				"aliases": {"johnny", "jonny"},
				"name":    {"john"},
			}),
		},
		{
			name:    "invalid target",
			input:   map[int]any{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var b bytes.Buffer
			encoder := encoding.NewEncoder(&b)
			err := encoder.Encode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if diff := diff(tt.want, b.Bytes()); diff != "" {
					t.Errorf("Encode() mismatch %s", diff)
				}
			}
		})
	}
}
