package makevalid

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
	"github.com/go-spatial/geom/planar/makevalid/hitmap"
)

func TestMakeValid(t *testing.T)      { checkMakeValid(t) }
func BenchmakrMakeValid(b *testing.B) { checkMakeValid(b) }

func checkMakeValid(tb testing.TB) {
	type tcase struct {
		MultiPolygon         *geom.MultiPolygon
		ExpectedMultiPolygon *geom.MultiPolygon
		ClipBox              *geom.Extent
		err                  error
		didClip              bool
	}

	fn := func(t testing.TB, tc tcase) {
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
	tests := map[string]tcase{}
	// fill tests from makevalidTestCases
	for i, mkvTC := range makevalidTestCases {
		tests[fmt.Sprintf("makevalidTestCases #%v %v", i, mkvTC.Description)] = tcase{
			MultiPolygon:         mkvTC.MultiPolygon,
			ExpectedMultiPolygon: mkvTC.ExpectedMultiPolygon,
			didClip:              true,
		}
	}
	for name, tc := range tests {
		tc := tc
		switch t := tb.(type) {
		case *testing.T:
			t.Run(name, func(t *testing.T) { fn(t, tc) })
		case *testing.B:
			t.Run(name, func(b *testing.B) {
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					fn(b, tc)
				}
			})
		}
	}
}
