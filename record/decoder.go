package record

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// List of error types that can be returned during Decode.
const (
	ErrEOF int = iota
	ErrInvalidInt
	ErrInvalidDate
)

// DecodingError is aggregated during decoding and returned
// within ErrorList.
type DecodingError struct {
	Type  int    // Error type
	Field string // Field name (in struct)
	Token string // The token that failed to decode
	Err   error  // The original error when decoding
}

// Error return the underlying error message
func (d DecodingError) Error() string {
	return fmt.Sprintf("decode error: %s, field=%s, token=%s", d.Err.Error(), d.Field, d.Token)
}

// ErrorList wraps a series of DecodingError in order to allow
// callers to inspect what went wrong when decoding a line.
type ErrorList struct {
	Errors []DecodingError
}

// Error return the list of errors, as printed by fmt.Sprintf("%v").
func (e ErrorList) Error() string {
	return fmt.Sprintf("record: several errors: %v", e.Errors)
}

// Add append a new DecodingError to the error list.
func (e *ErrorList) Add(errType int, field, token string, err error) {
	e.Errors = append(e.Errors, DecodingError{
		Type:  errType,
		Field: field,
		Token: token,
		Err:   err,
	})
}

// Decoder controls decoding of an io.Reader, one line at a time.
type Decoder struct {
	r  *bufio.Reader
	dt string
}

// NewDecoder initializes a new Decoder to parse the provided Reader.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r:  bufio.NewReader(r),
		dt: DateFormat,
	}
}

// TimeLayout overrides the date/time layout, used in the next
// call to Decode.
// Returns the same reference, for method chaining
func (d *Decoder) TimeLayout(layout string) *Decoder {
	d.dt = layout
	return d
}

// Decode decodes the next line in the buffer into the specified value.
// Value must be a struct or a struct pointer.
// Aways return a non-nil ErrorList, or a nil error, allowing the callee
// to inspect all errors that happened.
func (d *Decoder) Decode(value interface{}) error {
	v := reflect.ValueOf(value)
	t := reflect.TypeOf(value)
	switch t.Kind() {
	case reflect.Ptr:
		if v.IsNil() || v.Elem().Kind() != reflect.Struct {
			return fmt.Errorf("record: invalid pointer") // TODO: move to type const
		}
		return d.decodeStruct(v.Elem(), v.Elem().Type())
	}
	return fmt.Errorf("record: invalid value type: %s", t)
}

func (d *Decoder) decodeStruct(v reflect.Value, t reflect.Type) error {
	var (
		l     string        // line
		token string        // next token
		start int           // start
		tag   *tag          // struct tag with metadata
		fval  reflect.Value // field to set

		// Error handling
		err       error
		errorList = ErrorList{
			Errors: make([]DecodingError, 0),
		}
	)

	l, err = d.r.ReadString('\n')
	if err != nil && err != io.EOF {
		errorList.Add(ErrEOF, "", "", err)
		return errorList
	}

	for i := 0; i < t.NumField() && start < len(l); i++ {
		f := t.Field(i)
		if tag = parseTags(f); tag.skip {
			continue
		}

		if token, start, err = nextToken(l, start, tag); err != nil {
			return err
		}

		if fval = v.Field(i); !fval.CanSet() {
			log.Printf("can't set %v", f)
			continue
		}

		switch f.Type.Kind() {
		case reflect.String:
			fval.SetString(strings.TrimSpace(token))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			token = strings.TrimSpace(token)
			if intVal, err := strconv.ParseInt(token, 10, 64); err == nil {
				fval.SetInt(intVal)
			} else {
				// Optional tag avoids unwanted invalid syntax for whitespace values.
				if tag.optional && token == "" {
					continue
				}
				errorList.Add(ErrInvalidInt, f.Name, token, err)
			}
		case reflect.Struct:
			if f.Type.ConvertibleTo(dateType) {
				// We need to parse, reformat into the MarshallText, then unmarshal in t again
				if timeVal, err := time.Parse(d.dt, token); err == nil {
					fval.Set(reflect.ValueOf(timeVal))
				} else {
					if tag.optional && token == "" {
						continue
					}
					errorList.Add(ErrInvalidDate, f.Name, token, err)
				}
			} else {
				// Attempt to deep-decode nested structs
				if err := d.decodeStruct(fval, f.Type); err != nil {
					a := err.(ErrorList)
					for _, err := range a.Errors {
						errorList.Add(err.Type, err.Field, token, err.Err)
					}
					continue
				}
			}
		default:
			errorList.Add(ErrEOF, "", "", fmt.Errorf("record: unsupported type: %v", f))
			return errorList
		}
	}

	if len(errorList.Errors) > 0 {
		return errorList
	}

	return nil
}

func nextToken(l string, start int, t *tag) (string, int, error) {
	if t == nil {
		return "", start, fmt.Errorf("record: unexpected nil tag")
	}
	end := start + t.size
	if t.size == 0 {
		end = len(l) - 1
	}
	if end > len(l) {
		return "", end, fmt.Errorf("record: end of line")
	}
	return string([]rune(l)[start:end]), end, nil
}

// Unmarshal decodes the provided data into the target value.
// v must be a valid type for Decoder.Decode().
func Unmarshal(data []byte, v interface{}) error {
	return NewDecoder(bytes.NewBuffer(data)).Decode(v)
}
