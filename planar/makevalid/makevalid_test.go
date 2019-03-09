package makevalid

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
	"github.com/go-spatial/geom/internal/debugger"
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

func init() {
	debugger.DefaultOutputDir = "_test_output"
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

			hm, err := hitmap.NewFromPolygons(ctx, tc.ClipBox, tc.MultiPolygon.Polygons()...)
			if err != nil {
				panic("Was not expecting the hitmap to return error.")
			}

			mv := &Makevalid{Hitmap: hm}

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
				debugger.Record(ctx, tc.MultiPolygon, debugger.CategoryInput, "Original Polygon")
				debugger.Record(ctx, tc.ClipBox, debugger.CategoryInput, "Original Clipbox")
				debugger.Record(ctx, mp, debugger.CategoryGot, "Result Polygon")
				debugger.Record(ctx, tc.ExpectedMultiPolygon, debugger.CategoryExpected, "Expected Polygon")
			}

			if !cmp.MultiPolygonerEqual(tc.ExpectedMultiPolygon, mp) {
				t.Errorf("mulitpolygon, expected %v got %v", tc.ExpectedMultiPolygon, mp)
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
			ctx = debugger.AugmentContext(ctx, "makevalid.TestMakeValid")
			defer debugger.Close(ctx)
		}
	}

	for name, tc := range tests {
		tc := tc
		switch t := tb.(type) {
		case *testing.T:
			if debug {
				ctx = debugger.SetTestName(ctx, name)
			}
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

	fn := func(ctx context.Context, tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			if debug {
				ctx = debugger.SetTestName(ctx, t.Name())
			}

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
				debugger.Record(ctx, tc.ClipBox, debugger.CategoryInput, "Clipbox")
				debugger.Record(ctx, tc.MultiPolygon, debugger.CategoryInput, "MultiPolygon")
				for i := range tc.Segs {
					debugger.Record(ctx, tc.Segs[i], debugger.CategoryExpected, "Segments #%v", i)
				}
				for i := range segs {
					debugger.Record(ctx, segs[i], debugger.CategoryGot, "Segments #%v", i)
				}
			}

			if !cmp.GeomLineEqual(tc.Segs, segs) {
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

	ctx := context.Background()

	if debug {
		ctx = debugger.AugmentContext(ctx, "")
		defer debugger.Close(ctx)
	}

	for i, tc := range tests {
		sort.Sort(planar.LinesByXY(tc.Segs))
		name := strconv.Itoa(i)
		t.Run(name, fn(ctx, tc))
	}

}

func TestSplitIntersectingLines(t *testing.T) {
	type tcase struct {
		description string
		lines       []geom.Line
		clipbox     *geom.Extent

		newLines []geom.Line
	}
	fn := func(ctx context.Context, tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			if debug {
				ctx = debugger.SetTestName(ctx, t.Name())
			}

			segs := splitIntersectingLines(ctx, tc.clipbox, tc.lines)

			sort.Sort(planar.LinesByXY(segs))
			if debug {
				debugger.Record(ctx, tc.clipbox, debugger.CategoryInput, "clipbox")
				for i := range tc.lines {
					debugger.Record(ctx, tc.lines[i], debugger.CategoryInput, "original line #%v", i)
				}
				for i := range tc.newLines {
					debugger.Record(ctx, tc.newLines[i], debugger.CategoryExpected, "line #%v", i)
				}
				for i := range segs {
					debugger.Record(ctx, segs[i], debugger.CategoryGot, "Segments #%v", i)
				}
			}

			if !cmp.GeomLineEqual(tc.newLines, segs) {
				t.Errorf("lines, expected %v got %v", tc.newLines, segs)
			}
		}
	}
	tests := [...]tcase{
		{
			description: "Github Issue 32",
			lines: []geom.Line{
				{{1286931.429, 6138810.018}, {1286970.842, 6138810.018}},
				{{1286970.842, 6138810.018}, {1286970.842, 6138849.43}},
				{{1286970.842, 6138849.43}, {1286931.429, 6138849.43}},
				{{1286931.429, 6138849.43}, {1286931.429, 6138810.018}},
				{{1286957.514, 6138809.64}, {1286961.022, 6138815.253}},
				{{1286961.022, 6138815.253}, {1286966.229, 6138819.34}},
				{{1286966.229, 6138819.34}, {1286972.518, 6138821.397}},
			},
			clipbox: geom.NewExtent(
				[2]float64{1.2869314293799447e+06, 6.138810017620263e+06},
				[2]float64{1.286970842222649e+06, 6.138849430462967e+06},
			),
			newLines: []geom.Line{

				{{1286957.75, 6138810.018}, {1286970.842, 6138810.018}},
				{{1286970.842, 6138810.018}, {1286970.842, 6138820.849}},
				{{1286970.842, 6138820.849}, {1286970.842, 6138849.43}},
				{{1286957.75, 6138810.018}, {1286961.022, 6138815.252}},
				{{1286961.022, 6138815.252}, {1286966.229, 6138819.34}},
				{{1286966.229, 6138819.34}, {1286970.842, 6138820.849}},

				{{1286970.842, 6138849.43}, {1286931.429, 6138849.43}},
				{{1286931.429, 6138849.43}, {1286931.429, 6138810.018}},
				{{1286931.429, 6138810.018}, {1286957.75, 6138810.018}},
			},
		},
	}

	ctx := context.Background()

	if debug {
		ctx = debugger.AugmentContext(ctx, "")
		defer debugger.Close(ctx)
	}
	for i, tc := range tests {
		sort.Sort(planar.LinesByXY(tc.newLines))
		name := tc.description
		if name == "" {
			name = strconv.Itoa(i)
		}
		t.Run(name, fn(ctx, tc))
	}
}
