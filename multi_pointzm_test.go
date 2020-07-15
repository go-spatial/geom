package geom_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestMultiPointZMSetter(t *testing.T) {
	type tcase struct {
		points   [][4]float64
		setter   geom.MultiPointZMSetter
		expected geom.MultiPointZMSetter
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
			t.Errorf("PointZMs, expected %v got %v", tc.points, pts)
		}
	}
	tests := []tcase{
		{
			points: [][4]float64{
				{15, 20, 30, 40},
				{35, 40, 50, 60},
				{-15, -5, 0, 5},
			},
			setter: &geom.MultiPointZM{
				{10, 20, 30, 40},
				{30, 40, 50, 60},
				{-10, -5, 0, 5},
			},
			expected: &geom.MultiPointZM{
				{15, 20, 30, 40},
				{35, 40, 50, 60},
				{-15, -5, 0, 5},
			},
		},
		{
			setter: (*geom.MultiPointZM)(nil),
			err:    geom.ErrNilMultiPointZM,
		},
	}
	for i, tc := range tests {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}
