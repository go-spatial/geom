package util_test

import (
	"reflect"
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/util"
)

func TestBBox(t *testing.T) {
	testcases := []struct {
		geom     geom.Geometry
		expected util.BoundingBox
	}{

		{
			geom: &geom.Point{X: 1.0, Y: 2.0},
			expected: util.BoundingBox{
				1.0, 2.0,
				1.0, 2.0,
			},
		},
		{
			geom: &geom.LineString{[2]float64{0.0, 0.0},
				[2]float64{6.0, -4.0},
				[2]float64{-6.0, 4.0},
				[2]float64{3.0, 7.0}},
			expected: util.BoundingBox{
				-6.0, -4.0,
				6.0, 7.0,
			},
		},
		{
			geom: &geom.Collection{&geom.Polygon{[][2]float64{[2]float64{0.0, 0.0},
				[2]float64{6.0, -4.0},
				[2]float64{-6.0, 4.0},
				[2]float64{3.0, 7.0}}},
				&geom.Point{X: 1.0, Y: 2.0}},
			expected: util.BoundingBox{
				-6.0, -4.0,
				6.0, 7.0,
			},
		},
	}

	for i, tc := range testcases {
		output := util.BBox(tc.geom)

		if !reflect.DeepEqual(output, tc.expected) {
			t.Errorf("test case (%v) failed. output (%+v) does not match expected (%+v)", i, output, tc.expected)
		}
	}
}
