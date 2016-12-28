package record

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"
)

var (
	// DateFormat is used to format strings from
	DateFormat = "20060102"

	// dateType is used to signal that we are travessing a time.Time
	dateType = reflect.TypeOf(time.Time{})
)

// Marshal takes a struct value or pointer
// and returns the encoded bytes and a nil error,
// or a nil byte slice and the encoding error.
func Marshal(src interface{}) ([]byte, error) {
	var b bytes.Buffer
	if err := NewEncoder(&b).Encode(src); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// Encoder is responsible for encoding record values
// from struct fields.
type Encoder struct {
	w io.Writer
}

// NewEncoder returns an initialized encoder
// that writes the encoded bytes into the specified io.Writer w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w}
}

// Encode takes an struct value or pointer
// and encode its fields into the encoder writer.
// Returns any encoding errors, if any.
// When returning an error, src may have been partially written.
func (e *Encoder) Encode(src interface{}) error {
	v := reflect.ValueOf(src)
	t := reflect.TypeOf(src)

	switch t.Kind() {
	case reflect.Ptr:
		if v.IsNil() || v.Elem().Kind() != reflect.Struct {
			return fmt.Errorf("record: invalid pointer")
		}
		return e.encodeStruct(v.Elem(), v.Elem().Type())
	case reflect.Struct:
		return e.encodeStruct(v, t)
	}
	return fmt.Errorf("record: invalid value type: %s", t)
}

// encodeStruct encodes the content of struct s into w.
func (e *Encoder) encodeStruct(s reflect.Value, sType reflect.Type) error {
	for i := 0; i < sType.NumField(); i++ {
		f := sType.Field(i)
		fval := s.Field(i)
		tag := parseTags(f)
		if tag.skip {
			continue
		}
		switch f.Type.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			fmtSpec := "%d"
			if !tag.noPadding {
				fmtSpec = fmt.Sprintf("%%0%dd", tag.size)
			}
			fmt.Fprintf(e.w, fmtSpec, fval.Int())
		case reflect.String:
			fmtSpec := "%s"
			if !tag.noPadding {
				fmtSpec = fmt.Sprintf("%% %ds", tag.size)
			}
			str := fval.String()
			if tag.upper {
				str = strings.ToUpper(str)
			}
			if len(str) > tag.size && tag.size > 0 {
				str = string([]rune(str)[:tag.size])
			}
			fmt.Fprintf(e.w, fmtSpec, str)
		case reflect.Struct:
			// Special case some stdlib structs
			if f.Type.ConvertibleTo(dateType) {
				fmt.Fprintf(e.w, "%s", fval.Interface().(time.Time).Format(DateFormat))
			} else if err := e.encodeStruct(fval, f.Type); err != nil {
				return err
			}
		}
	}
	return nil
}
