package geom_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestPolygonZMSetter(t *testing.T) {
	type tcase struct {
		pointzms [][][4]float64
		lines    [][]geom.LineZM
		setter   geom.PolygonZMSetter
		expected geom.PolygonZMSetter
		err      error
	}
	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			err := tc.setter.SetLinearRings(tc.pointzms)
			if tc.err == nil && err != nil {
				t.Errorf("error, expected nil got %v", err)
				return
			}
			if tc.err != nil {
				if err.Error() != tc.err.Error() {
					t.Errorf("error, expected %v got %v", tc.err, err)
				}
				return
			}

			// compare the results
			if !reflect.DeepEqual(tc.expected, tc.setter) {
				t.Errorf("Polygon Setter, expected %v got %v", tc.expected, tc.setter)
				return
			}

			// compare the results of the Rings
			glr := tc.setter.LinearRings()
			if !reflect.DeepEqual(tc.pointzms, glr) {
				t.Errorf("linear rings, expected %v got %v", tc.pointzms, glr)
			}

			// compare the extracted segments
			segs, err := tc.setter.AsSegments()
			if err != nil {
				if !reflect.DeepEqual(tc.lines, segs) {
					t.Errorf("segments, expected %v got %v", tc.lines, segs)
				}
			}
		}
	}
	tests := []tcase{
		{
			pointzms: [][][4]float64{
				{
					{10, 20, 30, 3},
					{30, 40, 50, 5},
					{-10, -5, 0, 0.5},
					{10, 20, 30, 3},
				},
			},
			lines: [][]geom.LineZM{
				{
					{
						{10, 20, 30, 3},
						{30, 40, 50, 5},
					},
					{
						{30, 40, 50, 5},
						{-10, -5, 0, 0.5},
					},
					{
						{-10, -5, 0, 0.5},
						{10, 20, 30, 3},
					},
				},
			},
			setter: &geom.PolygonZM{
				{
					{15, 20, 30, 3},
					{35, 40, 50, 5},
					{-15, -5, 0, 0.5},
					{25, 20, 30, 3},
				},
			},
			expected: &geom.PolygonZM{
				{
					{10, 20, 30, 3},
					{30, 40, 50, 5},
					{-10, -5, 0, 0.5},
					{10, 20, 30, 3},
				},
			},
		},
		{
			setter: (*geom.PolygonZM)(nil),
			err:    geom.ErrNilPolygonZM,
		},
	}

	for i := range tests {
		t.Run(strconv.FormatInt(int64(i), 10), fn(tests[i]))
	}
}
