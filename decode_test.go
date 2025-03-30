package encoding_test

import (
	"net/url"
	"testing"
	"time"

	"github.com/tomasbasham/encoding/form"
)

func TestUnmarshal(t *testing.T) {
	optionalVal := "optional_value"
	baseTime := time.Date(2025, 2, 8, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		input   url.Values
		target  any
		want    any
		wantErr bool
	}{
		{
			name: "basic form",
			input: url.Values{
				"age":     {"20"},
				"aliases": {"johnny", "jonny"},
				"name":    {"john"},
			},
			target: &BasicForm{},
			want: &BasicForm{
				Name:    "john",
				Aliases: []string{"johnny", "jonny"},
				Age:     20,
			},
		},
		{
			name: "complex form with custom type",
			input: url.Values{
				"age":        {"25"},
				"aliases":    {"janet", "jan"},
				"created_at": {"2025.02.08"},
				"id":         {"1"},
				"name":       {"jane"},
				"optional":   {"optional_value"},
			},
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
			input:   url.Values{},
			target:  BasicForm{},
			wantErr: true,
		},
		{
			name:    "nil target",
			input:   url.Values{},
			target:  nil,
			wantErr: true,
		},
		{
			name: "ignored fields",
			input: url.Values{
				"Empty":   {"value"},
				"NoTag":   {"value"},
				"Omitted": {"present"},
				"complex": {"2025.02.08"},
				"ignored": {"skip"},
				"private": {"hidden"},
				"public":  {"visible"},
			},
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
			input: url.Values{
				"age": {"not_a_number"},
			},
			target:  &BasicForm{},
			wantErr: true,
		},
		{
			name: "invalid custom type format",
			input: url.Values{
				"created_at": {"invalid_date"},
			},
			target:  &ComplexForm{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := form.Unmarshal(tt.input, tt.target)
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
