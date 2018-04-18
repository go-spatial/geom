package geom_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestPolygonSetter(t *testing.T) {
	type tcase struct {
		points   [][][2]float64
		setter   geom.PolygonSetter
		expected geom.PolygonSetter
		err      error
	}
	fn := func(t *testing.T, tc tcase) {
		err := tc.setter.SetLinearRings(tc.points)
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
		if !reflect.DeepEqual(tc.points, glr) {
			t.Errorf("linear rings, expected %v got %v", tc.points, glr)
		}
	}
	tests := []tcase{
		{
			points: [][][2]float64{
				{
					{10, 20},
					{30, 40},
					{-10, -5},
					{10, 20},
				},
			},
			setter: &geom.Polygon{
				{
					{15, 20},
					{35, 40},
					{-15, -5},
					{25, 20},
				},
			},
			expected: &geom.Polygon{
				{
					{10, 20},
					{30, 40},
					{-10, -5},
					{10, 20},
				},
			},
		},
		{
			setter: (*geom.Polygon)(nil),
			err:    geom.ErrNilPolygon,
		},
	}

	for i, tc := range tests {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}
