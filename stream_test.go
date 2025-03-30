package encoding_test

import (
	"strings"
	"testing"

	"github.com/tomasbasham/encoding/form"
)

func TestDecoder(t *testing.T) {
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
			decoder := form.NewDecoder(strings.NewReader(tt.input))
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
	tests := []struct {
		name    string
		input   any
		want    string
		wantErr bool
	}{
		{
			name: "basic form",
			input: &BasicForm{
				Name:    "john",
				Age:     20,
				Aliases: []string{"johnny", "jonny"},
			},
			want: "name=john&age=20&aliases=johnny&aliases=jonny",
		},
		{
			name:    "invalid target",
			input:   BasicForm{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b strings.Builder
			encoder := form.NewEncoder(&b)
			err := encoder.Encode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got := b.String(); got != tt.want {
					t.Errorf("Encode() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
