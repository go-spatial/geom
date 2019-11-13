package testing

import (
	"math"
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
	"github.com/go-spatial/geom/winding"
)

func TestBoxPolygon(t *testing.T) {
	type tcase struct {
		dim float64
		res geom.Polygon
	}
	order := winding.Order{}

	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			got := BoxPolygon(tc.dim)

			wo := order.OfPoints(got[0]...)
			if !wo.IsClockwise() {
				t.Error("winding order of box not clockwise")
			}

			if !cmp.GeometryEqual(got, tc.res) {
				t.Error("geometries not equal")
			}
		}
	}

	tcases := map[string]tcase{
		"dim 1": {
			dim: 1.0,
			res: geom.Polygon{{{0, 0}, {1.0, 0}, {1.0, 1.0}, {0, 1.0}}},
		},
		"dim -1": {
			dim: -1.0,
			res: geom.Polygon{{{0, 0}, {-1.0, 0}, {-1.0, -1.0}, {0, -1.0}}},
		},
	}

	for k, v := range tcases {
		t.Run(k, fn(v))
	}
}

func TestSinLineString(t *testing.T) {
	type tcase struct {
		amp, start, end float64
		points          int
		out             geom.LineString
	}

	fn := func(tc tcase) func(t *testing.T) {
		return func(t *testing.T) {
			out := SinLineString(tc.amp, tc.start, tc.end, tc.points)

			if !cmp.GeometryEqual(out, tc.out) {
				t.Errorf("expected %v got %v", tc.out, out)
			}
		}
	}

	tcases := map[string]tcase{
		"amp 10": {
			amp:    10,
			start:  0,
			end:    math.Pi * 2,
			points: 5,
			out:    geom.LineString{{0, 0}, {math.Pi / 2, 10}, {math.Pi, 0}, {3 * math.Pi / 2, -10}, {2 * math.Pi, 0}},
		},
	}

	for k, v := range tcases {
		t.Run(k, fn(v))
	}
}
