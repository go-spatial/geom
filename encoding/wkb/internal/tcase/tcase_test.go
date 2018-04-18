package tcase

import (
	"reflect"
	"testing"

	"github.com/go-spatial/geom"
)

type tcase struct {
	filename string
	cases    []C
}

func TestParse(t *testing.T) {

	fn := func(t *testing.T, tc tcase) {
		cases, err := ParseFile(tc.filename)
		if err != nil {
			t.Errorf("error, expected nil got %v", err)
		}
		if len(cases) != len(tc.cases) {
			t.Errorf("number of cases, expected %v got %v", len(tc.cases), len(cases))
			return
		}
		for i, tcase := range tc.cases {
			acase := cases[i]
			if acase.Desc != tcase.Desc {
				t.Errorf("Desc, expected %v got %v", tcase.Desc, acase.Desc)
			}
			if !reflect.DeepEqual(tcase.Expected, acase.Expected) {
				t.Errorf("Expected, expected %#v got %#v", tcase.Expected, acase.Expected)
			}
			if !reflect.DeepEqual(tcase.Bytes, acase.Bytes) {
				t.Errorf("Bytes, expected %v got %v", tcase.Bytes, acase.Bytes)
			}
		}
	}
	tests := map[string]tcase{
		"1": {
			filename: "testdata/point.tcase",
			cases: []C{
				{
					Desc:     "This is a simple test",
					Expected: geom.Point{2, 4},
					Bytes:    []byte{0x00, 0x00, 0x00, 0x00, 0xaf, 0x00, 0xaf, 0x0c, 0xd0, 0x0d, 0xac, 0x00, 0xDE, 0xAF, 0xD0, 0x0d, 0xac, 0xff, 0x00, 0x00},
				},
				{
					Desc:     "This is a simple test",
					Expected: geom.Point{2, 4},
					Bytes:    []byte{0x00, 0x00, 0x00, 0x00, 0xaf, 0x00, 0xaf, 0x0c, 0xd0, 0x0d, 0xac, 0x00, 0xDE, 0xAF, 0xD0, 0x0d, 0xac, 0xff, 0x00, 0x00},
				},
			},
		},
	}
	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) { fn(t, tc) })
	}
}
