package geom_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestPolygonMSetter(t *testing.T) {
	type tcase struct {
		pointms  [][][3]float64
		lines    [][]geom.LineM
		setter   geom.PolygonMSetter
		expected geom.PolygonMSetter
		err      error
	}
	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			err := tc.setter.SetLinearRings(tc.pointms)
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
			if !reflect.DeepEqual(tc.pointms, glr) {
				t.Errorf("linear rings, expected %v got %v", tc.pointms, glr)
				return
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
			pointms: [][][3]float64{
				{
					{10, 20, 3},
					{30, 40, 5},
					{-10, -5, 0.5},
					{10, 20, 3},
				},
			},
			lines: [][]geom.LineM{
				{
					{
						{10, 20, 3},
						{30, 40, 5},
					},
					{
						{30, 40, 5},
						{-10, -5, 0.5},
					},
					{
						{-10, -5, 0.5},
						{10, 20, 3},
					},
				},
			},
			setter: &geom.PolygonM{
				{
					{15, 20, 3},
					{35, 40, 5},
					{-15, -5, 0.5},
					{25, 20, 3},
				},
			},
			expected: &geom.PolygonM{
				{
					{10, 20, 3},
					{30, 40, 5},
					{-10, -5, 0.5},
					{10, 20, 3},
				},
			},
		},
		{
			setter: (*geom.PolygonM)(nil),
			err:    geom.ErrNilPolygonM,
		},
	}

	for i := range tests {
		t.Run(strconv.FormatInt(int64(i), 10), fn(tests[i]))
	}
}
