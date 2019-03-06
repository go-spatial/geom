package makevalid

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
	"github.com/go-spatial/geom/encoding/wkt"
	"github.com/go-spatial/geom/planar"
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

	fn := func(ctx context.Context, tc tcase) func(t testing.TB) {
		return func(t testing.TB) {

			hm, err := hitmap.NewFromPolygons(tc.ClipBox, tc.MultiPolygon.Polygons()...)
			if err != nil {
				panic("Was not expecting the hitmap to return error.")
			}

			mv := &Makevalid{
				Hitmap: hm,
			}


			gmp, didClip, gerr := mv.Makevalid(ctx, tc.MultiPolygon, tc.ClipBox)
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
			if debug {
				debugRecordEntity(ctx, "Original Polygon", "input", tc.MultiPolygon)
				debugRecordEntity(ctx, "Result Polygon", "got", mp)
				debugRecordEntity(ctx, "Expected Polygon", "expected", tc.ExpectedMultiPolygon)
				debugRecordEntity(ctx, "Original Clipbox", "input", tc.ClipBox)
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

	ctx := context.Background()

	if debug {
		if _, ok := tb.(*testing.T); ok {
			ctx = debugContext("", ctx)
			defer debugClose(ctx)
		}
	}

	for name, tc := range tests {
		tc := tc
		switch t := tb.(type) {
		case *testing.T:
			name, ctx = debugAddTestName(ctx, name)
			t.Run(name, asT(fn(ctx, tc)))
		case *testing.B:
			t.Run(name, asB(fn(ctx, tc)))
		}
	}
}

func TestDestructure(t *testing.T) {

	type tcase struct {
		MultiPolygon *geom.MultiPolygon
		ClipBox      *geom.Extent

		Segs []geom.Line
		Err  error
	}

	ctx := debugContext("", context.Background())
	defer debugClose(ctx)

	fn := func(ctx context.Context, tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			segs, err := Destructure(ctx, tc.ClipBox, tc.MultiPolygon)
			if tc.Err == nil && err != nil {
				t.Errorf("error, expected nil, got %v", err)
				return
			}
			if tc.Err != nil {
				if err == nil || tc.Err.Error() != err.Error() {
					t.Errorf("error, expected %v, got %v", tc.Err, err)
					return
				}
				// We are good, if expected an error, the Segs is not valid
				return
			}
			sort.Sort(planar.LinesByXY(segs))
			if debug {
				debugRecordEntity(ctx, "Clipbox", "input", tc.ClipBox)
				debugRecordEntity(ctx, "MultiPolygon", "input", tc.MultiPolygon)
				for i := range tc.Segs {
					debugRecordEntity(ctx, fmt.Sprintf("Segments #%v", i), "expected", tc.Segs[i])
				}
				for i := range segs {
					debugRecordEntity(ctx, fmt.Sprintf("Segments #%v", i), "got", segs[i])
				}
			}

			if !cmp.GeomLineEqual(tc.Segs, segs) {
				if debug {
					t.Logf("Expected segs:")
					t.Logf(dumpWKTLineSegments("", tc.Segs, segs))
				}
				if len(tc.Segs) != len(segs) {
					t.Errorf("number of segs, expected %v, got %v", len(tc.Segs), len(segs))
				} else {
					t.Errorf("segs, expected %v, got %v", tc.Segs, segs)
				}
				return
			}

		}
	}
	tests := []tcase{
		{
			MultiPolygon: makevalidTestCases[4].MultiPolygon,
			ClipBox:      makevalidTestCases[4].ClipBox,
			Segs: []geom.Line{
				geom.Line{{1.2869404610000001e+06, 6.138807589e+06}, {1.2869404610000001e+06, 6.138830243e+06}},
				geom.Line{{1.2869404610000001e+06, 6.138807589e+06}, {1.28695708e+06, 6.138807589e+06}},
				geom.Line{{1.2869404610000001e+06, 6.138830243e+06}, {1.28696919e+06, 6.138830243e+06}},
				geom.Line{{1.28695708e+06, 6.138807589e+06}, {1.286957514e+06, 6.13880964e+06}},
				geom.Line{{1.28695708e+06, 6.138807589e+06}, {1.28696919e+06, 6.138807589e+06}},
				geom.Line{{1.286957514e+06, 6.13880964e+06}, {1.286961022e+06, 6.138815252e+06}},
				geom.Line{{1.286961022e+06, 6.138815252e+06}, {1.286966229e+06, 6.13881934e+06}},
				geom.Line{{1.286966229e+06, 6.13881934e+06}, {1.28696919e+06, 6.138820309e+06}},
				geom.Line{{1.28696919e+06, 6.138807589e+06}, {1.28696919e+06, 6.138820309e+06}},
				geom.Line{{1.28696919e+06, 6.138820309e+06}, {1.28696919e+06, 6.138830243e+06}},
			},
		},
		{
			MultiPolygon: makevalidTestCases[0].MultiPolygon,
			ClipBox:      makevalidTestCases[0].ClipBox,
			Segs: []geom.Line{
				geom.Line{{10, 20}, {1, 1}},
				geom.Line{{1, 1}, {15, 10}},
				geom.Line{{15, 10}, {10, 20}},
			},
		},
	}

	for i, tc := range tests {
		sort.Sort(planar.LinesByXY(tc.Segs))
		name, ctx := debugAddTestName(ctx, strconv.Itoa(i))
		t.Run(name, fn(ctx, tc))
	}

}

func TestSplitIntersectingLines(t *testing.T) {
	type tcase struct {
		Lines   []geom.Line
		ClipBox *geom.Extent

		NewLines []geom.Line
	}
	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {

		}
	}
	tests := [...]tcase{}
	for i, tc := range tests {
		sort.Sort(planar.LinesByXY(tc.NewLines))
		t.Run(strconv.Itoa(i), fn(tc))
	}
}
