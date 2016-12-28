package record

import (
	"reflect"
	"strconv"
	"strings"
)

// tag holds configuration options for encoding fields.
type tag struct {
	size      int
	noPadding bool
	upper     bool
	skip      bool
	optional  bool
}

// parseTags takes a field and returns the parsed tag.
// The default tag values are returned, when no tag is specified.
func parseTags(f reflect.StructField) *tag {
	t := &tag{}
	tagVal := f.Tag.Get("csv")
	if tagVal != "" {
		elem := strings.Split(f.Tag.Get("csv"), ",")
		if size, err := strconv.Atoi(elem[0]); err == nil {
			t.size = size
			elem = elem[1:]
		}
		for _, e := range elem {
			switch e {
			case "nopad", "nopadding":
				t.noPadding = true
			case "upper":
				t.upper = true
			case "optional":
				t.optional = true
			case "-":
				t.skip = true
			}
		}
	}
	return t
}
