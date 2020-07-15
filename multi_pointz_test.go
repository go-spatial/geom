package geom_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestMultiPointZSetter(t *testing.T) {
	type tcase struct {
		points   [][3]float64
		setter   geom.MultiPointZSetter
		expected geom.MultiPointZSetter
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
			t.Errorf("PointZs, expected %v got %v", tc.points, pts)
		}
	}
	tests := []tcase{
		{
			points: [][3]float64{
				{15, 20, 30},
				{35, 40, 50},
				{-15, -5, 0},
			},
			setter: &geom.MultiPointZ{
				{10, 20, 30},
				{30, 40, 50},
				{-10, -5, 0},
			},
			expected: &geom.MultiPointZ{
				{15, 20, 30},
				{35, 40, 50},
				{-15, -5, 0},
			},
		},
		{
			setter: (*geom.MultiPointZ)(nil),
			err:    geom.ErrNilMultiPointZ,
		},
	}
	for i, tc := range tests {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}
