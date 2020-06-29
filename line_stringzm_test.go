package geom_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestLineStringZMSetter(t *testing.T) {
	type tcase struct {
		points   [][4]float64
		setter   geom.LineStringZMSetter
		expected geom.LineStringZMSetter
		err      error
	}
	fn := func(t *testing.T, tc tcase) {
		err := tc.setter.SetVertices(tc.points)
		if tc.err == nil && err != nil {
			t.Errorf("error, expected nil got %v", err)
			return
		}
		if tc.err != nil {
			if tc.err.Error() != err.Error() {
				t.Errorf("error, expected %v got %v", tc.err, err)
			}
			return
		}
		// compare the results
		if !reflect.DeepEqual(tc.expected, tc.setter) {
			t.Errorf("setter, expected %v got %v", tc.expected, tc.setter)
		}
		lszm := tc.setter.Vertices()
		if !reflect.DeepEqual(tc.points, lszm) {
			t.Errorf("Vertices, expected %v got %v", tc.points, lszm)
		}
	}
	tests := []tcase{
		{
			points: [][4]float64{
				{15, 20, 30, 40},
				{35, 40, 30, 40},
				{-15, -5, 12, -3},
			},
			setter: &geom.LineStringZM{
				{10, 20, 30, 40},
				{30, 40, 30, 40},
				{-10, -5, -2, -3},
			},
			expected: &geom.LineStringZM{
				{15, 20, 30, 40},
				{35, 40, 30, 40},
				{-15, -5, 12, -3},
			},
		},
		{
			setter: (*geom.LineStringZM)(nil),
			err:    geom.ErrNilLineStringZM,
		},
	}
	for i, tc := range tests {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}
