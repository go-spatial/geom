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

func TestDestructure(t *testing.T) {

	type tcase struct {
		MultiPolygon *geom.MultiPolygon
		ClipBox      *geom.Extent

		Segs []geom.Line
		Err  error
	}
	ctx := context.Background()
	fn := func(tc tcase) func(*testing.T) {
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
			if len(segs) != len(tc.Segs) {
				if debug {
					tcSegs := true
					maxi := len(tc.Segs)
					if len(segs) > maxi {
						maxi = len(segs)
						tcSegs = false
					}

					t.Logf("Expected segs:")
					for i := 0; i < maxi; i++ {
						if tcSegs {
							if i < len(segs) {
								t.Logf("%04d : %v | %v\n", i, wkt.MustEncode(tc.Segs[i]), wkt.MustEncode(segs[i]))
							} else {
								t.Logf("%04d : %v | \n", i, wkt.MustEncode(tc.Segs[i]))
							}

						} else {
							if i < len(tc.Segs) {
								t.Logf("%04d : %v | %v\n", i, wkt.MustEncode(tc.Segs[i]), wkt.MustEncode(segs[i]))
							} else {
								t.Logf("%04d : \t | %v\n", i, wkt.MustEncode(segs[i]))
							}
						}
					}
				}
				t.Errorf("number of segs, expected %v, got %v", len(tc.Segs), len(segs))
				return
			}

		}
	}
	tests := []tcase{
		{
			MultiPolygon: &geom.MultiPolygon{
				{ // Polygon
					{ // Ring
						{1286956.1422558832, 6138803.15957211},
						{1286957.5138675969, 6138809.6399925},
						{1286961.0222077654, 6138815.252628375},
						{1286966.228733862, 6138819.3396373615},
						{1286972.5176202222, 6138821.397139203},
						{1286979.1330808033, 6138821.173193399},
						{1286985.2820067848, 6138818.695793352},
						{1286990.1992814348, 6138814.272866236},
						{1286993.3157325392, 6138808.436285537},
						{1286994.2394710402, 6138801.885883152},
						{1286992.8678593265, 6138795.40546864},
						{1286989.3781805448, 6138789.792845784},
						{1286984.1623237533, 6138785.719847463},
						{1286977.864106701, 6138783.662354196},
						{1286971.2486461198, 6138783.872302467},
						{1286965.1183815224, 6138786.349692439},
						{1286960.1824454917, 6138790.7726051165},
						{1286957.084655768, 6138796.623170342},
					},
				},
			},
			ClipBox: geom.NewExtent(
				[2]float64{1286940.46060967, 6138830.2432236},
				[2]float64{1286969.19030943, 6138807.58852643},
			),
		},
	}

	for i, tc := range tests {
		sort.Sort(planar.LinesByXY(tc.Segs))
		t.Run(strconv.Itoa(i), fn(tc))
	}

}
