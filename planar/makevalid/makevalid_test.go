package makevalid

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
	"github.com/go-spatial/geom/encoding/wkt"
	"github.com/go-spatial/geom/planar/makevalid/hitmap"
)

func asT(fn func(b testing.TB)) func(t *testing.T) {
	return func(t *testing.T) { fn(t) }
}
func asB(fn func(b testing.TB)) func(b *testing.B) {
	return func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			fn(b)
		}
	}
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
	}

	fn := func(tc tcase) func(t testing.TB) {
		return func(t testing.TB) {
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
				t.Errorf("didClip, expected %v got %v", tc.didClip, didClip)
				return
			}
			mp, ok := gmp.(geom.MultiPolygoner)
			if !ok {
				t.Errorf("return MultiPolygon, expected MultiPolygon got %T", gmp)
				return
			}
			if !cmp.MultiPolygonerEqual(tc.ExpectedMultiPolygon, mp) {
				t.Errorf("mulitpolygon, expected %v got %v", tc.ExpectedMultiPolygon, mp)
				if debug {
					t.Logf(strings.TrimSpace(`
Got:
%v
Expected:
%v
ClipBox:
%v
Original Geometry:
%v
`),
						wkt.MustEncode(mp),
						wkt.MustEncode(tc.ExpectedMultiPolygon),
						wkt.MustEncode(tc.ClipBox),
						wkt.MustEncode(tc.MultiPolygon),
					)

				}
				return
			}
		}
	}
	tests := map[string]tcase{}

	// fill tests from makevalidTestCases
	for i, mkvTC := range makevalidTestCases {
		tests[fmt.Sprintf("makevalidTestCases #%v %v", i, mkvTC.Description)] = tcase{
			MultiPolygon:         mkvTC.MultiPolygon,
			ExpectedMultiPolygon: mkvTC.ExpectedMultiPolygon,
			ClipBox:              mkvTC.ClipBox,
			didClip:              true,
		}
	}

	for name, tc := range tests {
		tc := tc
		switch t := tb.(type) {
		case *testing.T:
			t.Run(name, asT(fn(tc)))
		case *testing.B:
			t.Run(name, asB(fn(tc)))
		}
	}
}
