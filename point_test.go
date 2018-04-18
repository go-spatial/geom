package geom_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestPointSetter(t *testing.T) {
	type tcase struct {
		point    [2]float64
		setter   geom.PointSetter
		expected geom.PointSetter
		err      error
	}

	fn := func(t *testing.T, tc tcase) {
		err := tc.setter.SetXY(tc.point)
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
			return
		}
		xy := tc.setter.XY()
		if !reflect.DeepEqual(tc.point, xy) {
			t.Errorf("XY, expected %v, got %v", tc.point, xy)
		}

	}
	testcases := []tcase{
		{
			point:    [2]float64{10, 20},
			setter:   &geom.Point{15, 20},
			expected: &geom.Point{10, 20},
		},
		{
			setter: (*geom.Point)(nil),
			err:    geom.ErrNilPoint,
		},
	}

	for i, tc := range testcases {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}
