package encoding_test

import (
	"net/url"
	"testing"
	"time"

	"github.com/tomasbasham/encoding/form"
)

func TestMarshal(t *testing.T) {
	baseTime := time.Date(2025, 2, 8, 0, 0, 0, 0, time.UTC)
	optionalVal := "optional_value"

	tests := []struct {
		name    string
		input   any
		want    url.Values
		wantErr bool
	}{
		{
			name: "basic form",
			input: BasicForm{
				Name:    "john",
				Aliases: []string{"johnny", "jonny"},
				Age:     20,
			},
			want: url.Values{
				"age":     {"20"},
				"aliases": {"johnny", "jonny"},
				"name":    {"john"},
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
			want: url.Values{
				"age":        {"25"},
				"aliases":    {"janet", "jan"},
				"created_at": {"2025.02.08"},
				"id":         {"1"},
				"name":       {"jane"},
				"optional":   {"optional_value"},
			},
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
			want: url.Values{
				"Empty":   {"value"},
				"NoTag":   {"value"},
				"complex": {"2025.02.08"},
				"public":  {"visible"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := form.Marshal(tt.input)
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
