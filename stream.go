package encoding // import "github.com/tomasbasham/encoding/form"

import (
	"fmt"
	"io"
)

type Decoder struct {
	r io.Reader
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

func (d *Decoder) Decode(v any) error {
	body, err := io.ReadAll(d.r)
	if err != nil {
		return fmt.Errorf("form: failed to read body: %w", err)
	}

	return Unmarshal(body, v)
}

type Encoder struct {
	w io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

func (e *Encoder) Encode(v any) error {
	data, err := Marshal(v)
	if err != nil {
		return err
	}

	_, err = e.w.Write(data)
	return err
}
