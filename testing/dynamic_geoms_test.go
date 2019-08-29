package testing

import (
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
	"github.com/go-spatial/geom/windingorder"
)

func TestBoxPolygon(t *testing.T) {
	type tcase struct {
		dim float64
		res geom.Polygon
	}

	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			got := BoxPolygon(tc.dim)

			wo := windingorder.OfPoints(got[0]...)
			if !wo.IsClockwise() {
				t.Error("winding order of box not clockwise")
			}

			if !cmp.GeometryEqual(got, tc.res) {
				t.Error("geometries not equal")
			}
		}
	}

	tcases := map[string]tcase {
		"dim 1" : {
			dim: 1.0,
			res: geom.Polygon{{{0, 0}, {1.0, 0}, {1.0, 1.0}, {0, 1.0}}},
		},
		"dim -1" : {
			dim: -1.0,
			res: geom.Polygon{{{0, 0}, {-1.0, 0}, {-1.0, -1.0}, {0, -1.0}}},
		},
	}

	for k, v := range tcases {
		t.Run(k, fn(v))
	}
}
