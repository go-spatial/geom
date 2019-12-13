package makevalid

import (
	"context"
	"flag"
	"fmt"
	"strings"
	"testing"

	"github.com/go-spatial/geom/encoding/wkt"
	"github.com/go-spatial/geom/internal/test/must"
	"github.com/go-spatial/geom/slippy"
	"github.com/go-spatial/geom/winding"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/planar/makevalid/hitmap"
)

var runAll bool

func init() {
	flag.BoolVar(&runAll, "run-all", false, "to run tests marked to be skipped")
}

func TestMakeValid(t *testing.T)      { checkMakeValid(t) }
func BenchmakrMakeValid(b *testing.B) { checkMakeValid(b) }

func decodeMultiPolygon(content string) *geom.MultiPolygon {
	str := strings.NewReader(content)
	g, err := wkt.NewDecoder(str).Decode()
	if err != nil {
		panic(err)
	}
	mp, ok := g.(geom.MultiPolygon)
	if !ok {
		panic("Expected multipolygon")
	}
	return &mp
}

func checkMakeValid(tb testing.TB) {
	type tcase struct {
		MultiPolygon         *geom.MultiPolygon
		ExpectedMultiPolygon *geom.MultiPolygon
		ClipBox              *geom.Extent
		err                  error
		didClip              bool
		skip                 string
	}

	order := winding.Order{}

	fn := func(tc tcase) func(testing.TB) {
		return func(t testing.TB) {
			if tc.skip != "" && !runAll {
				t.Skipf(tc.skip)
				return
			}
			if tc.MultiPolygon == nil {
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
				//t.Logf("input: \n%v", wkt.MustEncode(tc.MultiPolygon))
				t.Errorf("multipolygon, expected \n%v\n got \n%v", wkt.MustEncode(tc.ExpectedMultiPolygon), wkt.MustEncode(mp))
				if !geom.IsEmpty(tc.ExpectedMultiPolygon) {
					for p, ply := range tc.ExpectedMultiPolygon.Polygons() {
						for l, ln := range ply {
							t.Logf("expected windorder %v:%v: %v", p, l, order.OfPoints(ln...))
						}
					}
				}
				if !geom.IsEmpty(mp) {
					for p, ply := range mp.Polygons() {
						for l, ln := range ply {
							t.Logf("got      windorder %v:%v: %v", p, l, order.OfPoints(ln...))
						}
					}
				}
			}
		}
	}
	tests := map[string]tcase{
		"issue#70": {
			MultiPolygon:         must.MPPointer(must.ReadMultiPolygon("testdata/issue/70/multipolygon_input.wkt")),
			ExpectedMultiPolygon: must.MPPointer(must.ReadMultiPolygon("testdata/issue/70/multipolygon_expected.wkt")),
			didClip:              true,
		},
		"issue#70_full_no_clip": {
			MultiPolygon:         must.MPPointer(must.ReadMultiPolygon("testdata/issue/70/multipolygon_input.wkt")),
			ExpectedMultiPolygon: must.MPPointer(must.ReadMultiPolygon("testdata/issue/70/multipolygon-full-no-clip_expected.wkt")),
			didClip:              true,
		},
		"issue#70_full": {
			ClipBox: webMercatorTileExtent(13, 8054, 2677).ExpandBy(64.0),
			MultiPolygon:         must.MPPointer(must.ReadMultiPolygon("testdata/issue/70/multipolygon_full_input.wkt")),
			ExpectedMultiPolygon: must.MPPointer(must.ReadMultiPolygon("testdata/issue/70/multipolygon_full_expected.wkt")),
			didClip:              true,
		},
	}
	// fill tests from makevalidTestCases
	tb.Logf("length: %v", len(makevalidTestCases))
	for i, mkvTC := range makevalidTestCases {
		name := fmt.Sprintf("makevalidTestCases #%v %v", i, mkvTC.Description)
		tc := tcase{
			MultiPolygon:         mkvTC.MultiPolygon,
			ExpectedMultiPolygon: mkvTC.ExpectedMultiPolygon,
			didClip:              true,
		}
		tests[name] = tc
	}

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

func webMercatorTileExtent(z, x, y uint) *geom.Extent {
	grid, err := slippy.NewGrid(3857)
	if err != nil {
		panic(err)
	}

	ext, ok := slippy.Extent(grid, slippy.NewTile(z, x, y))
	if !ok {
		panic("no tile")
	}

	return ext
}
