package makevalid

import (
	"context"
	"flag"
	"fmt"
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
	"github.com/go-spatial/geom/planar/makevalid/hitmap"
)

var runAll bool

func init() {
	flag.BoolVar(&runAll, "run-all", false, "to run tests marked to be skipped")
}

func TestMakeValid(t *testing.T)      { checkMakeValid(t) }
func BenchmakrMakeValid(b *testing.B) { checkMakeValid(b) }

func checkMakeValid(tb testing.TB) {
	type tcase struct {
		MultiPolygon         *geom.MultiPolygon
		ExpectedMultiPolygon *geom.MultiPolygon
		ClipBox              *geom.Extent
		err                  error
		didClip              bool
		skip                 string
	}

	fn := func(tc tcase) func(testing.TB) {
		return func(t testing.TB) {
			if tc.skip != "" && !runAll {
				t.Skipf(tc.skip)
				return
			}
			hm, err := hitmap.NewFromPolygons(tc.ClipBox, tc.MultiPolygon.Polygons()...)
			if err != nil {
				panic("Was not expecting the hitmap to return error.")
			}
			mv := &Makevalid{
				Hitmap: hm,
			}
			gmp, didClip, gerr := mv.Makevalid(context.Background(), tc.MultiPolygon, tc.ClipBox)
			if tc.err != nil {
				if tc.err != gerr {
					t.Errorf("error, expected %v got %v", tc.err, gerr)
					return
				}
			}
			if gerr != nil {
				t.Errorf("error, expected %v got %v", tc.err, gerr)
				return
			}
			if didClip != tc.didClip {
				t.Errorf("didClipt, expected %v got %v", tc.didClip, didClip)
			}
			mp, ok := gmp.(geom.MultiPolygoner)
			if !ok {
				t.Errorf("return MultiPolygon, expected MultiPolygon got %T", gmp)
				return
			}
			if !cmp.MultiPolygonerEqual(tc.ExpectedMultiPolygon, mp) {
				t.Errorf("mulitpolygon, expected %v got %v", tc.ExpectedMultiPolygon, mp)
			}
		}
	}
	tests := map[string]tcase{}
	// fill tests from makevalidTestCases
	for i, mkvTC := range makevalidTestCases {
		name := fmt.Sprintf("makevalidTestCases #%v %v", i, mkvTC.Description)
		tc := tcase{
			MultiPolygon:         mkvTC.MultiPolygon,
			ExpectedMultiPolygon: mkvTC.ExpectedMultiPolygon,
			didClip:              true,
		}
		switch name {
		case "makevalidTestCases #1 Four Square IO_OI":
			tc.skip = `failed: mulitpolygon, expected &[[[[1 4] [5 4] [5 8] [1 8]]] [[[5 0] [9 0] [9 4] [5 4]]]] got &[]`
		case "makevalidTestCases #2 Four columns invalid multipolygon":
			tc.skip = "failed: mulitpolygon, expected &[[[[0 3] [3 3] [3 0] [6 0] [6 8] [3 7] [0 7]] [[1 5] [3 7] [5 5] [3 4]]]] got &[]"
		}
		tests[name] = tc
	}

	// skip the following tests:
	for name, tc := range tests {
		tfn := fn(tc)
		switch t := tb.(type) {
		case *testing.T:
			t.Run(name, func(t *testing.T) { tfn(t) })
		case *testing.B:
			t.Run(name, func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					tfn(b)
				}
			})
		}
	}
}
