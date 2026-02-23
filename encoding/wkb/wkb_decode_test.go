package wkb_test

import (
	"reflect"
	"testing"

	"github.com/go-spatial/geom/encoding/wkb"
	"github.com/go-spatial/geom/encoding/wkb/internal/tcase"
)

func TestWKBDecode(t *testing.T) {
	fnames, err := tcase.GetFiles("testdata")
	if err != nil {
		t.Fatalf("error getting files: %v", err)
	}
	var fname string

	fn := func(tc tcase.C) func(*testing.T) {
		return func(t *testing.T) {

			if tc.Skip.Is(tcase.TypeDecode) {
				t.Skip("instructed to skip.")
			}

			geom, err := wkb.DecodeBytes(tc.Bytes)
			if !tc.DoesErrorMatch(tcase.TypeDecode, err) {

				eerr := "nil"
				if tc.HasErrorFor(tcase.TypeDecode) {
					eerr = tc.ErrorFor(tcase.TypeDecode)
				}
				t.Errorf("error, expected %v got %v", eerr, err)
				return

			}

			if tc.HasErrorFor(tcase.TypeDecode) {
				return
			}

			if !reflect.DeepEqual(geom, tc.Geometry()) {
				t.Errorf("decode, expected\n\t%v\ngot\n\t%v\n\n", tc.Expected, geom)
			}
		}
	}

	for _, fname = range fnames {
		cases, err := tcase.ParseFile(fname)
		if err != nil {
			t.Fatalf("error parsing file: %v : %v ", fname, err)
		}
		t.Run(fname, func(t *testing.T) {
			if len(cases) == 1 {
				t.Logf("found one test case in %v ", fname)
			} else {
				t.Logf("found %2v test cases in %v ", len(cases), fname)
			}
			for i := range cases {
				t.Run(cases[i].Desc, fn(cases[i]))
			}
		})
	}
}
