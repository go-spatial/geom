package geom_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestMultiLineStringZMSetter(t *testing.T) {
	type tcase struct {
		pointzms [][][4]float64
		setter   geom.MultiLineStringZMSetter
		expected geom.MultiLineStringZMSetter
		err      error
	}
	fn := func(t *testing.T, tc tcase) {
		err := tc.setter.SetLineStringZMs(tc.pointzms)
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
		mlzm := tc.setter.LineStringZMs()
		if !reflect.DeepEqual(tc.pointzms, mlzm) {
			t.Errorf("LineStringZMs, expected %v got %v", tc.pointzms, mlzm)
		}
	}
	tests := []tcase{
		{
			pointzms: [][][4]float64{
				{
					{15, 20, 30, 40},
					{35, 40, 50, 60},
				},
				{
					{-15, -5, 0, 5},
					{20, 20, 20, 20},
				},
			},
			setter: &geom.MultiLineStringZM{
				{
					{10, 20, 30, 40},
					{30, 40, 50, 60},
				},
				{
					{-10, -5, 0, 5},
					{15, 20, 20, 20},
				},
			},
			expected: &geom.MultiLineStringZM{
				{
					{15, 20, 30, 40},
					{35, 40, 50, 60},
				},
				{
					{-15, -5, 0, 5},
					{20, 20, 20, 20},
				},
			},
		},
		{
			setter: (*geom.MultiLineStringZM)(nil),
			err:    geom.ErrNilMultiLineStringZM,
		},
	}
	for i, tc := range tests {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}

}
