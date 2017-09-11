package util_test

import (
	"reflect"
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/core"
	"github.com/go-spatial/geom/util"
)

func TestBBox(t *testing.T) {
	testcases := []struct {
		geom     geom.Geometry
		expected util.BoundingBox
	}{
		{
			geom: core.Point{
				X: 1.0,
				Y: 2.0,
			},
			expected: util.BoundingBox{
				1.0, 2.0,
				1.0, 2.0,
			},
		},
		{
			geom: core.LineString{
				{
					X: 0.0,
					Y: 0.0,
				},
				{
					X: 6.0,
					Y: 4.0,
				},
				{
					X: 3,
					Y: 7,
				},
			},
			expected: util.BoundingBox{
				0.0, 0.0,
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
