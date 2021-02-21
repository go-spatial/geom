package geom_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestPolygonZSetter(t *testing.T) {
	type tcase struct {
		pointzs  [][][3]float64
		lines    [][]geom.LineZ
		setter   geom.PolygonZSetter
		expected geom.PolygonZSetter
		err      error
	}
	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			err := tc.setter.SetLinearRings(tc.pointzs)
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
			if !reflect.DeepEqual(tc.pointzs, glr) {
				t.Errorf("linear rings, expected %v got %v", tc.pointzs, glr)
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
			pointzs: [][][3]float64{
				{
					{10, 20, 30},
					{30, 40, 50},
					{-10, -5, 0},
					{10, 20, 30},
				},
			},
			lines: [][]geom.LineZ{
				{
					{
						{10, 20, 30},
						{30, 40, 50},
					},
					{
						{30, 40, 50},
						{-10, -5, 0},
					},
					{
						{-10, -5, 0},
						{10, 20, 30},
					},
				},
			},
			setter: &geom.PolygonZ{
				{
					{15, 20, 30},
					{35, 40, 50},
					{-15, -5, 0},
					{25, 20, 30},
				},
			},
			expected: &geom.PolygonZ{
				{
					{10, 20, 30},
					{30, 40, 50},
					{-10, -5, 0},
					{10, 20, 30},
				},
			},
		},
		{
			setter: (*geom.PolygonZ)(nil),
			err:    geom.ErrNilPolygonZ,
		},
	}

	for i := range tests {
		t.Run(strconv.FormatInt(int64(i), 10), fn(tests[i]))
	}
}
