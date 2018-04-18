package geom_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestMultiPointSetter(t *testing.T) {
	type tcase struct {
		points   [][2]float64
		setter   geom.MultiPointSetter
		expected geom.MultiPointSetter
		err      error
	}
	fn := func(t *testing.T, tc tcase) {
		err := tc.setter.SetPoints(tc.points)
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
		pts := tc.setter.Points()
		if !reflect.DeepEqual(tc.points, pts) {
			t.Errorf("Points, expected %v got %v", tc.points, pts)
		}
	}
	tests := []tcase{
		{
			points: [][2]float64{
				{15, 20},
				{35, 40},
				{-15, -5},
			},
			setter: &geom.MultiPoint{
				{10, 20},
				{30, 40},
				{-10, -5},
			},
			expected: &geom.MultiPoint{
				{15, 20},
				{35, 40},
				{-15, -5},
			},
		},
		{
			setter: (*geom.MultiPoint)(nil),
			err:    geom.ErrNilMultiPoint,
		},
	}
	for i, tc := range tests {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}
