package record

import (
	"fmt"
	"testing"
	"time"
)

type Record struct {
	Seq   int64
	Name  string `record:"6"`
	Code  int64  `record:"6"`
	Label string `record:"6,upper"`

	Internal string `record:"-"`

	Free string
}

type TsRecord struct {
	Record
	Date time.Time `record:"6"`
}

func TestMarshal(t *testing.T) {
	encodeTests := []struct {
		src    interface{}
		result string
	}{
		{
			Record{
				Seq:      1,
				Name:     "HELLO",
				Code:     0,
				Label:    "Hello World",
				Internal: "Unused",
				Free:     "world",
			},
			"1 HELLO000000HELLO world",
		},
		{
			Record{
				Seq:      1,
				Name:     "ERROR",
				Code:     12345,
				Label:    "Stack Overflow",
				Internal: "Unused",
				Free:     "Overflow",
			},
			"1 ERROR012345STACK Overflow",
		},
		{
			TsRecord{
				Record: Record{},
				Date:   time.Date(2014, time.November, 21, 20, 26, 0, 0, time.UTC),
			},
			"0      000000      20141121",
		},
	}
	for _, encTest := range encodeTests {
		b, err := Marshal(encTest.src)
		if err != nil {
			t.Errorf("Unexpected error for Marshal: %v", err)
		}
		if encTest.result != string(b) {
			t.Errorf("Unexpected value for Marshal: `%s`, expected `%s`", string(b), encTest.result)
		}
	}
}

func ExampleMarshal() {
	pbm := struct {
		Version string `record:"2,upper"`
		Width   string `record:"3"`
		Height  int    `record:"1"`
		Filler  string `record:"1"`

		// These fields are skipped
		Extension string `record:"-"`

		// Untagged field are encoded verbatim
		Pixels string
	}{
		Version:   "p1",
		Width:     "2 ",
		Height:    2,
		Extension: "pbm",
		Pixels:    "1 0 1 0",
	}
	b, _ := Marshal(&pbm)
	fmt.Printf("%s", string(b))
	// Output: P1 2 2 1 0 1 0
}
